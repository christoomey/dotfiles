#!/usr/bin/env bash
#
# herdr-new-thread's job runner, against a temp HERDR_THREADS_DIR and stub
# herdr / claude-launch binaries on PATH. Nothing here may touch the live herdr
# session: every herdr call goes to the stub, which only records what it was
# asked to do.

set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bin="$(cd "$here/../../../bin" && pwd)"
new_thread="$bin/herdr-new-thread"
job_up="$bin/herdr-job-up"

failures=0
current=""

start() { current="$1"; printf '  %s ... ' "$1"; }
pass() { printf 'ok\n'; }
fail() {
  printf 'FAIL\n    %s\n' "$*"
  failures=$((failures + 1))
}
assert_eq() {
  if [[ "$1" != "$2" ]]; then
    fail "$3: expected [$2], got [$1]"
    return 1
  fi
}
assert_contains() {
  if [[ "$1" != *"$2"* ]]; then
    fail "$3: [$2] not found in [$1]"
    return 1
  fi
}

sandbox() {
  root="$(mktemp -d "${TMPDIR:-/tmp}/herdr-runner-XXXXXX")"
  export HERDR_THREADS_DIR="$root/threads"
  export HERDR_STUB_LOG="$root/herdr-calls.log"
  export CLAUDE_LAUNCH_STUB_LOG="$root/claude-launch-calls.log"
  unset HERDR_STUB_NO_PANE
  jobs="$HERDR_THREADS_DIR/jobs"
  mkdir -p "$jobs" "$root/stubs"
  : >"$HERDR_STUB_LOG"
  : >"$CLAUDE_LAUNCH_STUB_LOG"

  cat >"$root/stubs/herdr" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$HERDR_STUB_LOG"
case "$1 $2" in
  "workspace list")
    cat <<'JSON'
{"result":{"workspaces":[
  {"workspace_id":"ws-dotfiles","label":"dotfiles"},
  {"workspace_id":"ws-moi","label":"moi","worktree":{"checkout_path":"MOI_ROOT","is_linked_worktree":false}}
]}}
JSON
    ;;
  "tab create")
    if [[ -n "${HERDR_STUB_NO_PANE:-}" ]]; then
      echo '{"result":{}}'
    else
      echo '{"result":{"root_pane":{"pane_id":"pane-1"}}}'
    fi
    ;;
  "pane run") exit "${HERDR_STUB_PANE_RUN_RC:-0}" ;;
esac
exit 0
STUB
  sed -i '' "s|MOI_ROOT|$HOME/code/moi|" "$root/stubs/herdr"

  cat >"$root/stubs/claude-launch" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$CLAUDE_LAUNCH_STUB_LOG"
job=""
while [[ $# -gt 0 ]]; do
  [[ "$1" == "--job" ]] && job="$2"
  shift
done
stamp() { date '+%Y-%m-%d %H:%M:%S'; }
if [[ -n "$job" ]]; then
  events="${HERDR_THREADS_DIR}/jobs/$job.events"
  if [[ -n "${CLAUDE_LAUNCH_STUB_RC:-}" && "${CLAUDE_LAUNCH_STUB_RC}" != "0" ]]; then
    printf '%s\tfailed\tclaude-launch: stub failure\n' "$(stamp)" >>"$events"
    exit "$CLAUDE_LAUNCH_STUB_RC"
  fi
  printf '%s\tclaude-started\tstub\n' "$(stamp)" >>"$events"
  printf '%s\tacked\tstub\n' "$(stamp)" >>"$events"
fi
exit 0
STUB
  chmod +x "$root/stubs/herdr" "$root/stubs/claude-launch"
  PATH="$root/stubs:$PATH"
  export PATH

  # ~/.zshenv prepends ~/bin for every zsh, script or not, so exporting PATH is
  # not enough to shadow the real herdr: give the zsh tree its own ZDOTDIR.
  mkdir -p "$root/zdot"
  printf 'export PATH=%q:"$PATH"\n' "$root/stubs" >"$root/zdot/.zshenv"
  export ZDOTDIR="$root/zdot"

  local resolved
  resolved=$(zsh -c 'command -v herdr')
  if [[ "$resolved" != "$root/stubs/herdr" ]]; then
    echo "refusing to run: zsh resolves herdr to ${resolved:-nothing}, not the stub" >&2
    exit 1
  fi
}

teardown() {
  [[ -n "${root:-}" && "$root" == *herdr-runner-* ]] && rm -rf "$root"
}

# mkjob ID MODE PROJECT NAME PROMPT [BRANCH] [SESSION_ID]
mkjob() {
  local id="$1" mode="$2" project="$3" name="$4" prompt="$5" branch="${6:-}" sid="${7:-}"
  mkdir -p "$jobs"
  PROMPT="$prompt" python3 - "$jobs/$id.json" "$id" "$mode" "$project" "$name" "$branch" "$sid" <<'PY'
import json, os, sys
path, jid, mode, project, name, branch, sid = sys.argv[1:8]
json.dump({"id": jid, "created_at": "2026-01-01T00:00:00Z", "mode": mode,
           "project": project, "name": name, "branch": branch,
           "prompt": os.environ["PROMPT"], "attachments": [], "session_id": sid},
          open(path, "w"))
PY
  printf '%s\tqueued\t\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/$id.events"
}

stages() { cut -f2 "$jobs/$1.events" | tr '\n' ' ' | sed 's/ $//'; }
last_detail() { tail -n1 "$jobs/$1.events" | cut -f3; }
age_events() { touch -t "$(date -v"$2" +%Y%m%d%H%M)" "$jobs/$1.events"; }

# ---------------------------------------------------------------- tests

test_claim_is_exclusive() {
  start "claim is exclusive"
  sandbox
  mkjob j-claim open dotfiles "Claim race" "hello"
  "$new_thread" --run-job j-claim >/dev/null 2>&1 &
  "$new_thread" --run-job j-claim >/dev/null 2>&1 &
  wait
  local claimed
  claimed=$(grep -c $'\tclaimed\t' "$jobs/j-claim.events")
  assert_eq "$claimed" "1" "claimed events" || { teardown; return; }
  assert_eq "$(grep -c . "$CLAUDE_LAUNCH_STUB_LOG")" "1" "claude-launch calls" || { teardown; return; }
  pass
  teardown
}

test_chat_events_in_order() {
  start "chat job writes its stages in order"
  sandbox
  mkjob j-chat open dotfiles "Chat job" "first prompt"
  "$new_thread" --run-job j-chat >/dev/null 2>&1
  assert_eq "$(stages j-chat)" \
    "queued claimed prompt-saved workspace tab-created claude-started acked done" \
    "chat stages" || { teardown; return; }
  local call
  call=$(cat "$CLAUDE_LAUNCH_STUB_LOG")
  assert_contains "$call" "--job j-chat" "claude-launch --job" || { teardown; return; }
  assert_contains "$call" "--prompt-file" "claude-launch --prompt-file" || { teardown; return; }
  assert_contains "$(cat "$HERDR_THREADS_DIR"/chat-prompts/*.md)" "first prompt" "saved prompt" || { teardown; return; }
  assert_contains "$(cat "$HERDR_STUB_LOG")" "notification show Starting Chat job" "start notification" || { teardown; return; }
  pass
  teardown
}

test_resume_job_maps_to_bg() {
  start "resume-bg job resumes in the background"
  sandbox
  mkjob j-res resume-bg dotfiles "Resume job" "" "" "sess-123"
  "$new_thread" --run-job j-res >/dev/null 2>&1
  assert_eq "$(stages j-res)" "queued claimed workspace tab-created claude-started acked done" \
    "resume stages" || { teardown; return; }
  assert_contains "$(cat "$CLAUDE_LAUNCH_STUB_LOG")" "--resume sess-123" "resume arg" || { teardown; return; }
  assert_contains "$(cat "$HERDR_STUB_LOG")" "--no-focus" "background tab" || { teardown; return; }
  pass
  teardown
}

test_legacy_launch_without_a_job() {
  start "--launch still resumes without a job (herdr-open-chat)"
  sandbox
  printf '{"project":"dotfiles"}' | "$new_thread" --launch "Legacy" open sess-9 >/dev/null 2>&1
  local rc=$?
  assert_eq "$rc" "0" "exit status" || { teardown; return; }
  assert_contains "$(cat "$CLAUDE_LAUNCH_STUB_LOG")" "--resume sess-9" "resume arg" || { teardown; return; }
  if compgen -G "$jobs/*.events" >/dev/null; then
    fail "a jobless launch wrote job events"
    teardown
    return
  fi
  pass
  teardown
}

test_feature_events_and_pane_command() {
  start "feature job hands the bootstrap to herdr-job-up"
  sandbox
  mkjob j-feat feature moi "Feature job" "do the thing" "ct/abc-123-thing"
  "$new_thread" --run-job j-feat >/dev/null 2>&1
  assert_eq "$(stages j-feat)" "queued claimed prompt-saved workspace tab-created" \
    "feature stages" || { teardown; return; }
  local run_line
  run_line=$(grep '^pane run' "$HERDR_STUB_LOG")
  assert_contains "$run_line" "herdr-job-up j-feat --" "job-up in pane command" || { teardown; return; }
  assert_contains "$run_line" "bin/up" "bootstrap command" || { teardown; return; }
  assert_contains "$run_line" "&& exit" "tab closes on success" || { teardown; return; }
  assert_contains "$run_line" "--prompt-file" "prompt file passed through" || { teardown; return; }
  pass
  teardown
}

test_job_up_success() {
  start "herdr-job-up records a clean bootstrap"
  sandbox
  mkjob j-up feature moi "Up job" "p" "ct/x"
  "$job_up" j-up -- /bin/echo bootstrapped >/dev/null 2>&1
  local rc=$?
  assert_eq "$rc" "0" "job-up exit status" || { teardown; return; }
  assert_eq "$(stages j-up)" "queued bootstrapping up-exited done" "job-up stages" || { teardown; return; }
  assert_contains "$(cat "$jobs/j-up.up.log")" "bootstrapped" "up log" || { teardown; return; }
  pass
  teardown
}

test_job_up_failure_notifies() {
  start "herdr-job-up reports a failed bootstrap"
  sandbox
  mkjob j-upfail feature moi "Up fail" "p" "ct/x"
  "$job_up" j-upfail -- /bin/sh -c 'echo boom >&2; exit 3' >/dev/null 2>&1
  local rc=$?
  assert_eq "$rc" "3" "job-up exit status" || { teardown; return; }
  assert_eq "$(stages j-upfail)" "queued bootstrapping up-exited failed" "job-up stages" || { teardown; return; }
  assert_eq "$(last_detail j-upfail)" "bootstrapping: up exited 3" "failure detail" || { teardown; return; }
  assert_contains "$(cat "$HERDR_STUB_LOG")" "notification show Feature bootstrap failed" "notification" || { teardown; return; }
  pass
  teardown
}

test_failing_stage_fails_the_job() {
  start "a dead stage fails the job and notifies"
  sandbox
  mkjob j-dead open dotfiles "Dead job" "keep me"
  HERDR_STUB_NO_PANE=1 "$new_thread" --run-job j-dead >/dev/null 2>"$jobs/j-dead.err"
  local rc=$?
  [[ $rc -eq 0 ]] && { fail "runner exited 0 on a failed stage"; teardown; return; }
  assert_eq "$(stages j-dead)" "queued claimed prompt-saved workspace failed" "stages" || { teardown; return; }
  assert_contains "$(last_detail j-dead)" "workspace: tab create returned no pane" "failure detail" || { teardown; return; }
  assert_contains "$(cat "$HERDR_STUB_LOG")" "notification show New session failed" "notification" || { teardown; return; }
  pass
  teardown
}

test_retry() {
  start "retry refuses done jobs and re-runs the rest"
  sandbox
  mkjob j-done open dotfiles "Done job" ""
  printf '%s\tdone\t\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/j-done.events"
  "$new_thread" --retry j-done >/dev/null 2>&1
  [[ $? -eq 0 ]] && { fail "retry accepted a done job"; teardown; return; }
  assert_contains "$(stages j-done)" "done" "done job untouched" || { teardown; return; }

  mkjob j-again open dotfiles "Retry job" "again"
  mkdir -p "$jobs/j-again.claim"
  printf '%s\tfailed\tworkspace: nope\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/j-again.events"
  "$new_thread" --retry j-again >/dev/null 2>&1
  local i
  for i in $(seq 1 40); do
    [[ "$(tail -n1 "$jobs/j-again.events" | cut -f2)" == "done" ]] && break
    sleep 0.25
  done
  assert_eq "$(stages j-again)" \
    "queued failed retry claimed prompt-saved workspace tab-created claude-started acked done" \
    "retry stages" || { teardown; return; }
  pass
  teardown
}

test_sweep_stranded_and_stalled() {
  start "sweep notices stranded and stalled jobs, once each"
  sandbox
  mkjob j-strand open dotfiles "Stranded" "p"
  age_events j-strand -5M
  mkjob j-stall open dotfiles "Stalled" "p"
  mkdir -p "$jobs/j-stall.claim"
  printf '%s\ttab-created\tpane-1\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/j-stall.events"
  age_events j-stall -10M
  mkjob j-slow feature moi "Slow feature" "p" "ct/x"
  mkdir -p "$jobs/j-slow.claim"
  printf '%s\tbootstrapping\taf\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/j-slow.events"
  age_events j-slow -10M

  "$new_thread" --sweep >/dev/null 2>&1
  local notices
  notices=$(grep -c 'notification show' "$HERDR_STUB_LOG")
  assert_eq "$notices" "2" "notifications (stranded + stalled)" || { teardown; return; }
  assert_contains "$(cat "$HERDR_STUB_LOG")" "Launch never started --body Stranded" "stranded notice" || { teardown; return; }
  assert_contains "$(cat "$HERDR_STUB_LOG")" "Launch stalled --body Stalled" "stalled notice" || { teardown; return; }

  "$new_thread" --sweep >/dev/null 2>&1
  assert_eq "$(grep -c 'notification show' "$HERDR_STUB_LOG")" "2" "second sweep stays quiet" || { teardown; return; }
  pass
  teardown
}

test_sweep_retention() {
  start "sweep keeps 14 days of terminal jobs"
  sandbox
  mkjob j-old open dotfiles "Old" "p"
  printf '%s\tdone\t\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/j-old.events"
  age_events j-old -15d
  mkjob j-recent open dotfiles "Recent" "p"
  printf '%s\tdone\t\n' "$(date '+%Y-%m-%d %H:%M:%S')" >>"$jobs/j-recent.events"
  age_events j-recent -13d
  mkjob j-openold open dotfiles "Never finished" "p"
  age_events j-openold -20d

  "$new_thread" --sweep >/dev/null 2>&1
  [[ -e "$jobs/j-old.json" || -e "$jobs/j-old.events" ]] && { fail "old terminal job was kept"; teardown; return; }
  [[ -e "$jobs/j-recent.events" ]] || { fail "13-day-old job was deleted"; teardown; return; }
  [[ -e "$jobs/j-openold.events" ]] || { fail "non-terminal job was deleted"; teardown; return; }
  pass
  teardown
}

test_jobs_listing() {
  start "--jobs and --job report what happened"
  sandbox
  mkjob j-list open dotfiles "Listed job" "a prompt"
  "$new_thread" --run-job j-list >/dev/null 2>&1
  local table
  table=$("$new_thread" --jobs 5)
  assert_contains "$table" "j-list" "job id in table" || { teardown; return; }
  assert_contains "$table" "Listed job" "job name in table" || { teardown; return; }
  assert_contains "$table" "done" "stage in table" || { teardown; return; }

  local json
  json=$("$new_thread" --jobs --json)
  if ! printf '%s' "$json" | python3 -c '
import json, sys
rows = json.load(sys.stdin)
row = [r for r in rows if r["id"] == "j-list"][0]
assert row["stage"] == "done", row
assert row["prompt_chars"] == len("a prompt"), row
assert row["duration_seconds"] is not None, row
'; then
    fail "--jobs --json did not describe the job"
    teardown
    return
  fi

  local detail
  detail=$("$new_thread" --job j-list)
  assert_contains "$detail" "a prompt" "prompt in --job" || { teardown; return; }
  assert_contains "$detail" "j-list.up.log" "log paths in --job" || { teardown; return; }
  pass
  teardown
}

for t in \
  test_claim_is_exclusive \
  test_chat_events_in_order \
  test_resume_job_maps_to_bg \
  test_legacy_launch_without_a_job \
  test_feature_events_and_pane_command \
  test_job_up_success \
  test_job_up_failure_notifies \
  test_failing_stage_fails_the_job \
  test_retry \
  test_sweep_stranded_and_stalled \
  test_sweep_retention \
  test_jobs_listing; do
  "$t"
done

if [[ $failures -gt 0 ]]; then
  printf '%d runner test(s) failed\n' "$failures"
  exit 1
fi
printf 'runner tests passed\n'
