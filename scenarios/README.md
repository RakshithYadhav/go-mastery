# Track 2 War Stories — design doc

Three ticket-shaped production problems, each with a measured before and after.
One resume bullet per closed scenario, landing in
[`../../go-stories/WAR-STORIES.md`](../../go-stories/WAR-STORIES.md) with the
full four-part package. The rules, the bullet formula and the voice all live in
that file — this document only covers what is specific to Track 2.

## Why these three

**All three are manufacturing.** Track 1 split its six scenarios between your
real domain and standalone portfolio pieces. Track 2 does not split: every
story here belongs to one fictional manufacturing-SaaS platform, the same shape
as the ERP-import and Gantt-conflict stories from Track 1. That is a deliberate
call. Three stories in one domain read as *one engineer who owns a system*,
which is what you actually want an interviewer to conclude. Six stories across
six domains read as a portfolio.

**None of them are concurrency stories.** Track 1 owns that material and you
already have two closed bullets there. These three are about the other half of
production Go: what your code allocates, what it keeps alive, and what happens
when a value is valid and wrong at the same time. Together they let you answer
"tell me about a performance problem you fixed" and "tell me about a bug that
was hard to find" without repeating yourself.

**The diagnostic tools are spread on purpose.** Every interviewer's real
question is *"how did you find it?"* Scenario 1 is found with a benchmark,
scenario 2 with a test and a print, scenario 3 with a heap profile. Three
different answers.

## Scale honesty

Each scenario runs a scaled-down synthetic version of its problem — seconds,
not a shift. `PROBLEM.md` states the real-world scale it models; `RESULT.md`
records only what you actually measured on your machine. Never quote the
modelled scale as if you measured it. The numbers in the resume bullet are the
measured ones.

## The scenarios

| # | Dir | Story | Concept under test | Keywords for the bullet |
|---|-----|-------|--------------------|-------------------------|
| 1 | `01-telemetry-alloc` | The shop-floor telemetry collector burns 60% of its CPU in garbage collection at peak; dashboards lag minutes behind the machines they show | allocation on the hot path — `[]byte`→`string` on map writes, un-sized slices, `fmt` in a loop (Modules 1 §5, 2) | Go, performance profiling, memory optimisation |
| 2 | `02-erp-nil-interface` | Overnight ERP sync reports ~8% of work orders as failed and retries them; the retries succeed, so the plant ends up with duplicate work orders on the floor | the typed nil inside a non-nil `error` — an interface's (type, value) pair (Modules 4, 5) | Go, error handling, data integrity |
| 3 | `03-label-retain` | The label-printing service's memory climbs steadily through every shift and the pod is OOM-killed most nights around 3am | a small slice retaining a large backing array (Module 1 §4) | Go, memory leak diagnosis, heap profiling |

### Unlock order

Each scenario needs the module that teaches its concept:

- **Scenario 1** unlocks after **Module 2** (it needs escape analysis and the
  allocation budget vocabulary, not just Module 1's map rules).
- **Scenario 2** unlocks after **Module 5**.
- **Scenario 3** unlocks after **Module 1**, but runs **last** anyway.

Suggested order **1 → 2 → 3**, and the reason is not difficulty of the fix —
all three fixes are small. It is difficulty of the *diagnosis*. Scenario 1's
evidence is handed to you by a benchmark you run on purpose. Scenario 2's
evidence is a test you have to think to write. Scenario 3 gives you a graph
that goes up and nothing else; you have to reach for `pprof` and read a heap
profile to find a line of code that looks completely innocent. That is the
hardest interview story to tell well, so it goes last, the same way Track 1
put the goroutine-leak hunt sixth.

## How one scenario runs

Unchanged from Track 1 — do not skip step 2:

1. Claude builds the broken service, the fake external system, and a
   measurement harness. `PROBLEM.md` reads like a ticket from a tech lead.
2. **You run the harness first and record the baseline.** That number is the
   "before" in every sentence you will ever say about this work.
3. You diagnose and fix it. Hints only. Scenarios 1 and 3 need a profiler, not
   just code reading.
4. You re-run the harness. The tests assert the target, so "done" is a number.
5. You write `RESULT.md`, then the four-part package in `WAR-STORIES.md`.
   Claude helps write the package — after the fix is measured, never before.

## Definition of done, per scenario

- [ ] Baseline recorded in `RESULT.md` before any code was changed
- [ ] The acceptance test passes, and it asserts a number, not a behaviour
- [ ] The DO-NOT-MODIFY file is unmodified (`git diff` proves it)
- [ ] `.\check.ps1` clean on the whole scenario
- [ ] `RESULT.md` complete: baseline → what changed and *why this shape* → after
- [ ] Four-part package added to `../../go-stories/WAR-STORIES.md`
- [ ] The scenario directory moved into `go-stories/go-mastery/`, per roadmap
      rule 9

## Status

Nothing scaffolded yet. Scenarios get built one at a time, when the module that
unlocks them closes. Scenario 1 is the first, after Module 2.
