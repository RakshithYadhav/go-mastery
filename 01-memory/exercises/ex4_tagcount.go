package exercises

// Exercise 4 — IMPLEMENT: count events from a byte stream without allocating.
//
// The telemetry collector reads machine status frames off a socket. Each frame
// carries a machine tag as raw bytes — "MACHINE-07", "PRESS-2", "OVEN-A" —
// and the collector keeps a running count per tag so the dashboard can show
// event rates. About 40,000 frames a second across the plant. The set of tags
// is small and fixed: roughly 2,000 machines, and a new one appears when
// somebody installs hardware, which is a few times a year.
//
// The obvious implementation is one line:
//
//	c.counts[string(tag)]++
//
// It is correct. It also allocates a fresh string on every single frame — 40,000
// allocations a second to look up one of 2,000 keys that were all created
// months ago. NOTES section 5 has the measured numbers and the reason: reading
// a map with a []byte key is free, writing one is not, and `++` is a write.
//
// THE CONTRACT
//
//	NewTagCounter() returns an empty counter.
//
//	Add(tag) records one event for that tag.
//	  For a tag that has been seen before, Add must perform ZERO allocations.
//	  For a tag seen for the first time, Add may allocate — once, ever.
//	  Add must not retain the caller's slice. The caller reuses one read
//	  buffer for every frame; if you keep a reference to it, every tag in your
//	  map turns into whatever the last frame contained.
//
//	Count(tag string) returns the count for a tag, or 0 if unseen.
//
//	Tags() returns every tag seen so far, SORTED. (The dashboard renders them
//	  in a table and the order must be stable between refreshes — NOTES
//	  section 5, consequence 2.)
//
// THINK BEFORE YOU TYPE
//
//	The shape:     if the hot path must not write to the map, what is in the
//	               map such that "increment" is not a map write at all?
//	The ownership: which single line in your Add is allowed to allocate, and
//	               what guarantees it runs only for genuinely new tags?
//	               Where exactly does the caller's buffer stop being borrowed
//	               and become yours?
//	The exits:     empty counter, unknown tag, the same tag 40,000 times, a
//	               zero-length tag.
//
// Do not change the signatures.

// TagCounter counts events per machine tag, reading tags as raw bytes.
type TagCounter struct {
	// TODO: your fields here.
}

// NewTagCounter returns an empty TagCounter.
func NewTagCounter() *TagCounter {
	// TODO: implement.
	return &TagCounter{}
}

// Add records one event for tag. Zero allocations for an already-known tag.
// Must not retain tag — the caller reuses that buffer.
func (c *TagCounter) Add(tag []byte) {
	// TODO: implement.
}

// Count returns how many events were recorded for tag.
func (c *TagCounter) Count(tag string) int64 {
	// TODO: implement.
	return 0
}

// Tags returns every tag seen so far, sorted.
func (c *TagCounter) Tags() []string {
	// TODO: implement.
	return nil
}
