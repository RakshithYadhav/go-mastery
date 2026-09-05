package exercises

// Exercise 3 — IMPLEMENT: filter in place without leaking what you dropped.
//
// The planning service holds the day's work orders in one long slice and
// re-filters it constantly: only active orders, only this line's orders, only
// what is still unstarted. The slice is large and the filtering is hot, so the
// standard `out := make(...); for ... { out = append(out, x) }` shape is the
// wrong tool — it allocates a second slice every call.
//
// The in-place filter is the right tool, and it has a trap that costs memory
// instead of correctness. You met it in NOTES section 4: shrinking a slice of
// POINTERS does not release what the dropped elements point at, because those
// pointers are still sitting in the backing array past the new length. The
// garbage collector cannot tell "past len" from "in use" — it only sees an
// array full of live pointers.
//
// THE CONTRACT
//
//	CompactActive(orders) returns a slice containing only the orders whose
//	  Status is "ACTIVE", in their original relative order.
//
//	  It must REUSE the caller's backing array. Zero allocations.
//
//	  Every slot in the backing array past the returned length must be nil
//	  when it returns. A *WorkOrder that did not survive the filter must not
//	  be reachable through the array afterwards.
//
//	  A nil or empty input returns an empty result and does not panic.
//
// THINK BEFORE YOU TYPE
//
//	The shape:     two indexes walking the same array, one reading and one
//	               writing, the writer never ahead of the reader. Write it out
//	               on paper for [A(active), B(done), C(active)] before you code.
//	The ownership: after the write index passes position 1, what is still
//	               sitting at position 2? Is it a pointer? Who else knows
//	               about it?
//	The exits:     what if nothing is active? What if everything is? Does your
//	               nil-out loop have an off-by-one at either end?
//
// AN IMPLEMENTATION HINT YOU ARE ALLOWED (a design fact, not a solution): Go
// has a builtin `clear` that zeroes a slice, and `slices.Delete` in the
// standard library does this nil-ing for exactly this reason. You may use
// either, or write the loop yourself.
//
// Do not change the signatures.

// WorkOrder is one job on the shop floor. It is deliberately fat: the leak this
// exercise is about is only visible when the dropped objects are big.
type WorkOrder struct {
	ID     string
	Status string
	Notes  []byte // operator notes, routing history, QA comments: kilobytes each
}

// CompactActive filters orders down to the ACTIVE ones, in place, and leaves no
// dropped *WorkOrder reachable through the backing array.
func CompactActive(orders []*WorkOrder) []*WorkOrder {
	// TODO: implement.
	return orders
}
