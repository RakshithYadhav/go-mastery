# Module 1 — Quiz Tracker

Oral quiz over the questions in `NOTES.md` §7. Graded honestly; the roadmap box
does not get ticked until every question clears.

Answer out loud, without notes and without the code open. Two of these
(Q2, Q8) have a "it depends, and here is exactly what it depends on" shape —
an answer that states the rule without naming the condition is a partial, not
a pass.

## Round 1 — not started

**Q1 (what a slice is).** What does a slice variable contain? Name its fields
and state its size on a 64-bit machine.

**Q2 (append visibility).** A function receives a slice and appends to it. Can
the caller read the appended element? Can any other code read it? State the
conditions precisely.

**Q3 (full slice expression).** What does `s[1:3:4]` mean, and when would you
write the third number on purpose?

**Q4 (the split bug).** Two slices are made from one array with `a[:2]` and
`a[2:]`. You append one element to the first. What happens to the second, and
why does no tool report it?

## Round 2 — not started

**Q5 (map addresses).** Why is `&m["key"]` a compile error when `&s[0]` is
legal?

**Q6 (retention).** A function returns a 60-byte slice of a 40 MB file's
contents. What is the memory cost of holding that return value, and why?

**Q7 (iteration order).** Why does the runtime randomise map iteration order,
and what does that require of your code?

**Q8 (string keys).** Reading a map with `m[string(b)]` allocates nothing, but
`m[string(b)]++` allocates on every call. Explain the difference, and describe
the structure you would use for a counter on a hot path.

## Open follow-ups (answer these to close a round)

*(none yet)*

---

*Answers are graded live in chat; this file is the persistent summary so the
quiz resumes across sessions without re-asking cleared questions. Record the
content of the answer, not just the verdict — what was right, what got
tightened, which misconception was corrected.*

*Quiz backlog across this track: Module 1 (both rounds). Track 1's backlog is
parked, not cleared — Module 3 Q6–Q8 and all of Modules 4, 5, 6, 7 in
`../../go-concurrency/`.*
