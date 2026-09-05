// Demo 04: five functions that allocate, and none of them look like it.
//
//	go run ./02-pointers-gc/demo/04-escape-traps
//
// Every function below is written the way it would be written in a real
// service. None of them contains the word `new`, none returns a pointer to a
// local, and every one of them allocates on every call.
//
// This is a prediction exercise. For each of the five, write down your answer
// before you run it: how many allocations per call, and which line causes them?
//
// Then find each one in the compiler's output:
//
//	go build -gcflags='-m -l' ./02-pointers-gc/demo/04-escape-traps
package main

import (
	"errors"
	"fmt"
	"testing"
)

type Reading struct {
	Machine string
	Celsius float64
}

var (
	sinkStr  string
	sinkErr  error
	sinkAny  any
	sinkBool bool
	sinkInt  int
)

var errTooHot = errors.New("temperature above limit")

// TRAP 1: a log line built the obvious way, on a path that runs per event.
func logReading(r Reading) string {
	return fmt.Sprintf("machine=%s temp=%.1f", r.Machine, r.Celsius)
}

// TRAP 2: an error created at the point of failure, with no formatting.
func checkTemp(r Reading) error {
	if r.Celsius > 200 {
		return errors.New("temperature above limit")
	}
	return nil
}

// TRAP 2b: the same check, with the error created once at package level.
func checkTempSentinel(r Reading) error {
	if r.Celsius > 200 {
		return errTooHot
	}
	return nil
}

// TRAP 3: collecting values through an interface, the way a logger or a
// reporting layer does. The conversion is the cost, not the method call.
//
// Note that this only allocates because the interface value OUTLIVES the
// function, via the package-level slice. A `var s Stringer = r` that stays
// inside one function does not allocate, because the compiler can put the
// boxed value in the frame. Escape analysis applies to boxing like everything
// else.
type Stringer interface{ String() string }

func (r Reading) String() string { return r.Machine }

var collected []Stringer

func collectViaInterface(r Reading) {
	collected = append(collected[:0], r)
}

// TRAP 4: joining two strings in the ordinary way.
func joinTags(a, b string) string {
	return a + ":" + b
}

// TRAP 4b: the same join, into a caller-supplied buffer.
func joinTagsBuf(dst []byte, a, b string) []byte {
	dst = append(dst, a...)
	dst = append(dst, ':')
	dst = append(dst, b...)
	return dst
}

// TRAP 5: a lookup table built inside the function, used once, dropped.
// A very small map that does not escape can be built in the frame; this one is
// past that threshold, which is what makes it the realistic case.
func isAlarm(code string) bool {
	alarms := map[string]bool{
		"E01": true, "E02": true, "E07": true, "E11": true,
		"E12": true, "E19": true, "E23": true, "E31": true,
		"E44": true, "E52": true, "E63": true, "E70": true,
	}
	return alarms[code]
}

// TRAP 5b: the same table, built once.
var alarmCodes = map[string]bool{
	"E01": true, "E02": true, "E07": true, "E11": true,
	"E12": true, "E19": true, "E23": true, "E31": true,
	"E44": true, "E52": true, "E63": true, "E70": true,
}

func isAlarmShared(code string) bool { return alarmCodes[code] }

func main() {
	defer func() { _ = sinksUsed() }()

	r := Reading{Machine: "MACHINE-07", Celsius: 214.5}
	buf := make([]byte, 0, 64)

	fmt.Println("Allocations per call. Predict each one before you read it.")
	fmt.Println()
	fmt.Printf("  %-26s %-6s %s\n", "function", "allocs", "what actually allocates")
	fmt.Println("  " + dashes(78))

	show("logReading", func() { sinkStr = logReading(r) },
		"Sprintf boxes both arguments into ...any, then builds a new string")
	show("checkTemp", func() { sinkErr = checkTemp(r) },
		"errors.New allocates a struct on every failure, and it escapes as an interface")
	show("checkTempSentinel", func() { sinkErr = checkTempSentinel(r) },
		"the same error value, created once at startup")
	show("collectViaInterface", func() { collectViaInterface(r) },
		"an interface holds a pointer, so a struct stored in one needs an address")
	show("joinTags", func() { sinkStr = joinTags("line", "A") },
		"strings are immutable, so a join is always a new allocation")
	show("joinTagsBuf", func() { buf = joinTagsBuf(buf[:0], "line", "A") },
		"writes into memory the caller already owns")
	show("isAlarm", func() { sinkBool = isAlarm("E07") },
		"a fresh map per call: the map header, its buckets, and every key")
	show("isAlarmShared", func() { sinkBool = isAlarmShared("E07") },
		"the same table, built once at startup")

	fmt.Println()
	fmt.Println("The pattern in every pair above is the same, and it is not a Go trick.")
	fmt.Println("The allocating version does work per call that only needed doing once, or")
	fmt.Println("produces a new value where the caller could have supplied the memory.")
	fmt.Println()
	fmt.Println("None of these matter at 100 calls a second. All of them matter at 40,000.")
	fmt.Println("That is the entire content of war-story scenario 1 on this track: the")
	fmt.Println("telemetry collector does exactly this, several times per frame.")
}

func show(name string, f func(), why string) {
	n := testing.AllocsPerRun(1000, f)
	fmt.Printf("  %-26s %-6.0f %s\n", name, n, why)
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
	return sinkStr != "" || sinkErr != nil || sinkAny != nil || sinkBool || sinkInt != 0
}
