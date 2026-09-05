# Mistakes Log — Module 1: Memory & Layout

Every real bug made while solving this module's exercises — what was written,
why it broke, what fixed it. Fill this in as you go, not at the end.

**Patterns worth rereading first**, carried forward from Track 1's log — these
are the ones you hit more than once there:

- Slice-range-as-index (`for i, v := range xs` and then using `v` where `i`
  belonged, or the reverse). Hit twice in Track 1 Module 4.
- Length versus capacity confusion. Hit in Track 1 Module 4. This module is
  *about* that distinction, so expect it to appear again in a different form.
- Shadowed variables from `:=` inside a loop or an `if`. Hit twice in Track 1
  Module 6, in `ex3_retry`.

**New traps native to this module**, predicted at scaffold time. Delete the
ones you never hit; that is useful information too:

- Writing `orders[:n]` when you needed `orders[:n:n]`, and finding it only
  because a *different* slice changed.
- Nil-ing the dropped tail with an off-by-one — either `clear(orders[w:])` written
  as `clear(orders[w+1:])`, or forgetting it entirely because the tests for
  ordering already passed.
- In the ring buffer: `Snapshot` returning a window onto the ring's own array
  instead of appending copies, so an operator's snapshot changes under them.
- In the ring buffer: getting `start` wrong for the not-yet-full case, so a
  half-full ring reports its readings rotated.
- In the tag counter: reaching for `map[string]int64` and then discovering the
  budget test, because `++` on a map value is a map write.
- In the tag counter: storing the caller's `[]byte` somewhere instead of
  converting it, so every tag becomes the last frame read.
- Modulo on a negative index. Go's `%` keeps the sign of the dividend.

---

## Exercise 2 — SplitShifts

*(not started)*

## Exercise 1 — Ring

*(not started)*

## Exercise 3 — CompactActive

*(not started)*

## Exercise 4 — TagCounter

*(not started)*

---

### Entry format

**Bug N — short label.** What you literally wrote, and the exact test that
failed with its exact output.

Why: the mechanism, in plain English.

**Fix:** what replaced it.

Lesson: one generalisable rule.
