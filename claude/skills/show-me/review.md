# Review walkthrough

A page Chris reads after an agent worked without Chris (a /rinse, a long autonomous session, a
fresh-perspective triage) to rebuild his understanding and decide, item by item, what to keep.
He reviews in a fixed order: what's on me -> each change, one at a time -> new or
pre-existing, and can it really happen? -> how it happens -> what the code does now ->
keep, drop, or ticket. The page follows that order.

## Gather (ground everything)

- The changes: `git log --oneline origin/<trunk>..HEAD` and per-commit diffs. A /rinse run also
  has its ledger (see the `rinse` section).
- Code on the page comes from `git show` / `git diff` / `sed -n`, never retyped. Each excerpt
  links to the line: the PR's files view when a PR exists, else the blob on GitHub at the
  commit (`https://github.com/<owner>/<repo>/blob/<sha>/<path>#L<n>`).
- Repros and before/after output come from what was actually run (test output, curl,
  screenshots). Where nothing was captured, label the item "not reproduced" instead of
  describing a run that didn't happen.
- If one change is really two independent ones, make two items.

## Page

Publish as a private Artifact (load `artifact-design` first). Title: the feature's name plus
"review" in a few words. Source file: `<scratch>/show-me/review.html`; republish the same path
when something changes (a drop, a follow-up), so the link stays the same.

1. **Header**: the original ask in a sentence, branch, PR link (or "no PR"), and a one-line
   bottom line ("3 fixes change behavior, 4 are tests only; 1 needs your call").
2. **Needs you**: decisions and questions, each linking to its item. Then **FYI**. Stable ids
   with anchors (`#R4`), so "drop R4" and links both work.
3. **Glossary**: only when the change uses two or more terms that are easy to mix up. Domain
   words, one line each.
4. **Items, riskiest first** (behavior changes before test-only, Bugs before Hardening; never
   file order). Each is a `<details>`, the first one open:
   - Summary line: id, one-sentence headline, and badges: class, new in this branch vs
     pre-existing, reachable in real use vs only with injected data, verified by (ran / read
     / inferred).
   - **What goes wrong**: one sentence, then the repro as a numbered sequence with the state
     at each step (rows, versions, what server and device each believe).
   - **Before / after**: a mermaid sequence pair when ordering or a race is involved; a
     worked-example table for boundaries; otherwise just the pseudocode diff.
   - **The change**: a pseudocode diff (5-15 lines) plus the real excerpt with its line link.
   - **Edge cases**: case | before | after | test that pins it | handled / deliberately not /
     unknown.
   - **Why tests missed it**, in one line, when it's a Bug.
   - **Agent's call**, boxed apart from the evidence: keep/drop recommendation, confidence,
     alternatives not taken.
   - **To drop**: the commit sha and "ask me to drop R4".
5. **Read the code yourself**: up to 5 places most likely to draw review comments, with line
   links and a one-line reason each.
6. **Not done**: tests not run, paths not verified, follow-ups deliberately left out.
7. **Evidence**: tests run with counts; screenshots or videos in `<details>`, a plain before
   video and a settled after screenshot, no on-screen annotations.

Length: the headline, badges and "What goes wrong" carry each item; everything else is
collapsed or short. Never customer or production data.

## Chat

Reply with the link and the "Needs you" list, nothing more. Then take items one at a time as
Chris asks. When Chris drops or changes something, update the page and republish to the same
path.

## `rinse`

Called at the end of /rinse with its ledger. On top of the above:

- Header bottom line: passes run, why it stopped (clean, no progress, handed back, cap), and
  counts by class.
- Every ledger row gets an item, including dismissed rows (collapsed, under FYI, with the
  dismissal evidence) so Chris can overrule a dismissal.
- Each item names the pass that found it, and whether a later pass re-raised it or a
  fix was refuted and reverted.
- Hardening items say plainly how unlikely they are and what injected data reaches them:
  these are the ones Chris most often drops.
- Anything rinse decided that touches a choice Chris made goes in Needs you, even when it was
  fixed.
- "Not done" includes the final pass's fixes if the refute check didn't run.
