// Demo 03: FAILURE MODE — the split that corrupts itself.
//
//	go run ./01-memory/demo/03-aliasing
//
// This is a deliberately sick program. It splits a day's work orders into a
// morning list and an evening list, appends one rush order to the morning
// list, and silently changes the evening list. No panic. No index out of
// range. No race. go vet and staticcheck both pass it.
//
// Before you run it: the split is `morning := all[:2]` and `evening :=
// all[2:]`. Write down the len and cap of each. The bug is entirely explained
// by one of those four numbers.
//
// After you run it, find the one-line fix in NOTES.md section 3 and check that
// phase 3 below matches what you expected.
package main

import "fmt"

func main() {
	fmt.Println("=== 1. The broken split ===")
	all := []string{"WO-1", "WO-2", "WO-3", "WO-4"}
	fmt.Printf("  the day's orders: %v\n\n", all)

	morning := all[:2]
	evening := all[2:]
	fmt.Printf("  morning := all[:2]  -> %v   len=%d cap=%d   <- cap is FOUR\n",
		morning, len(morning), cap(morning))
	fmt.Printf("  evening := all[2:]  -> %v   len=%d cap=%d\n\n",
		evening, len(evening), cap(evening))
	fmt.Printf("  same array? morning's index 2 and evening's index 0 are at %p and %p\n\n",
		&morning[:3][2], &evening[0])

	fmt.Println("  a rush order comes in for the morning shift:")
	morning = append(morning, "RUSH-9")

	fmt.Printf("  morning = %v          <- correct, this is what we wanted\n", morning)
	fmt.Printf("  evening = %v   <- WRONG. Nobody wrote to evening.\n", evening)
	fmt.Printf("  all     = %v\n\n", all)
	fmt.Println("  WO-3 is gone. It was never cancelled, never rescheduled, never logged.")
	fmt.Println("  A work order left the factory's schedule because len(morning) was 2")
	fmt.Println("  and cap(morning) was 4, so append had a free slot to write into —")
	fmt.Println("  and that slot was evening[0].")
	fmt.Println()

	fmt.Println("=== 2. Why no tool catches it ===")
	fmt.Println("  Every operation here is legal and in-bounds. The write went to memory")
	fmt.Println("  this slice is genuinely allowed to touch. There is no rule being")
	fmt.Println("  broken for a checker to notice — only an assumption.")
	fmt.Println()

	fmt.Println("=== 3. The same code with the full slice expression ===")
	all2 := []string{"WO-1", "WO-2", "WO-3", "WO-4"}
	morning2 := all2[:2:2] // low : high : max  <- the fix
	evening2 := all2[2:]
	fmt.Printf("  morning2 := all2[:2:2] -> %v   len=%d cap=%d   <- cap is TWO\n",
		morning2, len(morning2), cap(morning2))

	beforeArr := &morning2[0]
	morning2 = append(morning2, "RUSH-9")
	fmt.Printf("  after append: array moved? %v\n", &morning2[0] != beforeArr)
	fmt.Printf("  morning2 = %v\n", morning2)
	fmt.Printf("  evening2 = %v   <- intact\n", evening2)
	fmt.Printf("  all2     = %v\n\n", all2)

	fmt.Println("LESSON: len is what you can see, cap is what you can WRITE THROUGH.")
	fmt.Println("When you hand out a sub-slice of something you still own, cap it with")
	fmt.Println("the three-index form or copy it. Anything else is a promise you have")
	fmt.Println("not actually made.")
}
