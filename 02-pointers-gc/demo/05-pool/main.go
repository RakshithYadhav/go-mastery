// Demo 05: sync.Pool measured, emptied, and misused.
//
//	go run ./02-pointers-gc/demo/05-pool
//
// Four experiments:
//
//  1. The same encoding workload with and without a pool, allocations measured.
//  2. What a garbage collection does to the contents of a pool.
//  3. The bug that makes pooling dangerous rather than merely useless: an
//     object handed out without being reset.
//  4. The bug that makes a pool a memory leak: one oversized object retained
//     forever because nothing checks capacity on the way back in.
//
// Read the sync.Pool doc comment before this demo. The sentence that matters
// is the one saying any item may be removed automatically at any time without
// notification. Experiment 2 is that sentence, measured.
package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

type Frame struct {
	Machine string
	Celsius float64
	Seq     int
}

var sinkBytes []byte

// encodeNoPool allocates a fresh buffer for every frame.
func encodeNoPool(f Frame) []byte {
	buf := make([]byte, 0, 256)
	return appendFrame(buf, f)
}

var framePool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 256)
		return &buf
	},
}

// encodeWithPool borrows a buffer, uses it, and returns it.
func encodeWithPool(f Frame, out []byte) []byte {
	p := framePool.Get().(*[]byte)
	buf := (*p)[:0] // reset the LENGTH; the capacity is what we borrowed it for
	buf = appendFrame(buf, f)
	out = append(out[:0], buf...)
	*p = buf
	framePool.Put(p)
	return out
}

func appendFrame(dst []byte, f Frame) []byte {
	dst = append(dst, f.Machine...)
	dst = append(dst, ',')
	dst = strconv.AppendFloat(dst, f.Celsius, 'f', 1, 64)
	dst = append(dst, ',')
	dst = strconv.AppendInt(dst, int64(f.Seq), 10)
	return dst
}

func main() {
	defer func() { _ = sinksUsed() }()

	poolVsNoPool()
	gcEmptiesThePool()
	theResetBug()
	theOversizedBug()
}

func poolVsNoPool() {
	fmt.Println("=== 1. The same workload, with and without a pool ===")
	f := Frame{Machine: "MACHINE-07", Celsius: 214.5, Seq: 9001}
	out := make([]byte, 0, 256)

	noPool := testing.AllocsPerRun(1000, func() { sinkBytes = encodeNoPool(f) })
	withPool := testing.AllocsPerRun(1000, func() { out = encodeWithPool(f, out) })

	fmt.Printf("  encodeNoPool    %.0f allocations per frame\n", noPool)
	fmt.Printf("  encodeWithPool  %.0f allocations per frame\n", withPool)
	fmt.Println()
	fmt.Println("  At 40,000 frames a second the first row is 40,000 buffers a second for")
	fmt.Println("  the collector to find and free. The second row is the same buffer being")
	fmt.Println("  handed back and forth.")
	fmt.Println()
	fmt.Println("  This is also the honest caveat: the pool version is longer, harder to")
	fmt.Println("  read, and has two bugs available to it that the first version cannot")
	fmt.Println("  have. Experiments 3 and 4 are those bugs. Reach for a pool when you have")
	fmt.Println("  measured that you need it, not before.")
	fmt.Println()
}

func gcEmptiesThePool() {
	fmt.Println("=== 2. A garbage collection empties the pool ===")

	var p sync.Pool
	for i := 0; i < 10; i++ {
		b := make([]byte, 0, 256)
		p.Put(&b)
	}
	fmt.Println("  put 10 buffers into the pool")

	got := 0
	for i := 0; i < 10; i++ {
		if p.Get() != nil {
			got++
		}
	}
	fmt.Printf("  got %d of them straight back\n", got)

	for i := 0; i < 10; i++ {
		b := make([]byte, 0, 256)
		p.Put(&b)
	}
	fmt.Println("  put 10 more in, then ran two garbage collections")
	runtime.GC()
	runtime.GC()

	got = 0
	for i := 0; i < 10; i++ {
		if p.Get() != nil {
			got++
		}
	}
	fmt.Printf("  got %d of them back\n\n", got)

	fmt.Println("  Pooled objects survive one collection in the victim cache and are")
	fmt.Println("  discarded at the second. A pool is a cache with no guarantees, so it")
	fmt.Println("  cannot be used for anything you must get back: not connections, not")
	fmt.Println("  anything holding a file handle, not a fixed-size resource limit.")
	fmt.Println()
}

func theResetBug() {
	fmt.Println("=== 3. The reset bug: one machine's data in another machine's frame ===")

	var p sync.Pool
	p.New = func() any {
		buf := make([]byte, 0, 256)
		return &buf
	}

	// First user writes a frame and returns the buffer WITHOUT resetting it.
	first := p.Get().(*[]byte)
	*first = appendFrame(*first, Frame{Machine: "MACHINE-07", Celsius: 214.5, Seq: 1})
	fmt.Printf("  first user wrote:  %q\n", string(*first))
	p.Put(first)

	// Second user borrows it and appends, assuming it is empty.
	second := p.Get().(*[]byte)
	*second = appendFrame(*second, Frame{Machine: "PRESS-2", Celsius: 88.0, Seq: 2})
	fmt.Printf("  second user wrote: %q\n", string(*second))
	fmt.Println()
	fmt.Println("  The second frame contains the first machine's reading. Nothing failed,")
	fmt.Println("  nothing panicked, and the dashboard now shows a temperature that was")
	fmt.Println("  never recorded for that machine. In a service handling user data this")
	fmt.Println("  is how one customer's information ends up in another's response.")
	fmt.Println()
	fmt.Println("  The fix is one line, and it belongs on the GET side, not the Put side:")
	fmt.Println("  buf := (*p)[:0]. Resetting on Get is safer than resetting on Put,")
	fmt.Println("  because a Put can be skipped by an early return or a panic.")
	fmt.Println()
}

func theOversizedBug() {
	fmt.Println("=== 4. The oversized bug: a pool that never gives memory back ===")

	var p sync.Pool
	p.New = func() any {
		buf := make([]byte, 0, 256)
		return &buf
	}

	// The normal case: small frames.
	b := p.Get().(*[]byte)
	*b = append((*b)[:0], make([]byte, 200)...)
	p.Put(b)
	b2 := p.Get().(*[]byte)
	fmt.Printf("  after normal use, pooled buffer capacity: %d bytes\n", cap(*b2))
	p.Put(b2)

	// Once a day, one enormous batch arrives.
	big := p.Get().(*[]byte)
	*big = append((*big)[:0], make([]byte, 40<<20)...)
	p.Put(big)

	after := p.Get().(*[]byte)
	fmt.Printf("  after ONE 40 MB batch, pooled buffer capacity: %d bytes (%.1f MB)\n",
		cap(*after), float64(cap(*after))/(1<<20))
	runtime.KeepAlive(after)
	fmt.Println()
	fmt.Println("  Every small frame from now on borrows a 40 MB buffer, and the pool holds")
	fmt.Println("  it alive between uses. One outlier permanently changed the service's")
	fmt.Println("  memory profile.")
	fmt.Println()
	fmt.Println("  The fix is a capacity guard on Put: if cap(buf) is above some threshold,")
	fmt.Println("  do not return it to the pool at all. Let the collector have it. That")
	fmt.Println("  guard is part of exercise 3.")
}

// sinksUsed reads every package-level sink. Assigning to those variables is what
// stops the COMPILER deleting work that nothing observes; reading them here is
// what stops the LINTER reporting them as unused. Two different tools, two
// different definitions of "used".
func sinksUsed() bool {
	return sinkBytes != nil
}
