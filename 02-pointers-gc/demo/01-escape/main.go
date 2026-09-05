// Demo 01: eight functions, and where the compiler decided to put their values.
//
//	go run ./02-pointers-gc/demo/01-escape
//	go build -gcflags='-m -l' ./02-pointers-gc/demo/01-escape
//
// Run both commands. The first prints the measured number of allocations per
// call for each function below. The second prints the compiler's own reasoning
// for the same functions. Match every line of the second output to a row of the
// first; they are two views of one decision.
//
// The -l flag disables inlining. Without it the compiler merges these small
// functions into their callers and reports on code you did not write.
//
// Predict all eight numbers before you run anything. Most people get rows 4, 7
// and 8 wrong.
package main

import (
	"fmt"
	"testing"
)

type Order struct {
	ID    string
	Qty   int
	Line  string
	Shift int
}

// Package-level sinks. Assigning results here stops the compiler from deleting
// work that nothing observes, which would make every measurement zero.
var (
	sinkInt  int
	sinkStr  string
	sinkAny  any
	sinkPtr  *Order
	sinkFunc func() int
)

// Two values the compiler cannot fold into constants. Without these, it proves
// the interesting cases away and rows 4 and 8 both measure zero: a constant put
// into an interface is served from read-only data, and make([]byte, 64) with a
// literal 64 is provably bounded even when the parameter looks variable.
var (
	counter  = 1 << 20
	sliceLen = 64
)

// 1. A value created and used entirely inside one function.
func localOnly() int {
	o := Order{ID: "WO-1", Qty: 5}
	return o.Qty
}

// 2. A pointer to a local, returned to the caller.
func returnsPointer() *Order {
	o := Order{ID: "WO-1", Qty: 5}
	return &o
}

// 3. A pointer passed down into another function, not kept.
func sharedDown() int {
	o := Order{ID: "WO-1", Qty: 5}
	return readQty(&o)
}

func readQty(o *Order) int { return o.Qty }

// 4. An int stored in an interface. The value is 8 bytes and it still allocates.
func intoInterface() {
	counter++
	sinkAny = counter
}

// 5. A string built with fmt.
func withSprintf() string {
	o := Order{ID: "WO-1", Qty: 5}
	return fmt.Sprintf("%s:%d", o.ID, o.Qty)
}

// 6. A closure that captures a local and outlives the call.
func returnsClosure() func() int {
	n := 42
	return func() int { return n }
}

// 7. A slice whose size is a compile-time constant.
func sliceConstantSize() int {
	buf := make([]byte, 64)
	buf[0] = 'x'
	return len(buf)
}

// 8. The same slice, sized by a variable the compiler cannot bound.
func sliceVariableSize(n int) int {
	buf := make([]byte, n)
	buf[0] = 'x'
	return len(buf)
}

func main() {
	defer func() { _ = sinksUsed() }()

	fmt.Println("Allocations per call, measured. Compare with go build -gcflags='-m -l'.")
	fmt.Println()
	fmt.Printf("  %-28s %-12s %s\n", "function", "allocs/call", "why")
	fmt.Println("  " + line(76))

	report("localOnly", func() { sinkInt = localOnly() },
		"never leaves the frame; lives on the stack")
	report("returnsPointer", func() { sinkPtr = returnsPointer() },
		"the caller still holds it after the frame is gone")
	report("sharedDown", func() { sinkInt = sharedDown() },
		"shared DOWN the call stack; the callee cannot outlive the caller")
	report("intoInterface", func() { intoInterface() },
		"an interface holds a pointer, so the int needs an address")
	report("withSprintf", func() { sinkStr = withSprintf() },
		"variadic ...any boxes both arguments, and the result is a new string")
	report("returnsClosure", func() { sinkFunc = returnsClosure() },
		"the captured variable outlives the call, so it cannot be a stack slot")
	report("sliceConstantSize", func() { sinkInt = sliceConstantSize() },
		"64 bytes, known at compile time, does not escape: stack")
	report("sliceVariableSize", func() { sinkInt = sliceVariableSize(sliceLen) },
		"same 64 bytes, but the compiler cannot prove a bound: heap")

	fmt.Println()
	fmt.Println("Rows 7 and 8 allocate the same number of bytes and differ only in whether")
	fmt.Println("the compiler can see the size. Escape analysis is about what the compiler")
	fmt.Println("can PROVE, not about what is true.")
	fmt.Println()
	fmt.Println("Now run:  go build -gcflags='-m -l' ./02-pointers-gc/demo/01-escape")
	fmt.Println("and find the compiler's line for each row above.")
	fmt.Println()
	fmt.Println("ONE ROW WILL NOT MATCH, AND IT IS THE MOST USEFUL ROW HERE.")
	fmt.Println("For sliceVariableSize the compiler prints:")
	fmt.Println()
	fmt.Println("    make([]byte, n) does not escape")
	fmt.Println()
	fmt.Println("and the measurement above says it allocates once per call. Both are true.")
	fmt.Println("\"Does not escape\" means no reference outlives the function. It does NOT")
	fmt.Println("mean the value went on the stack. A stack frame's size is fixed when the")
	fmt.Println("function is compiled, so an allocation whose size is only known at run")
	fmt.Println("time cannot live in the frame no matter how short its life is. It goes on")
	fmt.Println("the heap and is collected normally.")
	fmt.Println()
	fmt.Println("So there are two separate questions, and -gcflags=-m only answers the first:")
	fmt.Println("  1. Does anything outlive this function?   -> escape analysis")
	fmt.Println("  2. Can this fit in a fixed-size frame?    -> size known at compile time")
	fmt.Println("A no to either one sends the value to the heap. Measure; do not infer.")
}

func report(name string, f func(), why string) {
	n := testing.AllocsPerRun(1000, f)
	fmt.Printf("  %-28s %-12.0f %s\n", name, n, why)
}

func line(n int) string {
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
	return sinkInt != 0 || sinkStr != "" || sinkAny != nil || sinkPtr != nil || sinkFunc != nil
}
