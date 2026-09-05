// Demo 01: a slice variable is 24 bytes that describe a region of an array stored elsewhere.
//
//	go run ./01-memory/demo/01-header
//
// This walks the exact table in NOTES.md section 1, printing the three fields
// of the slice header at every step: where it points, how much it can see
// (len), and how much room exists from that point to the end of the array
// (cap). Addresses are printed as "+N bytes from the start of base" so you can
// see re-basing happen instead of squinting at hex.
//
// Before you run it, predict two things and write them down: the cap of
// base[2:5], and the cap of that slice sliced again with [1:2]. Most people
// get the second one wrong.
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	base := [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	origin := uintptr(unsafe.Pointer(&base[0]))

	// describe prints one slice's header relative to the start of base.
	describe := func(label string, s []int) {
		off := uintptr(unsafe.Pointer(&s[:1][0])) - origin
		fmt.Printf("  %-18s ptr=base+%2d bytes (index %d)   len=%d   cap=%d   sees %v\n",
			label, off, off/unsafe.Sizeof(int(0)), len(s), cap(s), s)
	}

	fmt.Println("=== 1. The array holds the elements. The slice describes a region of it. ===")
	fmt.Printf("  base is [8]int, %d bytes of actual integers at %p\n\n",
		unsafe.Sizeof(base), &base[0])

	s := base[2:5]
	describe("s := base[2:5]", s)
	fmt.Println("     cap is 6, not 3: from index 2 there are 6 elements before the array ends.")
	fmt.Println()

	s2 := s[1:2]
	describe("s2 := s[1:2]", s2)
	fmt.Println("     Slicing a slice RE-BASES. s2 starts at base index 3, so its cap is 5.")
	fmt.Println("     s2's index 0 and s's index 1 are the same memory.")
	fmt.Println()

	s3 := s[:cap(s)]
	describe("s3 := s[:cap(s)]", s3)
	fmt.Println("     s could 'only see' 3 elements. Re-slicing to its cap legally hands")
	fmt.Println("     back 6 — including two that s was never given. len limits what a slice")
	fmt.Println("     may read; it is not a boundary on what exists in the array.")
	fmt.Println()

	s4 := base[2:5:6]
	describe("s4 := base[2:5:6]", s4)
	fmt.Println("     The FULL SLICE EXPRESSION low:high:max. cap = 6-2 = 4, so s4 can")
	fmt.Println("     never reach base[6] or base[7]. This is the fix in section 3.")
	fmt.Println()

	fmt.Println("=== 2. Copying a slice variable shares the array ===")
	t := s
	fmt.Printf("  before: s = %v   t = %v\n", s, t)
	t[0] = 999
	fmt.Printf("  after t[0] = 999:\n")
	fmt.Printf("          s = %v   t = %v   base = %v\n", s, t, base)
	fmt.Println("     One write, three names changed. t := s copied 24 bytes and zero data.")
	fmt.Println()

	fmt.Println("=== 3. The same syntax on an ARRAY copies everything ===")
	arr := [5]int{10, 20, 30, 40, 50}
	cp := arr
	cp[0] = 999
	fmt.Printf("  arr = %v   cp = %v\n", arr, cp)
	fmt.Println("     cp := arr copied five integers. arr is untouched.")
	fmt.Println()

	fmt.Println("LESSON: assignment copies the header, never the array. Whether two")
	fmt.Println("variables share data is decided by the TYPE, and nothing at the")
	fmt.Println("assignment site tells you which one you have.")
}
