package exercises

// Exercise 1 — IMPLEMENT: a ring buffer that never allocates after construction.
//
// The machine-monitoring service keeps the last N readings for every machine on
// the shop floor so an operator can pull up a short history when an alarm
// fires. There are ~2,000 machines and readings arrive several times a second.
// Anything that allocates on the write path allocates a few thousand times a
// second, forever.
//
// You met the underlying idea in NOTES section 2: if you know how many
// elements you will hold, say so once, at construction, and then never grow.
// A ring buffer is that idea taken to its conclusion — a fixed array plus two
// integers, where "adding" the 501st item to a 500-slot ring overwrites the
// oldest one instead of growing anything.
//
// THE CONTRACT
//
//	NewRing(capacity) returns a ring that can hold `capacity` readings.
//	  Panics if capacity < 1. This is the ONLY function allowed to allocate.
//
//	Push(r) adds a reading. When the ring is full it OVERWRITES THE OLDEST
//	  reading. Push must not allocate. Not once, not amortised — zero.
//
//	Len() returns how many readings are currently held: it climbs to capacity
//	  and then stays there.
//
//	Snapshot(dst) appends every held reading to dst in OLDEST-TO-NEWEST order
//	  and returns the result, exactly the way the standard library's append-style
//	  functions do. If dst has enough capacity, Snapshot must not allocate
//	  either. A nil dst is legal and may allocate.
//
// AN IMPLEMENTATION HINT YOU ARE ALLOWED (it is a design fact, not a solution):
// the usual shape is one slice made at capacity, plus a write position and a
// count. Oldest-to-newest is then an index calculation, not a data move.
//
// THINK BEFORE YOU TYPE
//
//	The shape:     what exactly does NewRing allocate, and how many times?
//	The ownership: after Push overwrites slot 3, who else could be looking at
//	               what used to be in slot 3? Does Snapshot hand out anything
//	               that a later Push can write through? (Re-read section 3.)
//	The exits:     what does Snapshot do when the ring is empty? When it has
//	               wrapped exactly once? When it has wrapped seventeen times?
//
// ASSUMPTIONS
//   - Single goroutine. No locking needed; this is not a concurrency exercise.
//   - capacity is small (tens to thousands), so a plain array is right.
//
// Do not change the signatures.

// Reading is one sample from one machine.
type Reading struct {
	MachineID string
	Celsius   float64
	Seq       int
}

// Ring holds the most recent N readings, overwriting the oldest when full.
type Ring struct {
	Readings []Reading
	Capacity int
	Count    int
}

// NewRing returns a Ring that holds up to capacity readings.
// It panics if capacity < 1. This is the only place allocation is allowed.
func NewRing(capacity int) *Ring {
	if capacity < 1 {
		panic("Capacity is less than 1 cannot accept such a value")
	}

	return &Ring{
		Readings: make([]Reading, capacity),
		Capacity: capacity,
		Count:    0,
	}
}

// Push adds r, overwriting the oldest reading when the ring is full.
// Push must perform zero allocations.
func (rg *Ring) Push(r Reading) {
	relativeLen := rg.Count % rg.Capacity
	rg.Readings[relativeLen] = r
	rg.Count += 1

}

// Len returns the number of readings currently held.
func (rg *Ring) Len() int {
	if rg.Count < rg.Capacity {
		return rg.Count
	}

	return rg.Capacity
}

// Snapshot appends the held readings to dst, oldest first, and returns the
// extended slice. It must not allocate when dst has enough spare capacity.
func (rg *Ring) Snapshot(dst []Reading) []Reading {
	start := (rg.Count - rg.Len()) % rg.Capacity
	for i := range rg.Len() {
		dst = append(dst, rg.Readings[(start+i)%rg.Capacity])
	}
	return dst
}
