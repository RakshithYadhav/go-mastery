# Go Language Mastery

> **Track 2 of the Backend Mastery roadmap** — see [`../BACKEND-MASTERY.md`](../BACKEND-MASTERY.md)
> for the full 11-track plan this belongs to.

How Go actually works underneath the code you already write: what a slice
really is, where your data lives, when it gets copied, why an interface can be
non-nil and nil at the same time, and how to prove any of it with the tools
professionals use. Track 1 taught you to make Go concurrent. This track teaches
you to make it *correct and cheap* — the half of the language that shows up in
production incidents and in every senior Go interview.

## Depth, set deliberately

Everything cannot be deep. Two areas get **production depth** — you should be
able to defend them under pressure and diagnose them in a live system:

- **Memory & layout** — slices, maps, strings, pointers, stack vs heap,
  escape analysis, GC (Modules 1–2)
- **Interfaces & errors** — implicit satisfaction, the nil-interface trap,
  wrapping, `errors.Is`/`As`, `defer`/`panic`/`recover` (Modules 4–5)

The rest is **working depth**: enough to use correctly and to recognise when
something is wrong, not enough to give a conference talk about.

## How each module works

1. **Explain** — read `NOTES.md`, run the programs in `demo/` and observe.
2. **Exercise** — implement the skeletons in `exercises/` until `check.ps1`
   passes. No peeking at solutions; ask for hints.
3. **Review** — code gets reviewed like a PR, then a short interview-style quiz.

```powershell
go run ./01-memory/demo/01-header      # run a demo
go test ./01-memory/...                # fast iteration loop

.\check.ps1 ./01-memory/...            # THE GATE: build + vet + staticcheck + test
.\check.ps1 -Bench ./01-memory/...     # the gate plus benchmarks with alloc counts
go test -run TestRing ./01-memory/exercises   # a single test
```

> **`go test` passing is not done on this track.** The bugs this track teaches
> — a sub-slice quietly sharing a backing array, a function allocating on every
> call in a hot loop, a typed nil hiding inside a non-nil `error` — all compile
> clean and pass a happy-path test. `check.ps1` adds the two things that
> actually catch them: the static analysers, and tests that assert an
> **allocation budget** (`testing.AllocsPerRun`) rather than just an answer.
>
> Some exercises also carry a `-race` gate where they touch shared memory
> across goroutines. Those say so, and `test-race.ps1` runs them in Docker,
> same as Track 1.

## The gate, named honestly

Every module on this track has a deterministic oracle, so "done" is never a
feeling. Two forms:

| Module type | Oracle | What the test asserts |
|---|---|---|
| Most modules | `go test` + `vet` + `staticcheck` | behaviour, plus the analysers finding nothing |
| Memory modules (1, 2) | the above **plus an allocation budget** | `testing.AllocsPerRun(...) == 0` (or a stated number) — a correct-but-wasteful answer fails |

## Roadmap

- [ ] **Module 1 — Memory & layout: slices, maps, strings**: the slice header,
      `append` growth and when it copies, the aliasing trap, full slice
      expressions, map internals and why `&m[k]` is illegal, iteration-order
      randomisation, the retained-backing-array leak, `string`↔`[]byte` cost
      *(production depth)*
- [ ] **Module 2 — Pointers, the stack, and the garbage collector**: pointers
      vs values, stack vs heap, escape analysis (`-gcflags=-m`), what the GC
      actually does, allocation budgets, `sync.Pool` honestly *(production depth)*
- [ ] **Module 3 — Structs, methods, embedding**: memory layout and field
      alignment, value vs pointer receivers, method sets, embedding vs
      inheritance, the copied-mutex trap
- [ ] **Module 4 — Interfaces deep**: implicit satisfaction, the (type, value)
      pair, the nil-interface trap, type assertions and switches, `any`,
      interface dispatch cost, accept-interfaces-return-structs *(production depth)*
- [ ] **Module 5 — Errors, `defer`, `panic`, `recover`**: wrapping with `%w`,
      `errors.Is`/`errors.As`, sentinel vs typed errors, exact `defer` and
      `recover` rules, named-return mutation, when to panic *(production depth)*
- [ ] **Module 6 — Generics and the `io` model**: type parameters, constraints,
      when NOT to use them, `Reader`/`Writer` composition, buffering, streaming
      without loading the file
- [ ] **Module 7 — Encoding, text, and time**: `encoding/json` tags, custom
      marshalers, streaming decode, `string` vs `rune` vs `byte` and real UTF-8,
      `time` discipline — UTC only, monotonic vs wall clock
- [ ] **Module 8 — Tooling and the professional loop**: modules, versioning and
      package boundaries, testing deep (table-driven, subtests, golden files,
      fuzzing, benchmarks, fakes vs mocks), `golangci-lint`, first contact with
      `pprof`, `delve`, build tags, cross-compilation, Git mastery (rebase,
      bisect, reflog, cherry-pick)
- [ ] **Scenarios — war stories**: three ticket-style production problems from
      the manufacturing domain, each with a measured before/after — telemetry
      parser allocation storm, an ERP adapter's typed-nil error marking good
      syncs as failed, and a label service pinning whole CSVs in memory. Design
      doc: [`scenarios/README.md`](scenarios/README.md); bullets land in
      [`../go-stories/WAR-STORIES.md`](../go-stories/WAR-STORIES.md)
- [ ] **Capstone — streaming data processor**: a production-quality Go CLI or
      library with full tests, benchmarks, docs, and a stated allocation
      budget. Spec written when Module 8 closes, then handed to the `mentor`
      skill.

A box gets ticked when the module's **quiz** clears, not when its exercises
compile.

## Progress log

| Date | What happened |
|------|---------------|
| 2026-09-04 | Repo scaffolded. Gate tooling installed and verified: `staticcheck` 2026.2.1, `golangci-lint` 2.13.2, `dlv` 1.27.1, on Go 1.26.4. `check.ps1` written — build → vet → staticcheck → test, stopping at the first failure; `test-race.ps1` carried over from Track 1 for the modules that touch shared memory. Depth set: production on Modules 1–2 and 4–5, working elsewhere. |
| 2026-09-04 | Module 1 (Memory & layout) opened: NOTES (the slice header and the shelf-and-bookmark analogy, a ptr/len/cap walkthrough table, `append`'s two-line rule with a step table, the split-and-append trap and the full slice expression, the retained-backing-array leak, map internals — no stable addresses, randomised order, `delete` never shrinking — and the read-vs-write string-key allocation rule), 5 demos (03 and 05 are failure-mode demos), 4 exercises (ring buffer, the shift-splitter planted bug, in-place compaction that nils its tail, zero-allocation tag counter), MISTAKES.md pre-seeded with Track 1 carry-forwards and 7 predicted traps, QUIZ.md with 8 questions in two rounds. All demos verified running; all exercises verified failing for the intended reason only; intended fixes privately verified against every target including the four allocation budgets, then discarded. |
| 2026-09-05 | **Module 2 (Pointers, the stack, and the GC) opened, with NOTES as a curated reading list rather than written notes** — his call: blogs, official documentation and the Ardan Labs series explain these fundamentals better than a paraphrase would. NOTES is now seven sections of links (Tour of Go, Alex Edwards, Effective Go, the Go FAQ, Ardan Labs' four-part language-mechanics series, the official GC guide, Rick Hudson's ISMM keynote, Dave Cheney, Cloudflare, Julia Evans, VictoriaMetrics, and Vincent Blanchon's *A Journey With Go*), each with why that source, what to extract, the demo to run against it, and self-check questions. Every link verified on 2026-09-05. Demos and exercises are original as usual. |
| 2026-09-05 | Module 2 built and verified: 5 demos and 4 exercises. Demos — escape analysis measured beside `-gcflags='-m -l'`; a goroutine stack observed moving under a live pointer (five stack growths, address rewritten each time); GC cost across GOGC=100/400/off and GOMEMLIMIT; five deceptive allocation traps; `sync.Pool` measured, emptied by a GC, and shown with both of its bugs. Exercises — zero-allocation formatting, a three-cause escape hunt, a safe buffer pool, and the disabled-log-line boxing bug. All demos run; all exercises fail for the intended reason only; intended fixes privately verified against every target then discarded. |
| 2026-09-05 | Three scaffold bugs caught by verification, all where the measurement contradicted the text. (1) Demo 01 rows for interface boxing and variable-sized `make` both measured 0, because a constant boxed into an interface comes from read-only data and a literal argument lets the compiler prove the bound — fixed with package-level variables. (2) Demo 04's small non-escaping map and non-escaping interface box were stack-allocated, so two "traps" cost nothing — fixed by making them realistic. (3) **Exercise 3's contract was impossible as first written**: with a `[]byte` API, `Put` must box a three-word slice header into a one-word interface, so the zero-allocation budget could never be met — my own reference solution failed my own test. The signature is now `*[]byte`, and the reason is the exercise's opening question. |
| 2026-09-05 | Demo 01 turned up a distinction worth its own callout: `-gcflags=-m` printing `make([]byte, n) does not escape` while the benchmark reports one allocation per call. Both are true — "does not escape" means nothing outlived the function, not that the value went on the stack, and a stack frame's size is fixed at compile time so a run-time-sized allocation cannot live there. Written into NOTES §4 and into the demo's closing narration as the most common way to misread escape-analysis output. |
| 2026-09-04 | **Module 1 NOTES rewritten in full**, on his instruction, after the incremental analogy removal left the prose uneven. Every section now states the mechanism directly in complete sentences: a slice variable is an address, a length and a capacity; `append` follows two rules chosen by comparing length to capacity; a function receives a copy of the three values and can change elements but never the caller's variable; the collector frees an array whole or not at all; map entries have no stable address; map reads with a `[]byte` key do not allocate and map writes do. All figurative language removed, including from the demo headers and exercise comments. Interview questions in §7 and `QUIZ.md` reworded to match. Technical content, tables, measured numbers and demo references unchanged. |
| 2026-09-04 | **All analogies stripped from Module 1's notes**, on his instruction — the "shelf and bookmark" metaphor for arrays and slices was making the material harder, because it forced him to hold the metaphor and the mechanism in his head at once and map between them. Sections 1, 2 and 3 now explain the mechanism directly and in complete sentences: a slice variable holds an address, a length and a capacity, and every consequence is reasoned in those terms. Tables, numbered traces, measured numbers and the demos are unchanged — those show the mechanism rather than standing in for it. This applies to every future module and to the other tracks. |
| 2026-09-04 | NOTES §3 rewritten from a question: the `scale` vs `addOne` pair was too compressed, and its comment "you MIGHT see this" was wrong about the caller specifically — the caller can *never* see an appended element, because a function gets a copy of the header and the caller's `len` is unchanged. Section now separates "the note" from "the shelf", carries a three-way table (full → new shelf, nobody sees it / spare room → written in place, anyone with a longer `len` sees it), shows the three-party `all`/`first` example that is the shift-splitter bug with the names changed, and closes with the rule: a function can change your elements, never your header — which is *why* `append`'s result must be reassigned. |
| 2026-09-04 | Notes corrected during verification. Demo 04 initially claimed the textbook rule that `m[string(b)]` avoids an allocation *because it is a map key*, and printed 1 alloc for both forms — the claim was wrong. Measured on Go 1.26: reads are free (`v := m[string(b)]`, comma-ok, and even `k := string(b); v := m[k]`, which escape analysis keeps on the stack) and writes always allocate (`m[string(b)]++`, `m[string(b)] = v`), because a write may have to store the key permanently. NOTES §5 and §6 and demo 04 rewritten around the measured numbers; exercise 4 rebuilt on the pointer-value counter shape that the real rule implies. `staticcheck`'s SA6001 asserts the same folklore, so that one line carries a `//lint:ignore` and a comment explaining why the analyser is wrong. |
