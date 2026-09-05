package exercises

// Exercise 1 — IMPLEMENT: format a line with no allocations at all.
//
// The telemetry collector writes one CSV line per frame to a rolling log. The
// current implementation is a single call to fmt.Sprintf, which demo 04
// measured at three allocations per call. At 40,000 frames a second that is
// 120,000 allocations a second, all of them dead within microseconds.
//
// Your job is the replacement: a function that writes the same line into a
// buffer the caller already owns, and allocates nothing.
//
// THE CONTRACT
//
//	AppendFrame(dst, f) appends the CSV encoding of f to dst and returns the
//	  extended slice, exactly the way `append` itself does. The caller owns dst
//	  and is responsible for its capacity.
//
//	The format, with no spaces anywhere:
//	  machine,celsius,seq,status\n
//	where celsius has exactly one decimal place, seq is a base-10 integer, and
//	status is the literal text OK or ALARM. A frame is ALARM when Celsius is
//	strictly greater than 200.
//
//	Example: MACHINE-07,214.5,9001,ALARM\n
//
//	AppendFrame must perform ZERO allocations when dst has spare capacity.
//	That rules out fmt entirely — every function in that package takes ...any
//	and boxes its arguments, which allocates before it formats anything.
//
// AN IMPLEMENTATION HINT YOU ARE ALLOWED (it is a design fact, not a solution):
// the standard library has an append-style formatter for every primitive type.
// Look at strconv.AppendInt and strconv.AppendFloat, and read their signatures
// carefully — the argument order and the format arguments are not obvious.
//
// THINK BEFORE YOU TYPE
//
//	The shape:     every line of your implementation should end up assigning
//	               back into dst. If a line does not, ask why not.
//	The ownership: who owns the memory the returned slice points at, before and
//	               after the call? What happens if the caller keeps the return
//	               value and calls you again with the same dst?
//	The exits:     a nil dst, a dst with existing content that must be
//	               preserved, and a machine name containing no characters.
//
// Do not change the signatures.

// Frame is one telemetry frame from one machine.
type Frame struct {
	Machine string
	Celsius float64
	Seq     int
}

// AppendFrame appends f's CSV encoding to dst and returns the extended slice.
// It must not allocate when dst has spare capacity.
func AppendFrame(dst []byte, f Frame) []byte {
	// TODO: implement.
	return dst
}
