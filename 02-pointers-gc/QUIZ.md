# Module 2 — Quiz Tracker

Oral quiz over the questions in `NOTES.md` §8, plus the per-section self-checks.
Graded honestly; the roadmap box does not get ticked until every question
clears.

Because this module's notes are a reading list, the quiz is the only check that
the reading actually happened. Answer without the articles open.

## Round 1 — not started

**Q1 (the decision).** What decides whether a Go value lives on the stack or
the heap? Who makes that decision, and when?

**Q2 (the tool).** Walk me through what `go build -gcflags=-m` prints and how
you use it. Why does `-l` matter?

**Q3 (the trap).** The compiler says `make([]byte, n) does not escape` and the
benchmark says one allocation per call. Which is wrong? Explain.

**Q4 (boxing).** Why does putting an `int` into an `interface{}` allocate?

## Round 2 — not started

**Q5 (the collector).** Describe Go's garbage collector: algorithm, what runs
concurrently with your program, and what it costs you.

**Q6 (the knob).** What does `GOGC` control? What would `GOGC=off` do to a
service that allocates steadily, and why is it not simply faster?

**Q7 (diagnosis).** A service has stable memory and rising CPU, and the profile
blames the runtime. How do you diagnose it, and which pprof profile and flag do
you reach for?

**Q8 (stacks).** Goroutine stacks start small and grow. How does that work, and
what does the runtime have to fix up when it happens? Why can Go do this when C
cannot?

## Round 3 — not started

**Q9 (pooling).** When is `sync.Pool` the right answer, and when is it a
mistake? Name both bugs that make a pool worse than no pool.

**Q10 (semantics).** How do you choose between a value receiver and a pointer
receiver? Answer without mentioning the size of the struct.

## Open follow-ups (answer these to close a round)

*(none yet)*

---

*Answers are graded live in chat; this file is the persistent summary so the
quiz resumes across sessions without re-asking cleared questions. Record the
content of the answer, not just the verdict.*

*Quiz backlog across this track: Module 1 (both rounds), Module 2 (all three).
Track 1's backlog is parked, not cleared — Module 3 Q6–Q8 and all of Modules 4,
5, 6, 7 in `../../go-concurrency/`.*
