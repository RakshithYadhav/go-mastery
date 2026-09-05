package exercises

// Exercise 2 — FIX THE BUG: a hot path that allocates, and the compiler will
// tell you why if you ask it.
//
// ShiftStats accumulates running statistics over a shift's telemetry frames.
// Observe is called once per frame, 40,000 times a second. It is correct — its
// numbers are right and the tests for its output pass. It also allocates on
// every single call, and the CPU profile for this service says 60% of its time
// is garbage collection.
//
// Nothing in this file returns a pointer to a local. Nothing here looks like an
// allocation. Every one of them comes from a cause in NOTES section 4 or from
// Module 1 section 5.
//
// YOUR TASKS
//
//  1. Find them with evidence, not by guessing. Run:
//
//         go build -gcflags='-m -l' ./02-pointers-gc/exercises
//
//     and read the lines that name this file. There are THREE separate causes,
//     on three different lines, and they are different kinds of problem. Write
//     down which line each compiler message refers to, and what it is telling
//     you, before you change any code.
//
//     One of the three will not appear in that output at all. Measure as well
//     as read — `.\check.ps1 -Bench ./02-pointers-gc/...` reports allocs/op.
//     Working out which one the compiler stayed quiet about, and why, is the
//     most valuable part of this exercise.
//
//  2. Fix all three so TestStats_ObserveDoesNotAllocate passes, while
//     TestStats_StillCorrect and TestStats_SummaryUnchanged keep passing.
//     The observable behaviour must not change: same numbers, same rounding,
//     same summary text, same per-machine counts.
//
// THE CONSTRAINTS
//
//   - The exported API does not change. Observe takes the same arguments and
//     returns the same bool. Summary returns the same string.
//   - Summary runs ONCE at the end of a shift. It is allowed to allocate. Do
//     not contort it to save allocations that do not matter.
//   - Observe is the hot path and must reach exactly zero.
//   - You may add, remove or retype unexported fields freely. That is where
//     all three fixes live.
//
// THINK BEFORE YOU TYPE
//
//	The shape:     for each of the three, ask what work is being done per frame
//	               that only needs doing once per shift — or not at all until
//	               someone asks for the report.
//	The ownership: `recent` stores values that came from the caller. After you
//	               retype it, who owns those values, and does anything outlive
//	               the call?
//	The exits:     a machine seen for the first time, a shift with zero frames,
//	               and the very first frame, where Max is still negative
//	               infinity.
//
// STRETCH GOAL: once all three are fixed, look at ns/op as well as allocs/op.
// One fix improved speed far more than the other two. Work out which, and be
// able to say why.
//
// Do not change the exported signatures.

import (
	"fmt"
	"math"
	"strconv"
)

// ShiftStats accumulates statistics over one shift on the shop floor.
type ShiftStats struct {
	Count  int
	Sum    float64
	Max    float64
	Alarms int

	// recent holds the most recent event, for the end-of-shift report.
	recent []any

	// hottest is a rendered "machine@temperature" label for the report.
	hottest string

	// perMachine counts frames per machine, keyed by machine and sensor bank.
	perMachine map[string]int

	// scratch is a reusable buffer for building the perMachine key.
	scratch []byte
}

// NewShiftStats returns an empty accumulator.
func NewShiftStats() *ShiftStats {
	return &ShiftStats{
		Max:        math.Inf(-1),
		recent:     make([]any, 0, 3),
		perMachine: make(map[string]int, 64),
		scratch:    make([]byte, 0, 64),
	}
}

// Observe records one frame and reports whether it was an alarm.
// THIS IS THE HOT PATH: 40,000 calls a second.
func (s *ShiftStats) Observe(machine string, celsius float64, seq int) bool {
	s.Count++
	s.Sum += celsius

	if celsius > s.Max {
		s.Max = celsius
		s.hottest = fmt.Sprintf("%s@%.1f", machine, celsius)
	}

	alarm := celsius > 200
	if alarm {
		s.Alarms++
	}

	s.recent = append(s.recent[:0], machine, celsius, seq)

	s.scratch = append(s.scratch[:0], machine...)
	s.scratch = append(s.scratch, '#')
	s.scratch = strconv.AppendInt(s.scratch, int64(seq%4), 10)
	s.perMachine[string(s.scratch)]++

	return alarm
}

// Summary renders the end-of-shift report. Called ONCE per shift.
// This function is allowed to allocate.
func (s *ShiftStats) Summary() string {
	if s.Count == 0 {
		return "no frames"
	}
	mean := s.Sum / float64(s.Count)
	return fmt.Sprintf("frames=%d mean=%.2f max=%.1f alarms=%d hottest=%s",
		s.Count, mean, s.Max, s.Alarms, s.hottest)
}

// MachineCount returns how many frames were seen for one machine and bank.
func (s *ShiftStats) MachineCount(machine string, bank int) int {
	return s.perMachine[machine+"#"+strconv.Itoa(bank)]
}
