// Demo 05: FAILURE MODE — sixty bytes that keep forty megabytes alive.
//
//	go run ./01-memory/demo/05-retain
//
// A hunting exercise. This program reads a "file" (a 40 MB buffer), keeps only
// the first line out of it, drops every other reference, forces a garbage
// collection, and then asks the runtime how much heap is still in use.
//
// Predict the number before you run it. Then predict it again for phase 2,
// which does exactly the same thing with one line changed.
//
// The symbol you should end up pointing at is the return statement in
// keepFirstLine. Nothing else in this program is unusual.
package main

import (
	"bytes"
	"fmt"
	"runtime"
)

const fileSize = 40 << 20 // 40 MB

// loadFile stands in for os.ReadFile on a large CSV export.
func loadFile() []byte {
	buf := make([]byte, fileSize)
	copy(buf, []byte("WO-1001,LINE-A,2026-09-04T06:00:00Z,COMPLETED,412\n"))
	for i := 50; i < fileSize; i++ {
		buf[i] = 'x'
	}
	return buf
}

// keepFirstLine returns the header line of the file. It looks free.
func keepFirstLine(contents []byte) []byte {
	i := bytes.IndexByte(contents, '\n')
	return contents[:i]
}

// keepFirstLineCopied returns the same bytes, owned outright.
func keepFirstLineCopied(contents []byte) []byte {
	i := bytes.IndexByte(contents, '\n')
	out := make([]byte, i)
	copy(out, contents[:i])
	return out
}

func heapMB() float64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / (1 << 20)
}

func main() {
	fmt.Printf("  baseline heap in use: %.1f MB\n\n", heapMB())

	fmt.Println("=== 1. Keeping a sub-slice of the file ===")
	kept := keepFirstLine(loadFile())
	fmt.Printf("  kept %d bytes: %q\n", len(kept), kept)
	fmt.Printf("  cap(kept) = %d   <- the capacity runs to the end of the 40 MB array\n", cap(kept))
	fmt.Printf("  heap in use after dropping every other reference: %.1f MB\n", heapMB())
	fmt.Println("  The 40 MB array cannot be collected. One slice header still points")
	fmt.Println("  into it, and the garbage collector frees arrays whole or not at all.")
	runtime.KeepAlive(kept)
	kept = nil
	fmt.Printf("  after kept = nil:  %.1f MB   <- now it goes\n\n", heapMB())

	fmt.Println("=== 2. The same function with the bytes copied out ===")
	kept2 := keepFirstLineCopied(loadFile())
	fmt.Printf("  kept %d bytes: %q\n", len(kept2), kept2)
	fmt.Printf("  cap(kept2) = %d   <- it owns exactly what it can see\n", cap(kept2))
	fmt.Printf("  heap in use after dropping every other reference: %.1f MB\n", heapMB())
	runtime.KeepAlive(kept2)
	fmt.Println()

	fmt.Println("LESSON: a slice keeps its whole backing array alive, not the part you")
	fmt.Println("can see. Returning a small slice of a large buffer is a leak whenever")
	fmt.Println("the caller outlives the buffer — which in a server is always. Copy on")
	fmt.Println("the way out: make+copy, slices.Clone, or strings.Clone.")
}
