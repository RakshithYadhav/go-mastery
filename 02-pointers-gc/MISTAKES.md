# Mistakes Log — Module 2: Pointers, the Stack, and the Garbage Collector

Every real bug made while solving this module's exercises — what was written,
why it broke, what fixed it. Fill this in as you go, not at the end.

**Patterns worth rereading first**, carried forward:

- Length versus capacity (Module 1). Exercise 3 turns on it: resetting a
  pooled buffer means resetting its length and keeping its capacity, and
  getting that backwards makes the pool pointless rather than broken, which is
  harder to notice.
- Map writes with a `[]byte`-to-`string` key allocate; map reads do not
  (Module 1, section 5). One of exercise 2's three causes is exactly this, and
  the compiler will not tell you about it.
- Shadowed variables from `:=` inside a loop or an `if` (Track 1, Module 6,
  hit twice).

**New traps native to this module**, predicted at scaffold time. Delete the
ones you never hit; that is useful information too:

- Reading `does not escape` from `-gcflags=-m` and concluding the value is on
  the stack. It means nothing outlived the function, not that the allocation
  went away. A variable-sized `make` gets that message and still allocates.
- Forgetting `-l` and then trying to match compiler output against functions
  that were inlined out of existence.
- Writing a benchmark whose work the compiler deletes, then celebrating
  `0 allocs/op`. Assign to a package-level sink.
- In exercise 1: reaching for `strconv.FormatFloat` (returns a new string,
  allocates) instead of `strconv.AppendFloat` (writes into your buffer).
- In exercise 1: getting `AppendFloat`'s argument order or its `bitSize`
  wrong, and producing `214.50000000000003`.
- In exercise 3: storing `[]byte` in the pool rather than `*[]byte`, so every
  `Put` allocates to box the three-word slice header into a one-word
  interface.
- In exercise 3: putting the reset on `Put` instead of `Get`, which works until
  a caller returns early or panics and skips its `Put`.
- In exercise 4: moving the level check to the first line of the function and
  expecting it to help. The boxing already happened at the call site.

---

## Exercise 1 — AppendFrame

*(not started)*

## Exercise 4 — Logger boxing

*(not started)*

### Which fix I chose, and why

*(two sentences, once you have chosen — the reasoning is the exercise)*

## Exercise 2 — ShiftStats

*(not started)*

### The cause the compiler did not report, and why it did not

*(answer this one explicitly; it is the most valuable thing in the module)*

## Exercise 3 — BufferPool

*(not started)*

---

### Entry format

**Bug N — short label.** What you literally wrote, and the exact test that
failed with its exact output.

Why: the mechanism, in plain English.

**Fix:** what replaced it.

Lesson: one generalisable rule.
