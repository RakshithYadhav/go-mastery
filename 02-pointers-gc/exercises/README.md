# Module 2 Exercises

Solve until `go test ./02-pointers-gc/...` passes, then the real gate
`.\check.ps1 ./02-pointers-gc/...`. Hints on request, one nudge at a time. No
solutions.

| # | File | Kind | What it proves |
|---|------|------|----------------|
| 1 | `ex1_format.go` | implement | You can format output with no allocations at all, without `fmt` |
| 2 | `ex2_escape.go` | fix the bug | You can find allocations with `-gcflags=-m` and a benchmark, and tell the two apart when they disagree |
| 3 | `ex3_pool.go` | implement | You can build a `sync.Pool` wrapper that has neither of the two bugs that make pooling dangerous |
| 4 | `ex4_boxing.go` | fix the bug | You know why a disabled log line still costs three allocations, and which of the two standard fixes you would defend |

## Rules of engagement

- **Don't change the tests or the exported signatures.** Unexported fields are
  yours to change freely — in exercise 2 that is where every fix lives.
- **Evidence before edits.** Exercise 2 in particular is not a code-reading
  exercise. Run `go build -gcflags='-m -l' ./02-pointers-gc/exercises` and the
  benchmark first, write down what each tells you, and only then change
  anything. One of its three causes appears in the compiler output and one does
  not; noticing that gap is worth more than the fix.
- **The banned crutch is `fmt` on a hot path.** Every function in that package
  takes `...any` and boxes its arguments before formatting anything, so no call
  to it can reach zero allocations. Exercises 1, 2 and 4 all have a budget test
  that `fmt` cannot pass. `strconv`'s append-style functions are the
  replacement.
- **A budget test is a correctness test on this track.** A right answer that
  allocates is a failing answer, and every budget failure message states the
  cost in events per second so you can see it is not pedantry.
- **Stuck is not failure.** Ask for a hint — you get a nudge, not a solution.

## What to read first

`../NOTES.md` sections 1 to 4 before exercises 1 and 2. Section 7 before
exercise 3. Section 4 again before exercise 4. Run the demo each section points
at; the exercises assume you have seen the behaviour, not just read about it.

Suggested order: **1 → 4 → 2 → 3**.

Exercise 1 is the smallest and teaches the append-style pattern the other three
reuse. Exercise 4 is one decision and a small edit, and it is the one whose
lesson shows up most often in real code review. Exercise 2 is the real work of
this module. Exercise 3 is last because a pool is the thing you reach for only
after everything in exercises 1, 2 and 4 has failed to be enough.

## Definition of done

```powershell
go test ./02-pointers-gc/...                            # fast loop
go build -gcflags='-m -l' ./02-pointers-gc/exercises     # the compiler's reasoning
.\check.ps1 ./02-pointers-gc/...                        # THE GATE
.\check.ps1 -Bench ./02-pointers-gc/...                 # the allocation numbers
```

Green gate, `MISTAKES.md` updated with every real bug, an
`// ORIGINAL (before fix)` block at the bottom of `ex2_escape.go` and
`ex4_boxing.go`, and — for exercise 4 — two sentences in `MISTAKES.md` on which
fix you chose and why.

Then the quiz. The roadmap box does not get ticked until the quiz clears.
