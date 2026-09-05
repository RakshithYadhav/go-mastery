// Demo 02: watch a goroutine's stack move underneath a live pointer.
//
//	go run ./02-pointers-gc/demo/02-stack
//
// A goroutine starts with a small stack. When a call would overflow it, the
// runtime allocates a bigger stack, copies every live frame into it, and then
// rewrites every pointer that referred into the old stack so that it refers
// into the new one.
//
// This program takes the address of one variable in an early frame, recurses
// deep enough with large frames to force several of those moves, and prints the
// address of that same variable as it goes down.
//
// Predict before running: does the address stay the same?
//
// The address is printed as a number rather than passed to Printf as a pointer.
// Passing the pointer itself to a ...any function would make the compiler move
// the variable to the heap, and a heap address never changes — the demo would
// silently prove nothing. Confirm that for yourself with:
//
//	go build -gcflags='-m -l' ./02-pointers-gc/demo/02-stack
//
// and check that "x escapes to heap" does NOT appear.
package main

import (
	"fmt"
	"unsafe"
)

const (
	maxDepth  = 96
	frameSize = 2048 // bytes of padding per frame, to force growth quickly
	reportPer = 8
)

func main() {
	fmt.Println("Recursing on a fresh goroutine stack, holding a pointer to one")
	fmt.Printf("variable in the first frame. Each frame adds %d bytes.\n\n", frameSize)

	done := make(chan struct{})
	go func() {
		defer close(done)
		observe()
	}()
	<-done
}

func observe() {
	x := 42
	addr := uintptr(unsafe.Pointer(&x))

	fmt.Printf("  %-8s %-20s %s\n", "depth", "address of x", "what happened")
	fmt.Println("  " + dashes(64))
	fmt.Printf("  %-8d 0x%-18x starting frame\n", 0, addr)

	recurse(&x, 1, addr)

	fmt.Println()
	fmt.Println("Every change of address is one stack growth: a larger stack was")
	fmt.Println("allocated, all live frames were copied into it, and every pointer into")
	fmt.Println("the old stack — including the one this function is still holding — was")
	fmt.Println("rewritten. The value of x was never wrong at any point.")
	fmt.Println()
	fmt.Println("This is why Go can move stacks and C cannot: the runtime knows exactly")
	fmt.Println("which words on the stack are pointers, so it can find and fix them all.")
}

func recurse(p *int, depth int, prev uintptr) {
	var pad [frameSize]byte
	pad[0] = byte(depth)

	cur := uintptr(unsafe.Pointer(p))

	switch {
	case cur != prev:
		fmt.Printf("  %-8d 0x%-18x STACK MOVED (was 0x%x)\n", depth, cur, prev)
	case depth%reportPer == 0:
		fmt.Printf("  %-8d 0x%-18x unchanged\n", depth, cur)
	}

	if depth < maxDepth {
		recurse(p, depth+1, cur)
	}

	if pad[0] != byte(depth) {
		panic("frame corrupted")
	}
}

func dashes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}
