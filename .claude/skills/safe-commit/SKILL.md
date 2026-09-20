---
name: safe-commit
description: Commit the working tree of this PUBLIC dotfiles repo as one commit per logical change, after screening every added line for personal, company or secret material. Use when the user asks to commit, "safe commit", or split changes into commits here.
---

This repo is public on GitHub. Everything committed, including the commit
message, is published for good. The goal is dev tooling only: nothing that
identifies people, exposes the employer's internals, or leaks a secret.

Only run when the user asked for commits. Never push, never amend, never
rewrite history.

## 1. Survey

- `git status --short`, `git diff`, and read every untracked file in full.
- Group the changes into logical commits (one idea each). A file with hunks
  for two ideas is split with a filtered patch and `git apply --cached`. When
  two ideas share hunks and a split would leave a commit that does not build,
  make one commit whose subject names both, and say so in the report.

## 2. Screen the added lines

Read the diff yourself; grep only backs that up. Check added lines, new files,
file names, test fixtures, demo data, comments, and the commit messages you are
about to write.

Block (do not commit, stop and tell the user):
- Secrets: tokens, API keys, passwords, private keys, session cookies, signed
  URLs, `.env` contents.
- Personal data: phone numbers, home address, personal email other than the
  public git identity, anything about family or health.
- Other people: colleague names, handles, emails, Slack or Linear user ids.
- Customer or patient data of any kind, real or realistic-looking.

Ask (show the line, propose a neutral rewrite, wait for the answer):
- Employer internals that are new to the repo: product behaviour, bug
  descriptions, domain terms, architecture, table or service names, issue
  titles, branch names that describe real work, internal hostnames, AWS account
  ids, customer names.
- Text copied from Linear, Slack, PRs or internal docs.
- Absolute paths or hostnames that reveal more than `~/code/...`.

Fine (mention only if unusual):
- Names already public in this repo. Check with `git grep -i <term>` against
  HEAD: if it is already there in the same sense, it is baseline, not a leak.
- Links to dev tooling consoles (CloudWatch, Sentry, GitHub) with no
  credentials in them.
- Generic tool config, keybindings, scripts.

Fixtures and examples are the usual leak: prefer obviously made-up text
("fix login bug", `user/team-123-short-description`) over anything that came
from real work.

Mechanical backstop over the added lines:

```sh
git diff HEAD | grep '^+' | grep -nEi \
  '(api[_-]?key|secret|token|passw|bearer|BEGIN [A-Z ]*PRIVATE|AKIA[0-9A-Z]{16}|xox[bap]-|gh[pousr]_[A-Za-z0-9]{20,}|[0-9]{12}\.dkr|arn:aws|[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[a-z]{2,}|https?://)'
```

Run the same over untracked files. Every hit gets a verdict; a hit is not
automatically a problem and a clean grep is not a pass.

## 3. Commit

- Anything blocked or unanswered stays out. Commit the clean groups and leave
  the rest in the working tree.
- Stage explicit paths per group (`git add <paths>`), never `git add -A`.
- Message: one line, imperative, in the style of `git log --oneline -10`. No
  body. Follow it with the session's attribution trailer.
- Run the relevant checks first when the group has them (for example
  `go vet ./... && go test ./...` in `herdr/new-session-form`).

## 4. Report

Tersely: the commits made (hash + subject), each screening finding with its
verdict, and anything left uncommitted and why.
