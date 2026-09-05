# Module 1 Exercises

Same rules as Track 1: solve until `go test ./01-memory/...` passes, then the
real gate `.\check.ps1 ./01-memory/...`. Hints on request, one nudge at a time.
No solutions.

| # | File | Kind | What it proves |
|---|------|------|----------------|
| 1 | `ex1_ring.go` | implement | You can build a hot-path data structure that allocates exactly once, at construction |
| 2 | `ex2_shifts.go` | fix the bug | You can spot the shared-backing-array corruption that no tool reports, and fix it without copying |
| 3 | `ex3_compact.go` | implement | You can filter in place and still release what you dropped |
| 4 | `ex4_tagcount.go` | implement | You can tell a map read from a map write, and keep 40,000 events a second off the allocator |

## Rules of engagement

- **Don't change the tests or the signatures.** If a test looks wrong, say so
  and argue it — that's a legitimate move, and it has been right before.
- **The allocation budget is part of correctness on this track.** Four of these
  tests assert `testing.AllocsPerRun(...) == 0`. A correct answer that
  allocates is a failing answer. That is not pedantry: every one of these
  budgets is stated in the exercise's own scenario in events per second.
- **The banned crutch is `copy` as a reflex.** Copying the input makes ex2 and
  ex3 pass their correctness tests and fail their budgets. Copying is
  sometimes exactly right — Section 4 of the notes tells you to — but it has
  to be a decision you can defend, not the first thing you reach for. The
  budgets are rigged so it cannot pass silently.
- **Trace before you run.** ex2 in particular is a paper exercise: four work
  orders, two slices, write down eight numbers. Running it first tells you
  *that* it breaks, which you already know from the ticket.
- **Stuck is not failure.** Ask for a hint — you get a nudge, not a solution.

## What to do first

Read `../NOTES.md` sections 1–3 before ex1 and ex2. Sections 4–5 before ex3 and
ex4. Run the demo each section points at; the demos exist so that the exercises
are about *applying* the idea, not discovering it.

Suggested order: **2 → 1 → 3 → 4**. Exercise 2 is the shortest and it is the
one that makes the rest make sense — everything else in this module is a
consequence of the thing ex2 shows you.

## Definition of done

```powershell
go test ./01-memory/...                 # fast loop, run this constantly
.\check.ps1 ./01-memory/...             # THE GATE: build + vet + staticcheck + test
.\check.ps1 -Bench ./01-memory/...      # optional: see the allocation numbers
```

Green gate, `MISTAKES.md` updated with every real bug (including the ones you
fixed in thirty seconds — repeats are the most useful entries), and an
`// ORIGINAL (before fix)` block added at the bottom of `ex2_shifts.go` once
you have closed it.

Then the quiz. The roadmap box does not get ticked until the quiz clears.
