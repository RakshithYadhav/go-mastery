package exercises

// Exercise 2 — FIX THE BUG: the shift splitter that rewrites its own schedule.
//
// SplitShifts takes the day's work orders in sequence and splits them into the
// morning shift and the evening shift. It is three lines long, it is obviously
// correct, and every test anyone wrote for it passes.
//
// It is also the reason work order WO-3 vanished from the evening board last
// Tuesday. Not cancelled. Not rescheduled. Not logged anywhere. Gone, and
// replaced by a rush order that belonged to the morning.
//
// It passes a quick glance, it passes single-call tests, and it is quietly,
// badly broken the moment a caller does the most natural thing in the world:
// append a late-arriving order to the morning shift.
//
// YOUR TASKS
//
//  1. Find it by tracing, not by running. Write down len and cap for BOTH
//     returned slices when orders has 4 elements and morningCount is 2.
//     Then answer: when the caller appends one order to `morning`, does
//     append have a free slot to write into? Which index is that slot?
//     Which element of `evening` lives at that index? (NOTES section 3 has
//     the same trace with the same four work orders.)
//
//  2. Fix SplitShifts so that TestSplit_EveningSurvivesMorningAppend passes
//     while TestSplit_NoAllocations keeps passing. Read that second test
//     before you start — it rules out the fix you are about to reach for.
//     The scheduler runs this on ~50,000 orders on every planning request,
//     so copying the input is not on the table.
//
// STRETCH GOAL: make AppendRush safe from the other side too, so that a caller
// who got `morning` from somewhere else — an older version of this function, a
// different service, a test fixture — still cannot corrupt anything. What does
// that cost, and when would you accept it?
//
// Do not change the signatures.

// SplitShifts divides the day's work orders into the morning shift (the first
// morningCount orders) and the evening shift (everything after).
//
// BUG: the two halves share one backing array, and the morning half can write
// into the evening half's memory.
func SplitShifts(orders []string, morningCount int) (morning, evening []string) {
	if morningCount < 0 {
		morningCount = 0
	}
	if morningCount > len(orders) {
		morningCount = len(orders)
	}
	return orders[:morningCount], orders[morningCount:]
}

// AppendRush adds a late-arriving rush order to a shift.
// This function is correct. It is the victim, not the culprit.
func AppendRush(shift []string, orderID string) []string {
	return append(shift, orderID)
}
