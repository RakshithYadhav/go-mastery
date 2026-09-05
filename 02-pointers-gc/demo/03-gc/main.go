// Demo 03: what the garbage collector actually costs, and what the knobs do.
//
//	go run ./02-pointers-gc/demo/03-gc
//
// The same workload is run four times: at the default GOGC=100, at GOGC=400,
// at GOGC=off, and under a GOMEMLIMIT. Each run reports how many collections
// happened, how much CPU time went into them, the peak heap, and the total
// bytes allocated.
//
// The workload is deliberately wasteful. It allocates a 4 KB buffer per
// iteration and drops it immediately, which is what a request handler that
// builds a response buffer per request looks like from the collector's point
// of view.
//
// Predict the direction of each number before you read the table. In
// particular: does raising GOGC change the TOTAL bytes allocated?
//
// For the collector's own running commentary, run it again with tracing on:
//
//	$env:GODEBUG='gctrace=1'; go run ./02-pointers-gc/demo/03-gc; $env:GODEBUG=''
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	iterations = 400_000
	bufSize    = 4096
)

var sink []byte

type result struct {
	label      string
	numGC      uint32
	pauseTotal float64 // milliseconds
	elapsedMs  float64
	peakHeapMB float64
	allocedMB  float64
}

func main() {
	defer func() { _ = sinksUsed() }()

	fmt.Println("Same workload, four garbage collector settings.")
	fmt.Printf("Workload: %d iterations, each allocating a %d-byte buffer and dropping it.\n\n",
		iterations, bufSize)

	// Keep the machine's core count out of the comparison.
	runtime.GOMAXPROCS(runtime.NumCPU())

	results := []result{
		run("GOGC=100 (default)", func() { debug.SetGCPercent(100); debug.SetMemoryLimit(mathMaxInt64) }),
		run("GOGC=400", func() { debug.SetGCPercent(400); debug.SetMemoryLimit(mathMaxInt64) }),
		run("GOGC=off", func() { debug.SetGCPercent(-1); debug.SetMemoryLimit(mathMaxInt64) }),
		run("GOGC=off + GOMEMLIMIT=64MiB", func() { debug.SetGCPercent(-1); debug.SetMemoryLimit(64 << 20) }),
	}

	// Restore defaults so nothing below is measured under the last setting.
	debug.SetGCPercent(100)
	debug.SetMemoryLimit(mathMaxInt64)

	fmt.Printf("  %-28s %8s %10s %11s %10s %10s\n",
		"setting", "GCs", "pause ms", "elapsed", "peak heap", "allocated")
	fmt.Println("  " + dashes(80))
	for _, r := range results {
		fmt.Printf("  %-28s %8d %10.2f %8.0f ms %7.1f MB %7.0f MB\n",
			r.label, r.numGC, r.pauseTotal, r.elapsedMs, r.peakHeapMB, r.allocedMB)
	}

	fmt.Println()
	fmt.Println("Read the last column first: it barely moves. The knobs do not change how")
	fmt.Println("much your program allocates. They only change how often the collector")
	fmt.Println("cleans up after it, which trades memory for CPU in one direction and CPU")
	fmt.Println("for memory in the other.")
	fmt.Println()
	fmt.Println("GOGC=100 means: start the next cycle when the heap has grown to twice the")
	fmt.Println("live heap from the last one. GOGC=400 means five times, so there are far")
	fmt.Println("fewer cycles and a much larger peak heap.")
	fmt.Println()
	fmt.Println("Now look at the GOGC=off row, which is the interesting one. It does no")
	fmt.Println("collection at all, and on most runs it is the SLOWEST of the four despite")
	fmt.Println("doing none of the work the other three are paying for.")
	fmt.Println()
	fmt.Println("Turning the collector off does not make the memory free. It makes the")
	fmt.Println("program ask the operating system for roughly 1.5 GB of fresh pages instead")
	fmt.Println("of reusing a few megabytes over and over. Every new page has to be mapped")
	fmt.Println("and faulted in, and none of it is in cache. Reusing a small hot heap beats")
	fmt.Println("touching a large cold one, and that is why the default is not 'off'.")
	fmt.Println()
	fmt.Println("The fourth row is the production answer when memory is capped. GOMEMLIMIT")
	fmt.Println("is a soft ceiling that makes the collector work as hard as it must to stay")
	fmt.Println("under the limit. GOGC off plus GOMEMLIMIT set is a real configuration for")
	fmt.Println("a service in a container with a fixed memory allocation: collect only when")
	fmt.Println("approaching the ceiling, rather than on a ratio that ignores it.")
	fmt.Println()
	fmt.Println("The fix that actually helps is none of these. It is allocating less, which")
	fmt.Println("moves the last column. That is what the exercises are about.")
}

// mathMaxInt64 avoids importing math for one constant.
const mathMaxInt64 = 1<<63 - 1

func run(label string, configure func()) result {
	configure()

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	peak := uint64(0)
	start := time.Now()
	for i := 0; i < iterations; i++ {
		buf := make([]byte, bufSize)
		buf[0] = byte(i)
		sink = buf

		if i%500 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.HeapAlloc > peak {
				peak = m.HeapAlloc
			}
		}
	}

	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)

	return result{
		label:      label,
		numGC:      after.NumGC - before.NumGC,
		pauseTotal: float64(after.PauseTotalNs-before.PauseTotalNs) / 1e6,
		elapsedMs:  float64(elapsed.Milliseconds()),
		peakHeapMB: float64(peak) / (1 << 20),
		allocedMB:  float64(after.TotalAlloc-before.TotalAlloc) / (1 << 20),
	}
}

func dashes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}

// sinksUsed reads every package-level sink. Assigning to those variables is what
// stops the COMPILER deleting work that nothing observes; reading them here is
// what stops the LINTER reporting them as unused. Two different tools, two
// different definitions of "used".
func sinksUsed() bool {
	return sink != nil
}
