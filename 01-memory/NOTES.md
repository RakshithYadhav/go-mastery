# Module 1 — Memory & Layout: Slices, Maps, Strings

*Read this from top to bottom. Run each demo when the text tells you to.*

This module explains what a slice, a map, and a string actually are in memory.
Once you know that, you can answer four questions that come up constantly in
real Go code: when an assignment copies data and when it copies only a
reference, why `append` sometimes changes data that another variable can see,
why you cannot write `&m[key]`, and why a slice of 60 bytes can keep 40
megabytes of memory from being freed.

---

## 0. Why this module exists

There are two reasons, and both come from work you have already done.

**First.** In Track 1, Module 4, you built pipelines and worker pools that
passed slices between goroutines. Nothing broke. But if any stage had appended
to a slice it received, it could have written into memory that another
goroutine was reading, and the race detector would not necessarily have
reported it, because each goroutine would have been using a different slice
variable. The two variables would simply have referred to the same array. You
cannot reason fully about concurrent code until you can state precisely what is
shared when a slice is passed to a function. This module gives you that.

**Second.** The `webhook-go` shutdown bug you found on 2026-09-04 was a
context derived from a parent that had already been cancelled, which made the
ten-second grace period zero on every run. That was not a concurrency bug. The
value was valid, it compiled, and it was built from the wrong input. Slices,
maps, and strings each have a bug of the same kind: a value that is valid,
prints correctly, passes a test on the normal path, and is wrong underneath.
This module shows you each one.

---

## 1. Arrays and slices are different types

### The problem

Most Go tutorials introduce `[]int{1, 2, 3}` and move on. As a result, most
people believe a slice is a list of values. It is not. A slice is a small
structure that describes a region of an array that exists somewhere else. Every
surprising behaviour of slices follows from that fact.

### What an array is

An array is a fixed-size sequence of values. Its length is part of its type, so
`[5]int` and `[6]int` are different types. An array variable holds the values
directly, and assigning an array copies every element:

```go
a := [5]int{10, 20, 30, 40, 50}
b := a          // copies all five values; a and b are independent
b[0] = 999      // a[0] is still 10
```

### What a slice is

A slice variable holds exactly three values, and none of them is an element:

```
a slice variable ([]int) contains:

  address   the memory address of the first element this slice can access
  length    the number of elements this slice is allowed to read and write
  capacity  the number of elements from that address to the end of the array
```

On a 64-bit machine each of these is 8 bytes, so a slice variable is 24 bytes
regardless of how many elements it refers to. The elements live in an array
elsewhere in memory, and the address field says where that array begins.

Assigning a slice copies those three values. It does not copy the array. After
the assignment, both variables hold the same address, so both refer to the same
elements:

```go
s := []int{10, 20, 30, 40, 50}
t := s          // copies address, length, and capacity; the array is shared
t[0] = 999      // s[0] is now 999 as well
```

Compare the two code blocks. `b := a` on an array copies five integers.
`t := s` on a slice copies three numbers and leaves the integers where they
are. The syntax is identical and the behaviour is opposite. Nothing at the
assignment site tells you which case you are in; you have to know the type.

### Numeric walkthrough

Let `base` be an array of eight integers. The table below shows the address,
length, and capacity of each slice made from it. The address is written as an
index into `base` rather than as a hexadecimal number.

| Expression | address | length | capacity | Can read | Why the capacity is that number |
|---|---|---|---|---|---|
| `base := [8]int{0,1,2,3,4,5,6,7}` | — | — | — | all 8 | `base` is an array, not a slice |
| `s := base[2:5]` | index 2 | 3 | 6 | 2, 3, 4 | from index 2 to the end of `base` is 6 elements |
| `s2 := s[1:2]` | index 3 | 1 | 5 | 3 | `s2` starts at `base` index 3; from there to the end is 5 elements |
| `s3 := s[:cap(s)]` | index 2 | 6 | 6 | 2 through 7 | re-slicing up to the capacity is legal and exposes elements past the old length |
| `s4 := base[2:5:6]` | index 2 | 3 | 4 | 2, 3, 4 | the third number sets the capacity: 6 − 2 = 4 |

Two rows deserve attention.

The third row shows that the length is a limit on what a slice may read, not a
boundary on what exists. `s` has length 3, but `s[:cap(s)]` produces a slice of
length 6 over the same array, including elements that `s` was never given. The
elements past the length are present in memory and can be reached by re-slicing.

The last row uses the **full slice expression**, written `low:high:max`. The
third number sets the capacity of the result to `max − low`. This is how you
produce a slice that cannot reach beyond a chosen index of the underlying array.
It is the fix for the bug in Section 3, and most Go developers have never
written one.

▶ **Run:** `go run ./01-memory/demo/01-header` — prints the address, length,
and capacity at each step of the table above, using real addresses. Confirm that
`s2` starts 8 bytes after `s` and that no element was copied at any step.

---

## 2. `append`

### The problem

`append` is the only builtin whose result you are required to assign:

```go
s = append(s, 4)     // correct
append(s, 4)         // compile error: result of append is not used
```

The reason is that `append` sometimes writes into the array the slice already
refers to, and sometimes allocates a new, larger array and copies everything
into it. In the second case the old address is no longer the right one, so
`append` must return a new slice value, and you must store it.

### The rule

`append` follows exactly two rules:

1. **If the length is less than the capacity**, the array has unused space.
   `append` writes the new element into the next unused index of the existing
   array, increases the length by one, and returns a slice with the same
   address. Every other slice that refers to this array now has a changed
   element in it, whether or not that slice's length lets it read the change.
2. **If the length equals the capacity**, the array is full. `append` allocates
   a larger array, copies every existing element into it, writes the new
   element, and returns a slice with the new address. Every other slice that
   refers to the old array continues to refer to the old array. It will not see
   this write or any later write to the new array.

Which rule applies depends on the capacity, and the capacity is not visible at
the call site. Two `append` calls that look identical in source code can behave
differently because of what happened to the slice earlier.

### Numeric walkthrough

Start with `s := make([]int, 0, 4)`, which is a slice of length 0 and capacity
4 over a new four-element array.

| Step | Operation | length | capacity | Address changed? | What happened |
|---|---|---|---|---|---|
| 1 | `s = append(s, 1)` | 1 | 4 | no | wrote into index 0 of the existing array |
| 2 | `s = append(s, 2)` | 2 | 4 | no | wrote into index 1 |
| 3 | `s = append(s, 3)` | 3 | 4 | no | wrote into index 2 |
| 4 | `s = append(s, 4)` | 4 | 4 | no | wrote into index 3; the array is now full |
| 5 | `s = append(s, 5)` | 5 | 8 | **yes** | allocated a new array, copied 4 elements, wrote 5 into index 4 |

Step 5 is where sharing bugs happen. If another slice referred to the original
array, it saw the writes in steps 1 through 4 and did not see the write in step
5. Nothing in the source code changed between step 4 and step 5.

**The growth factor is not a fixed number.** In current Go versions the runtime
doubles the capacity while the slice is small, switches to growing by roughly
25% once the slice is large, and then rounds the result up to one of the memory
allocator's fixed size classes. The rounding is why you will see capacities such
as 6 and 10 where a pure doubling rule predicts 4 and 8. These thresholds have
changed between Go releases. Measure them; do not memorise them.

▶ **Run:** `go run ./01-memory/demo/02-append-growth` — appends 20,000
integers to a nil slice and prints a row each time the array is reallocated,
with the measured growth factor. It then repeats the workload with
`make([]int, 0, 20000)` and reports the number of reallocations and the total
number of elements copied in each case.

### The practical rule

If you know approximately how many elements the result will hold, allocate that
capacity once:

```go
out := make([]Record, 0, len(in))   // one allocation; no element is ever copied
for _, r := range in {
    out = append(out, transform(r))
}
```

The alternative, `var out []Record` followed by the same loop, performs about
fifteen reallocations for 20,000 elements and copies roughly 60,000 elements in
total. The output is identical. The cost is not. Exercise 1 depends on this
rule.

---

## 3. Sharing between slices

### Passing a slice to a function

When you pass a slice to a function, Go copies the address, the length, and the
capacity into the function's parameter. The function now has its own copy of
those three values. Both copies hold the same address, so both refer to the same
array. The array is not copied.

```go
func scale(xs []int) {
    for i := range xs {
        xs[i] *= 2          // the caller sees this change, every time
    }
}

func addOne(xs []int) {
    xs = append(xs, 1)      // the caller never sees this change
}
```

`scale` uses the address to modify the elements. It never assigns to `xs`. The
elements live in the array, the caller's slice refers to the same array, and so
the caller reads the modified values. This is true in every case.

`addOne` assigns to `xs`. The statement `xs = append(xs, 1)` stores a new
address, length, and capacity into the function's own copy. The caller's three
values are not affected, because the function never had access to them. When
the function returns, its copy is discarded. The caller's length was 3 before
the call and is 3 after the call, so the caller cannot read an appended
element. This is not sometimes true. It is always true.

What varies is whether `append` modified the array that the caller's slice
refers to. There are two cases:

| Caller's slice | What `append` does | Is the caller's array modified? | Which slices can read the new element? |
|---|---|---|---|
| length equals capacity | allocates a new array, copies the elements, writes there | no | none; the new array is unreachable once the function returns |
| length is less than capacity | writes the new element into the next unused index of the existing array | **yes** | any slice over that array whose length includes that index |

The caller is not among those slices, because the caller's length ends one
index before the new element. But a different slice that refers to the same
array with a greater length can read that index, and it will find a value that
no code ever assigned to it through that slice:

```go
all := []int{10, 20, 30, 0, 0}
first := all[:3]      // length 3, capacity 5
addOne(first)         // writes 1 into index 3 of the shared array
fmt.Println(first)    // [10 20 30]     — length 3, cannot read index 3
fmt.Println(all)      // [10 20 30 1 0] — length 5, reads index 3
```

No code wrote to `all`. No index was out of range. `all[3]` changed anyway.

The rule that follows is this: a function can modify the elements of a slice
you pass to it, but a function cannot modify your slice variable. `append` must
return a new length and possibly a new address, and Go has no mechanism for a
function to write those into the caller's variable. Returning the value and
requiring you to assign it is the only mechanism available. `s = append(s, x)`
is not a style preference.

### The bug in exercise 2

The version of this that reaches production looks like the following. One slice
is divided into two, and both parts refer to the same array:

```go
all := []string{"WO-1", "WO-2", "WO-3", "WO-4"}

morning := all[:2]                   // length 2, capacity 4
evening := all[2:]                   // length 2, capacity 2

morning = append(morning, "RUSH-9")  // length 2 < capacity 4, so this writes
                                     // into index 2 of the array — which is
                                     // evening[0]
```

After the append, `evening[0]` is `"RUSH-9"`. No code assigned to `evening`.
No index was out of range. There is no panic, no `go vet` warning, and no data
race, because every operation was legal. A work order was removed from the
evening schedule and nothing recorded that it happened.

The fix is the full slice expression from Section 1. Setting the capacity of
`morning` equal to its length removes the unused space that `append` would
otherwise write into:

```go
morning := all[:2:2]                 // length 2, capacity 2
```

Now `append` finds the length equal to the capacity, allocates a new array,
copies the two elements, and writes `"RUSH-9"` into the new array. The original
array is not modified and `evening` is unaffected.

The general rule: when you give another piece of code a sub-slice of an array
you still use, or when you keep a sub-slice of an array you have given away,
either limit its capacity with the three-index form or copy it with
`slices.Clone`. Make the choice deliberately each time.

▶ **Run:** `go run ./01-memory/demo/03-aliasing` — this demo is deliberately
broken. It performs the split and append above, prints the corrupted `evening`
slice beside the intended one, then repeats the operation using the three-index
form. Compare the array addresses it prints: in the broken run they are equal,
and in the fixed run they differ from the `append` onward.

---

## 4. A small slice keeps a large array alive

The garbage collector frees an array when no slice or pointer refers to any
part of it. An array is a single allocation. It is either entirely reachable or
entirely free; the collector does not free the part of an array that no slice
covers.

The following function therefore causes a memory leak:

```go
func firstLine(fileContents []byte) []byte {
    i := bytes.IndexByte(fileContents, '\n')
    return fileContents[:i]
}
```

If `fileContents` is a 40 MB file, the returned slice has a length of perhaps
60 bytes and an address inside the 40 MB array. For as long as the caller holds
that return value, the entire 40 MB array remains reachable and cannot be
freed. In a service that processes one such file per minute and keeps the first
line of each, memory use grows by 40 MB per minute until the process is killed.
This is war-story scenario 3 on this track, and it is one of the most common
memory leaks in production Go.

The fix is to copy the bytes you intend to keep into a new, correctly sized
array, so that nothing refers to the large one:

```go
line := fileContents[:i]
out := make([]byte, len(line))
copy(out, line)
return out                      // or: return slices.Clone(line)
```

The same problem occurs in the other direction when you shorten a slice of
pointers. The elements past the new length are still present in the array, and
each still holds a pointer to an object:

```go
items = items[:0]     // length 0, but every *Order is still in the array,
                      // and every one of them is still reachable
```

Reducing the length does not release those objects. Calling `clear(items)`
before shortening, or setting each dropped element to `nil`, does. Exercise 3
tests this directly.

▶ **Run:** `go run ./01-memory/demo/05-retain` — allocates a 40 MB buffer,
keeps a 49-byte slice of it, forces a garbage collection, and prints the heap
in use. It then repeats the sequence with a copied slice. The two heap figures
differ by approximately 40 MB.

---

## 5. Maps

### How a map is stored

A Go map is a hash table. The runtime divides the table into buckets, each of
which holds a small fixed number of key/value pairs. To find a key, the runtime
hashes it, uses part of the hash to select a bucket, and compares a few bits of
the hash against a compact array inside the bucket before performing any full
key comparison. When the table becomes too full, the runtime allocates a larger
bucket array and moves entries into it incrementally, a few on each subsequent
write, so that no single operation pays the full cost of rehashing.

You do not need to reproduce this structure. You need to know three
consequences of it.

### Consequence 1: map entries have no stable address

Because the runtime moves entries between buckets during growth, it does not
allow you to take the address of one:

```go
m := map[string]int{"a": 1}
p := &m["a"]        // compile error: cannot take the address of m["a"]
```

For the same reason, assigning to a field of a struct stored in a map is a
compile error:

```go
m := map[string]Order{"WO-1": {}}
m["WO-1"].Qty = 5   // compile error: cannot assign to struct field m["WO-1"].Qty
```

The expression `m["WO-1"]` produces a copy of the stored struct. Assigning to a
field of that copy would change nothing in the map, so the compiler rejects it
rather than compiling code that silently does nothing. You have two options:

```go
o := m["WO-1"]; o.Qty = 5; m["WO-1"] = o          // read the copy, modify it, store it back

m := map[string]*Order{...}; m["WO-1"].Qty = 5    // store pointers; the assignment goes through the pointer
```

Consider what would happen if Go allowed `&m["a"]`. An insert into the map
elsewhere in the program could trigger growth, the entry would move, and your
pointer would refer to memory that no longer holds it. The compile error exists
to prevent that. Slices differ here: `&s[0]` is legal because the array a slice
refers to is only replaced when you call `append`, and that call is visible in
your code.

### Consequence 2: iteration order is randomised

The order in which `range` visits a map's entries is not merely unspecified. The
runtime deliberately starts each iteration at a random position, so the order
differs between two `range` loops over the same unchanged map in the same
process. This is intentional. It prevents programs from accidentally depending
on an order that the language never guaranteed.

If you need a fixed order, sort the keys:

```go
keys := slices.Sorted(maps.Keys(m))   // Go 1.23 and later
for _, k := range keys {
    // ...
}
```

A test that iterates a map and compares the output to a fixed string will pass
on some runs and fail on others. Neither `go vet` nor `staticcheck` detects
this. Only your own discipline does.

### Consequence 3: `delete` does not shrink the map

If you insert one million entries into a map and then delete all of them, the
bucket array remains sized for one million entries. That memory is released
only when the map itself becomes unreachable. A long-lived map used as a cache
without eviction therefore grows to its maximum size and stays there. The
standard remedy is to build a new map periodically and replace the old one.

### Reading a map with a `[]byte` key does not allocate; writing does

Converting a `[]byte` to a `string` normally allocates a new string and copies
the bytes into it, because a string is immutable and the byte slice is not, so
they cannot share memory. However, the compiler omits the allocation when it can
prove that the string does not outlive the expression it appears in. A map read
is exactly that case.

The following figures were measured by `demo/04-maps` on Go 1.26:

```go
v := m[string(buf)]            // 0 allocations: the lookup hashes and compares, then the string is gone
_, ok := m[string(buf)]        // 0 allocations
k := string(buf); v := m[k]    // 0 allocations: escape analysis proves k never leaves the function

m[string(buf)]++               // 1 allocation on every call
m[string(buf)] = v             // 1 allocation on every call
```

The distinction is between reading and writing, not between one syntactic form
and another. A read can use the bytes temporarily because nothing retains the
string after the lookup completes. A write may need to store the key in the map
permanently, so a real string must be allocated first, and the compiler cannot
know in advance whether the key is already present.

This matters as soon as you write a counter. The statement
`counts[string(tag)]++` allocates once per call, indefinitely, even when the key
was inserted on the first call and never changes again. The following structure
avoids that cost:

```go
m := map[string]*int64{}
if c := m[string(tag)]; c != nil {   // read: 0 allocations
    *c++                             // increment through the pointer; not a map write
} else {
    k := string(tag)                 // 1 allocation, once per new tag
    var n int64 = 1
    m[k] = &n
}
```

The first branch runs on every call after the first for a given tag and
allocates nothing. The second branch runs once per distinct tag. Exercise 4 is
built on this structure, and its test asserts zero allocations for a known tag.
The `counts[string(tag)]++` version produces correct counts and fails that test.
That is why this track's gate includes an allocation budget.

▶ **Run:** `go run ./01-memory/demo/04-maps` — four measured experiments:
iteration order across repeated `range` loops over one map, read-modify-write
on struct values, memory retained after deleting every entry, and allocation
counts for each of the five `[]byte`-key forms above.

---

## 6. Strings and `[]byte`

A string variable holds two values: an address and a length. There is no
capacity field, because a string is immutable and can never be appended to.

Every cost in the table below follows from that immutability:

| Operation | Allocates? | Reason |
|---|---|---|
| `s[2:6]` (substring) | no | produces a new address and length over the same bytes; the entire original string remains reachable, which is the same leak as Section 4 |
| `string(byteSlice)` | yes | the bytes must be copied, because the slice can be modified later and the string must not change |
| `[]byte(str)` | yes | the bytes must be copied, because the result can be modified and the string must not change |
| `m[string(b)]` used to read a map | no | Section 5: the string does not outlive the lookup |
| `m[string(b)]` used to write a map | yes | Section 5: the key may be stored permanently |
| `s1 + s2` inside a loop | yes, on every iteration | each concatenation produces a new string; use `strings.Builder` instead |

The substring row is the one that causes production problems. A 20-character
substring of a 2 MB request body keeps the entire 2 MB reachable for as long as
the substring is held. `strings.Clone` exists in the standard library for this
case: it copies the substring into its own allocation so the original can be
freed.

Indexing a string returns a byte, not a character. In `"café"`, the `é` is
encoded as two bytes, so `"café"[3]` is the first of those two bytes and is not
a valid character on its own. `len(s)` returns the number of bytes, not the
number of characters. Module 7 covers this in full. For now, do not index a
string whose contents you did not construct yourself.

---

## 7. Interview questions this module answers

Answer these aloud, without notes. They are copied into `QUIZ.md`.

1. What does a slice variable contain? Name its fields and state its size on a
   64-bit machine.
2. A function receives a slice and appends to it. Can the caller read the
   appended element? Can any other code read it? State the conditions
   precisely.
3. What does `s[1:3:4]` mean, and when would you write the third number on
   purpose?
4. Two slices are made from one array with `a[:2]` and `a[2:]`. You append one
   element to the first. What happens to the second, and why does no tool
   report it?
5. Why is `&m["key"]` a compile error when `&s[0]` is legal?
6. A function returns a 60-byte slice of a 40 MB file's contents. What is the
   memory cost of holding that return value, and why?
7. Why does the runtime randomise map iteration order, and what does that
   require of your code?
8. Reading a map with `m[string(b)]` allocates nothing, but `m[string(b)]++`
   allocates on every call. Explain the difference, and describe the structure
   you would use for a counter on a hot path.

## 8. Files in this module

```
01-memory/
  NOTES.md            you are here
  MISTAKES.md         every real bug made in this module, with cause and fix
  QUIZ.md             the oral quiz tracker
  demo/
    01-header/        address, length, and capacity through every slicing operation (§1)
    02-append-growth/ measured reallocation points and growth factors (§2)
    03-aliasing/      DELIBERATELY BROKEN: the split-and-append corruption (§3)
    04-maps/          iteration order, struct values, delete, []byte-key allocation counts (§5)
    05-retain/        DELIBERATELY BROKEN: a 49-byte slice retaining 40 MB, measured (§4)
  exercises/
    README.md         the rules, the exercise table, and the definition of done
    ex1_ring.go       implement: a fixed-capacity ring buffer with zero allocations after construction
    ex2_shifts.go     fix the bug: the shift splitter that overwrites the evening shift
    ex3_compact.go    implement: an in-place filter that releases the elements it drops
    ex4_tagcount.go   implement: a per-tag counter over a byte stream with zero allocations for known tags
```

```powershell
go run ./01-memory/demo/01-header       # run any demo
go test ./01-memory/...                 # fast test loop
.\check.ps1 ./01-memory/...             # the gate: build, vet, staticcheck, test
.\check.ps1 -Bench ./01-memory/...      # the gate plus allocation counts
```
