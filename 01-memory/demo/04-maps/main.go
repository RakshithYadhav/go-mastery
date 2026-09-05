// Demo 04: four things about maps that people get wrong.
//
//	go run ./01-memory/demo/04-maps
//
// Sectioned experiment. Each section prints what the runtime actually does,
// not what the documentation says it may do. Sections 1 and 3 produce numbers
// that differ from run to run — that is the point of section 1 and not the
// point of section 3.
//
// Two of these — the struct-value read-modify-write in section 2 and the m[string(b)]
// optimisation in section 4 — are compile-time facts, so the demo shows the
// working versions and the comments name the versions that will not build.
package main

import (
	"fmt"
	"runtime"
	"testing"
)

type Order struct {
	ID  string
	Qty int
}

func main() {
	iterationOrder()
	structValues()
	deleteDoesNotShrink()
	stringKeyOptimisation()
}

func iterationOrder() {
	fmt.Println("=== 1. Iteration order is randomised on purpose ===")
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	for i := 0; i < 6; i++ {
		fmt.Print("  range ", i+1, ": ")
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}
	fmt.Println("  Same map, same process, six different orders. The runtime starts at a")
	fmt.Println("  random bucket every time so nobody can depend on an order it never")
	fmt.Println("  promised. If you need order, sort the keys yourself.")
	fmt.Println()
}

func structValues() {
	fmt.Println("=== 2. Map entries have no address, so struct values are read-only ===")
	m := map[string]Order{"WO-1": {ID: "WO-1", Qty: 1}}

	// m["WO-1"].Qty = 5   // does not compile: cannot assign to struct field
	// p := &m["WO-1"]     // does not compile: cannot take the address

	fmt.Printf("  before: %+v\n", m["WO-1"])
	o := m["WO-1"] // read a COPY
	o.Qty = 5      // modify the copy
	m["WO-1"] = o  // write it back
	fmt.Printf("  after read-modify-write: %+v\n", m["WO-1"])

	pm := map[string]*Order{"WO-1": {ID: "WO-1", Qty: 1}}
	pm["WO-1"].Qty = 5 // legal: the map value IS a pointer
	fmt.Printf("  map of pointers, assigned directly: %+v\n", *pm["WO-1"])

	fmt.Println("  Why the compile error? Growth moves entries between buckets. A pointer")
	fmt.Println("  into a map would go stale on an unrelated write. Compare &s[0] on a")
	fmt.Println("  slice, which is legal — a slice's array only moves when YOU append.")
	fmt.Println()
}

func deleteDoesNotShrink() {
	fmt.Println("=== 3. delete() never gives the memory back ===")
	var ms runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&ms)
	before := ms.HeapAlloc

	m := make(map[int]int)
	for i := 0; i < 1_000_000; i++ {
		m[i] = i
	}
	runtime.GC()
	runtime.ReadMemStats(&ms)
	full := ms.HeapAlloc

	for i := 0; i < 1_000_000; i++ {
		delete(m, i)
	}
	runtime.GC()
	runtime.ReadMemStats(&ms)
	emptied := ms.HeapAlloc

	fmt.Printf("  1,000,000 entries:   %6.1f MB above baseline\n", mb(full-before))
	fmt.Printf("  all entries deleted: %6.1f MB still held, len(m) = %d\n", mb(emptied-before), len(m))
	runtime.KeepAlive(m)
	fmt.Println("  The bucket array stays at its high-water mark for the life of the map.")
	fmt.Println("  A long-lived map used as a cache with no eviction is a leak with extra")
	fmt.Println("  steps. The fix is to build a fresh map periodically and swap it in.")
	fmt.Println()
}

func stringKeyOptimisation() {
	fmt.Println("=== 4. Reading a map with a []byte key is free. Writing one is not. ===")
	counts := map[string]int{"MACHINE-07": 1}
	ptrCounts := map[string]*int64{"MACHINE-07": new(int64)}
	key := []byte("MACHINE-07")
	var sink int

	fmt.Printf("  v := m[string(key)]              -> %.0f alloc/call\n",
		testing.AllocsPerRun(1000, func() { sink = counts[string(key)] }))
	fmt.Printf("  _, ok := m[string(key)]          -> %.0f alloc/call\n",
		testing.AllocsPerRun(1000, func() { _, ok := counts[string(key)]; _ = ok }))
	// staticcheck's SA6001 says this form "would be more efficient" written as
	// m[string(key)]. Measure it: on this Go version they are both zero. The
	// analyser is repeating the same widely held but incorrect rule this section corrects.
	// Kept deliberately, and silenced deliberately.
	//lint:ignore SA6001 the point of this line is that the analyser is wrong here
	named := testing.AllocsPerRun(1000, func() { k := string(key); sink = counts[k] })
	fmt.Printf("  k := string(key); v := m[k]      -> %.0f alloc/call\n", named)
	fmt.Printf("  m[string(key)]++                 -> %.0f alloc/call   <- the counter people write\n",
		testing.AllocsPerRun(1000, func() { counts[string(key)]++ }))
	fmt.Printf("  c := pm[string(key)]; *c++       -> %.0f alloc/call   <- the counter you should write\n",
		testing.AllocsPerRun(1000, func() {
			if c := ptrCounts[string(key)]; c != nil {
				*c++
			}
		}))
	_ = sink

	fmt.Println()
	fmt.Println("  The split is READING versus WRITING, not the shape of the expression.")
	fmt.Println("  A lookup can borrow the bytes: it hashes, compares, returns, and nobody")
	fmt.Println("  kept the string. Even naming it in a local is free, because escape")
	fmt.Println("  analysis proves it stays put. A write is different — the key might have")
	fmt.Println("  to be stored in the map forever, and the compiler cannot know in advance")
	fmt.Println("  whether it is already there. So it copies, every single call.")
	fmt.Println()
	fmt.Println("LESSON: maps trade address stability for speed. No pointers into them, no")
	fmt.Println("promised order, no shrinking. And a counter written the obvious way")
	fmt.Println("allocates once per event forever — hold pointer values and the hot path")
	fmt.Println("costs nothing. That is exercise 4.")
}

func mb(b uint64) float64 { return float64(b) / (1 << 20) }
