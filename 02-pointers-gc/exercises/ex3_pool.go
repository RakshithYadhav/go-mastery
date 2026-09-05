package exercises

// Exercise 3 — IMPLEMENT: a buffer pool that is actually safe to use.
//
// Demo 05 showed a pool doing its job, and then showed the two bugs that make
// pooling worse than not pooling: an object handed out without being reset, so
// one machine's data appears in another machine's frame; and one oversized
// object retained forever, so a 40 MB batch permanently changes the service's
// memory profile.
//
// Build the version that has neither bug.
//
// FIRST, A QUESTION ABOUT THE SIGNATURES BELOW
//
// Get returns *[]byte and Put takes *[]byte. That is deliberate and it is ugly,
// and before you write any code you should be able to say why it is not
// []byte.
//
// sync.Pool stores `any`. An interface value is one word: a pointer. A slice
// header is three words: address, length, capacity. So a []byte cannot be
// stored in an interface directly — the three words have to be put somewhere
// the one word can point at, and putting them somewhere is an allocation, on
// every single Put.
//
// A pool whose Put allocates on every call has given back a large part of what
// pooling was for. This is why the standard library's own pools hand out
// *bytes.Buffer and *[]byte rather than the values themselves. Look at how
// demo 05 moves the header in and out; it does the same thing.
//
// If you doubt this, do not take my word for it. Write the []byte version
// first, watch TestBufferPool_SteadyStateDoesNotAllocate fail at exactly one
// allocation per cycle, and then change the signature. Fifteen minutes, and
// you will never forget it.
//
// THE CONTRACT
//
//	NewBufferPool(maxCap) returns a pool that hands out byte buffers.
//	  maxCap is the largest capacity the pool will retain. Panics if maxCap < 1.
//
//	Get() returns a pointer to a buffer whose LENGTH IS ZERO, always, whatever
//	  the previous user did with it. A caller must never be able to observe a
//	  previous caller's bytes. In steady state Get must not allocate.
//
//	Put(bufp) offers a buffer back to the pool. If the buffer's capacity is
//	  greater than maxCap, Put must NOT retain it — drop it and let the
//	  collector have it. Put must be safe to call with nil.
//
//	Stats() returns two counters, for the tests and for your own understanding:
//	  news  — how many times the pool had to create a fresh buffer
//	  drops — how many times Put refused a buffer for being too large
//
//	The whole type must be safe for concurrent use, because the collector it
//	serves runs one goroutine per socket. sync.Pool is already concurrent-safe.
//	Your counters are not, and that is the part to think about.
//
// THINK BEFORE YOU TYPE
//
//	The shape:     which side does the reset belong on, Get or Put? Demo 05
//	               argues for one of them. What happens to the other choice
//	               when a caller returns early or panics before its Put?
//	The ownership: after Put(bufp), may the caller keep using the buffer? Write
//	               down your answer, then write the doc comment that states it.
//	The exits:     Put(nil). A buffer far over maxCap. Get on a brand new pool.
//	               Two goroutines calling Get at the same instant.
//
// Do not change the signatures.

// BufferPool hands out reusable byte buffers with a capacity ceiling.
type BufferPool struct {
	// TODO: your fields here.
}

// NewBufferPool returns a pool that retains buffers up to maxCap bytes.
// It panics if maxCap < 1.
func NewBufferPool(maxCap int) *BufferPool {
	// TODO: implement.
	return &BufferPool{}
}

// Get returns a pointer to a zero-length buffer.
// It must not allocate in steady state.
func (p *BufferPool) Get() *[]byte {
	// TODO: implement.
	return nil
}

// Put offers a buffer back to the pool. Buffers with a capacity above the
// pool's maxCap are dropped rather than retained.
func (p *BufferPool) Put(bufp *[]byte) {
	// TODO: implement.
}

// Stats reports how many buffers the pool created and how many it refused.
func (p *BufferPool) Stats() (news, drops int) {
	// TODO: implement.
	return 0, 0
}
