// Demo 02: find out where append really moves the array, and what pre-sizing saves.
//
//	go run ./01-memory/demo/02-append-growth
//
// Phase 1 appends 20,000 ints to a nil slice and prints a row every time the
// backing array actually moves, with the real growth factor. Do not memorise
// these numbers — the growth rule has changed between Go versions and the
// allocator rounds every request up to a fixed size class, which is why you
// will see caps like 6 and 10 where the doubling story predicts 4 and 8.
//
// Phase 2 runs the identical workload against make([]int, 0, 20000) and counts
// the moves and the total elements copied. Watch the second number. That is
// the cost of not saying how many you expect.
package main

import "fmt"

func main() {
	const n = 20000

	fmt.Println("=== 1. Growing from nil: every move, measured ===")
	fmt.Printf("  %-8s %-8s %-8s %-8s %s\n", "at len", "old cap", "new cap", "factor", "elements copied")

	var s []int
	moves, copied := 0, 0
	for i := 0; i < n; i++ {
		before := cap(s)
		s = append(s, i)
		if cap(s) != before {
			moves++
			copied += before
			factor := 0.0
			if before > 0 {
				factor = float64(cap(s)) / float64(before)
			}
			if moves <= 12 || cap(s) > 8000 {
				fmt.Printf("  %-8d %-8d %-8d %-8.2f %d\n", i, before, cap(s), factor, before)
			} else if moves == 13 {
				fmt.Println("  ... (middle moves omitted) ...")
			}
		}
	}
	fmt.Printf("\n  RESULT: %d elements, %d array moves, %d elements copied along the way.\n",
		n, moves, copied)
	fmt.Printf("  Final cap %d for len %d — %d slots of headroom you paid for.\n\n",
		cap(s), len(s), cap(s)-len(s))

	fmt.Println("=== 2. The same workload, pre-sized ===")
	p := make([]int, 0, n)
	pmoves, pcopied := 0, 0
	for i := 0; i < n; i++ {
		before := cap(p)
		p = append(p, i)
		if cap(p) != before {
			pmoves++
			pcopied += before
		}
	}
	fmt.Printf("  RESULT: %d elements, %d array moves, %d elements copied.\n",
		n, pmoves, pcopied)
	fmt.Printf("  Final cap %d for len %d — exactly what was asked for.\n\n", cap(p), len(p))

	fmt.Println("=== 3. Sharing stops the moment the array moves ===")
	base := make([]int, 3, 4) // len 3, cap 4: one spare slot
	base[0], base[1], base[2] = 10, 20, 30
	watcher := base[:3]

	fmt.Printf("  base and watcher both see: %v (cap %d)\n", watcher, cap(base))
	base = append(base, 40) // fits in the spare slot: SAME array
	fmt.Printf("  after append(40)  -> cap %d, array moved: %v\n", cap(base), &base[0] != &watcher[0])
	base[0] = 111
	fmt.Printf("  wrote base[0]=111 -> watcher sees %v   <- still shared\n", watcher)

	base = append(base, 50) // full: NEW array
	fmt.Printf("  after append(50)  -> cap %d, array moved: %v\n", cap(base), &base[0] != &watcher[0])
	base[0] = 222
	fmt.Printf("  wrote base[0]=222 -> watcher sees %v   <- sharing is over\n", watcher)
	fmt.Println()

	fmt.Println("LESSON: the same append call shares or stops sharing depending on")
	fmt.Println("spare capacity you cannot see at the call site. If you know the size,")
	fmt.Println("say so with make([]T, 0, n) — one allocation, zero copies.")
}
