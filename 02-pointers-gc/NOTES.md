# Module 2 — Pointers, the Stack, and the Garbage Collector

**This module has no written notes. It has a reading list.**

The people who wrote the articles below know this material better than I can
summarise it, and several of them wrote the compiler and runtime code in
question. Reading the original is better than reading my paraphrase of it. What
this file gives you instead is the order to read in, what to extract from each
piece, which demo to run against it, and the questions you should be able to
answer afterwards.

Every link was checked on 2026-09-05.

---

## 0. Why this module exists

Module 1 gave you the allocation budget as a gate. You can now tell that a
function allocates. You cannot yet tell **why**, and you cannot yet tell what
that allocation costs after it happens.

This module answers both. Escape analysis is the compiler deciding whether a
value lives on the stack or the heap, and it is the answer to "why does this
allocate?" The garbage collector is what runs later because of that decision,
and it is the answer to "what does it cost?" War-story scenario 1 on this
track — the telemetry collector spending 60% of its CPU in GC — is exactly the
gap between those two questions.

## How to work this module

For each section below:

1. Read the **core** items. They are not optional; the exercises assume them.
2. Run the demo the section points at, and check that what you just read
   predicts what you see.
3. Answer the **self-check** questions out loud. If you cannot, re-read before
   moving on rather than continuing and hoping.
4. Read the **going deeper** items when you want them, or when a self-check
   question exposed a gap. They are genuinely optional.

The estimated reading time is about four hours for the core items across all
seven sections. Do not try to do it in one sitting.

---

## 1. Pointers

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| A Tour of Go — Pointers | https://go.dev/tour/moretypes/1 | The mechanical basics: `&`, `*`, and the `*T` type. Five minutes, and it runs in the browser. |
| A Tour of Go — Pointers to structs | https://go.dev/tour/moretypes/4 | Why `p.X` works when `(*p).X` is what you meant. |
| Alex Edwards — A Gentle Introduction to Pointers | https://www.alexedwards.net/blog/a-gentle-introduction-to-pointers | The clearest single explanation of pointers in Go I can point you at. Read it even if the Tour felt easy. |
| Alex Edwards — Demystifying Function Parameters in Go | https://www.alexedwards.net/blog/demystifying-function-parameters-in-go | Everything in Go is passed by value. This article makes that statement stop being confusing, including why maps and channels appear to break it. They do not. |
| Go FAQ — new vs make, and passing by value | https://go.dev/doc/faq | Read the entries "What's the difference between new and make?" and "When are function parameters passed by value?" |

**What to extract**

- Every argument in Go is copied. A pointer argument copies the pointer, not
  the thing it refers to. This is the same rule you learned for slices in
  Module 1, section 3, stated for the whole language.
- A map or channel parameter behaves like a pointer parameter because the map
  and channel *values themselves* are pointers to runtime structures. It is not
  an exception to the rule.
- `new(T)` returns a `*T` pointing at a zeroed `T`. `make` is only for slices,
  maps and channels, and it returns an initialised value, not a pointer.

**Self-check**

1. A function takes a `*Order` and assigns `*p = Order{}`. Does the caller see
   the change? What if the function assigns `p = &Order{}` instead?
2. Why does a `map[string]int` parameter let a function add entries that the
   caller sees, when Go copies every argument?
3. When would you write `new(T)` in real code rather than `&T{}`?

---

## 2. Value semantics and pointer semantics

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| Ardan Labs — Design Philosophy On Data And Semantics | https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html | William Kennedy's rule for choosing between a value and a pointer, based on the *nature of the data* rather than on size or convenience. This is the part most Go developers never get taught and then argue about in code review for years. |
| Effective Go — Pointers vs. Values | https://go.dev/doc/effective_go#pointers_vs_values | The official short version. |

**What to extract**

- The choice is driven by whether the data is something you copy (a value) or
  something you share (an entity with identity). Not by how many bytes it is.
- Being consistent within a type matters more than being right in any single
  method. Mixing semantics for one type is where bugs come from.
- This decision has an allocation consequence, which is section 4's material:
  choosing pointer semantics is often what sends a value to the heap.

**Self-check**

1. A `time.Time` is 24 bytes and its methods use value receivers. A
   `sync.Mutex` is 8 bytes and must never be copied. Explain both choices
   without mentioning size.
2. You have a `Config` struct read by many goroutines and never modified after
   startup. Value or pointer? Defend it.

---

## 3. The stack

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| Ardan Labs — Language Mechanics On Stacks And Pointers | https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-stacks-and-pointers.html | Part 1 of the four-part series that is the best thing written on this subject. Frames, what "sharing down" and "sharing up" mean, and why the distinction decides everything in section 4. |
| Cloudflare — How Stacks are Handled in Go | https://blog.cloudflare.com/how-stacks-are-handled-in-go/ | Why a goroutine can start at 2 KB and grow, what `morestack` does, and why Go moved from segmented to contiguous stacks. Explains the demo in this section directly. |

**Going deeper**

| Source | Link |
|---|---|
| Dave Cheney — Five things that make Go fast | https://dave.cheney.net/2014/06/07/five-things-that-make-go-fast |

**What to extract**

- Each goroutine has its own stack, it starts small, and it grows by
  **allocating a bigger stack and copying the old one into it**.
- Because the stack is copied, every pointer that referred into it must be
  rewritten by the runtime during the copy. That is why Go can move stacks and
  C cannot.
- Stack memory costs nothing to reclaim. The frame disappears when the function
  returns. This is the entire reason section 4 matters.

▶ **Run:** `go run ./02-pointers-gc/demo/02-stack` — this demo takes the
address of a local variable, recurses deep enough to force the stack to grow,
and prints the address of that same variable again. The address changes. You
are watching the runtime move a stack and rewrite the pointers into it.

**Self-check**

1. A goroutine's stack grows from 2 KB to 8 KB. What happened to the data that
   was on the old stack, and what happened to every pointer that referred to it?
2. Why is reclaiming stack memory free, while reclaiming heap memory needs a
   garbage collector?
3. You start 100,000 goroutines. Roughly how much memory is that before any of
   them does anything, and why is the answer not 100,000 × 1 MB?

---

## 4. Escape analysis

This is the centre of the module. Everything else here supports it.

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| Ardan Labs — Language Mechanics On Escape Analysis | https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-escape-analysis.html | Part 2 of the series. Read it slowly. The "sharing up the call stack" versus "sharing down" framing is the thing to take away. |
| Go FAQ — heap or stack? | https://go.dev/doc/faq#stack_or_heap | Four paragraphs, and the official position: from a correctness standpoint you do not need to know. From a performance standpoint you do. |
| freeCodeCamp — Understanding Escape Analysis in Go | https://www.freecodecamp.org/news/understanding-escape-analysis-in-go/ | Worked examples with the compiler output next to each one. |
| JetBrains — Escape Analysis in Go: Stack vs. Heap Allocations Explained | https://blog.jetbrains.com/go/2026/07/20/escape-analysis/ | Recent, and good on reading `-gcflags=-m` output specifically. |

**Going deeper**

| Source | Link |
|---|---|
| Medium — Golang escape analysis (mannharleen) | https://medium.com/@mannharleen/golang-escape-analysis-33e9cc0563ee |
| Medium — Go: Memory Management and Allocation (Vincent, A Journey With Go) | https://medium.com/a-journey-with-go/go-memory-management-and-allocation-a7396d430f44 |

**The command you must be able to run and read**

```powershell
go build -gcflags='-m -l' ./02-pointers-gc/demo/01-escape
```

`-m` prints the compiler's allocation decisions. `-l` disables inlining, so the
output describes the functions you actually wrote instead of the merged
versions. `-m -m` prints more detail about *why*.

**What to extract**

- The compiler moves a value to the heap when it cannot prove the value is dead
  by the time the function returns. "Escapes to heap" is a statement about
  proof, not about pointers as such.
- Returning a pointer to a local is the obvious case. The non-obvious cases are
  the ones that cost real money: storing a value in an interface, capturing a
  variable in a closure that outlives the call, passing anything to a variadic
  `...any` function such as `fmt.Println`, and slices whose size is not known at
  compile time.
- The compiler is conservative. If it cannot see inside a call, it assumes the
  worst.
- **"Does not escape" does not mean "on the stack."** These are two separate
  conditions and a value needs both to stay off the heap: nothing may outlive
  the function, *and* the size must be known at compile time, because a stack
  frame's layout is fixed when the function is compiled. `make([]byte, n)` with
  a variable `n` gets `does not escape` from `-gcflags=-m` and still allocates
  on every call. The demo shows exactly this row. It is the single most common
  way to misread escape analysis output, and measuring is what catches it.

▶ **Run:** `go run ./02-pointers-gc/demo/01-escape` — eight small functions,
each printed with its measured allocations per call. Then build the same
package with `-gcflags='-m -l'` and match every line of compiler output to a
row of the demo's table.

▶ **Run:** `go run ./02-pointers-gc/demo/04-escape-traps` — the failure-mode
demo. Five functions that look allocation-free and are not. Predict each one
before you read the number.

**Self-check**

1. Explain "escapes to heap" without using the word pointer.
2. `fmt.Println(x)` causes `x` to escape even when `x` is an `int`. Why?
3. Two functions are identical except one returns `*Order` and the other returns
   `Order`. Which allocates, and is the answer always the same?
4. Why does `-l` matter when reading `-gcflags=-m` output?

---

## 5. The garbage collector

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| A Guide to the Go Garbage Collector (official) | https://go.dev/doc/gc-guide | The authoritative document, written by the people who maintain the collector. It is interactive and it is the one to read properly. Focus on the sections on the GOGC knob, the cost model, and latency. |
| Go blog — Getting to Go: The Journey of Go's Garbage Collector | https://go.dev/blog/ismmkeynote | Rick Hudson's ISMM keynote. History and, more useful to you, *why* the design chose latency over throughput. This is the article to read before an interview question about Go's GC. |

**Going deeper**

| Source | Link |
|---|---|
| Medium — Go: How Does the Garbage Collector Mark the Memory? (Vincent) | https://medium.com/a-journey-with-go/go-how-does-the-garbage-collector-mark-the-memory-72cfc12c6976 |
| Medium — Go: How Does the Garbage Collector Watch Your Application? (Vincent) | https://medium.com/a-journey-with-go/go-how-does-the-garbage-collector-watch-your-application-dbef99be2c35 |
| GOMEMLIMIT — the original design proposal (Michael Knyszek) | https://github.com/golang/proposal/blob/master/design/48409-soft-memory-limit.md |
| Weaviate — GOMEMLIMIT is a game changer for high-memory applications | https://weaviate.io/blog/gomemlimit-a-game-changer-for-high-memory-applications |

**What to extract**

- Go's collector is concurrent, tri-colour, mark-and-sweep, and it is
  **non-moving** — an object's address never changes once allocated. Contrast
  this with the stack, which does move (section 3).
- `GOGC` sets the ratio of new heap to live heap that triggers the next cycle.
  `GOGC=100`, the default, means "collect when the heap has doubled since the
  last collection."
- The collector's cost is paid in CPU time stolen from your program, not
  primarily in pause length. Modern Go pauses are sub-millisecond; the tax is
  the 25% of CPU the collector may take while running.
- `GOMEMLIMIT` is a soft limit that makes the collector work harder as you
  approach it. It is what you reach for in a container with a memory cap.

▶ **Run:** `go run ./02-pointers-gc/demo/03-gc` — allocates garbage in a loop
and prints the number of collections, heap size, and total pause time. It then
repeats the same workload at `GOGC=400` and under a `GOMEMLIMIT`, so you can see
each knob move a number you care about.

**Self-check**

1. What does `GOGC=100` actually mean? What would `GOGC=off` do to a service
   that allocates steadily?
2. Your service's heap is stable but CPU is high and the profile blames
   `runtime.gcDrain`. What is happening, and what is the fix — fewer bytes or
   fewer objects?
3. Why does Go's collector not move objects, and what does it give up by not
   moving them?
4. When is `GOMEMLIMIT` the right tool, and what happens if you set it too low?

---

## 6. Measuring allocation

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| Dave Cheney — How to write benchmarks in Go | https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go | `-benchmem`, `B/op`, `allocs/op`, and how not to write a benchmark the compiler optimises away. |
| Ardan Labs — Language Mechanics On Memory Profiling | https://www.ardanlabs.com/blog/2017/06/language-mechanics-on-memory-profiling.html | Part 3 of the series. Connects the escape analysis output to a real profile. |
| Go blog — Profiling Go Programs | https://go.dev/blog/pprof | The original, still the best walkthrough of the tool. |

**Going deeper**

| Source | Link |
|---|---|
| Julia Evans — Profiling Go programs with pprof | https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/ |
| Go — Diagnostics | https://go.dev/doc/diagnostics |
| Dave Cheney — High Performance Go Workshop | https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html |

**What to extract**

- `allocs/op` counts allocations; `B/op` counts bytes. They are different
  problems. Many small allocations hurt the collector; few large ones hurt your
  memory ceiling.
- `-alloc_space` versus `-inuse_space` in pprof answers two different
  questions: what did this program ever allocate, and what is it holding right
  now. A leak is an `inuse_space` question. GC pressure is an `alloc_space`
  question.
- `testing.AllocsPerRun`, which Module 1's gate already uses, is the cheap
  assertion. A benchmark with `-benchmem` is the measurement. A profile is the
  investigation.

**Self-check**

1. A benchmark reports `120 ns/op, 48 B/op, 2 allocs/op`. Which of those three
   numbers do you attack first, and how do you decide?
2. You suspect a leak. Which pprof profile, and which flag?
3. Why can a benchmark report zero allocations for code that obviously
   allocates?

---

## 7. `sync.Pool`, honestly

**Core reading**

| Source | Link | Why this one |
|---|---|---|
| `sync.Pool` package documentation | https://pkg.go.dev/sync#Pool | Read the doc comment first, particularly the sentence about items being removed at any time without notification. Most misuse comes from not believing that sentence. |
| VictoriaMetrics — Go sync.Pool and the Mechanics Behind It | https://victoriametrics.com/blog/go-sync-pool/ | Phuong Le, and it is thorough: per-P private and shared slots, the victim cache, and the two-GC-cycle lifetime. Long, and worth it. |

**Going deeper**

| Source | Link |
|---|---|
| Medium — Go: Understand the Design of Sync.Pool (Vincent) | https://medium.com/a-journey-with-go/go-understand-the-design-of-sync-pool-2dde3024e277 |
| WunderGraph — Golang sync.Pool | https://wundergraph.com/blog/golang-sync-pool |

**What to extract**

- A pool is a cache with no guarantees. `Get` may return a new object at any
  time, and anything you `Put` may be discarded at the next collection.
- It is for short-lived, uniformly sized, frequently allocated objects. The
  standard library uses it for HTTP/2 frame buffers, which is exactly that
  shape.
- Two failure modes to remember: not resetting an object before reuse, which
  leaks one request's data into another, and pooling objects of wildly varying
  size, which makes the pool hold the largest one forever.
- Measure before and after. A pool can be slower than allocating.

▶ **Run:** `go run ./02-pointers-gc/demo/05-pool` — measures the same workload
with and without a pool, then shows the pool being emptied by a garbage
collection, then shows the un-reset-object bug producing wrong output.

**Self-check**

1. Why can `sync.Pool` not be used as a connection pool or a cache with a size
   limit?
2. What is the victim cache, and how many collections does a pooled object
   survive?
3. You pool `[]byte` buffers. Most are 1 KB; one request a day is 40 MB. What
   goes wrong, and what is the fix?

---

## 8. Interview questions this module answers

Answer these out loud, without notes.

1. What decides whether a Go value lives on the stack or the heap? Who makes
   that decision and when?
2. Walk me through what `go build -gcflags=-m` prints and how you use it.
3. Why does putting an `int` into an `interface{}` allocate?
4. Describe Go's garbage collector: algorithm, what it does concurrently, and
   what it costs.
5. What does `GOGC` control, and when have you changed it?
6. A service has stable memory and rising CPU, and the profile blames the
   runtime. How do you diagnose it?
7. When is `sync.Pool` the right answer, and when is it a mistake?
8. Goroutine stacks start at 2 KB and grow. How, and what does the runtime have
   to fix up when they do?

## 9. Files in this module

```
02-pointers-gc/
  NOTES.md            you are here — a reading list, not notes
  MISTAKES.md         every real bug made in this module, with cause and fix
  QUIZ.md             the oral quiz tracker
  demo/
    01-escape/        eight functions with measured allocations, to read beside -gcflags=-m (§4)
    02-stack/         a local variable's address changing as the stack grows and is copied (§3)
    03-gc/            collections, heap size and pause time under GOGC and GOMEMLIMIT (§5)
    04-escape-traps/  DELIBERATELY MISLEADING: five functions that allocate and do not look like it (§4)
    05-pool/          sync.Pool measured, emptied by a GC, and misused (§7)
  exercises/
    README.md         the rules, the exercise table, and the definition of done
    ex1_format.go     implement: build a CSV line into a caller's buffer with zero allocations
    ex2_escape.go     fix the bug: stop a hot path escaping to the heap, using -gcflags=-m as evidence
    ex3_pool.go       implement: a correct buffer pool, with reset and a size guard
    ex4_boxing.go     fix the bug: a logging helper that allocates on every call because of interface boxing
```

```powershell
go run ./02-pointers-gc/demo/01-escape                    # run any demo
go build -gcflags='-m -l' ./02-pointers-gc/demo/01-escape  # the compiler's own reasoning
go test ./02-pointers-gc/...                              # fast test loop
.\check.ps1 ./02-pointers-gc/...                          # the gate
.\check.ps1 -Bench ./02-pointers-gc/...                   # the gate plus allocation counts
```
