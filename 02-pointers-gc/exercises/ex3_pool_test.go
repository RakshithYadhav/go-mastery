package exercises

import (
	"sync"
	"testing"
)

func TestBufferPool_GetReturnsEmptyBuffer(t *testing.T) {
	p := NewBufferPool(1024)

	bufp := p.Get()
	if bufp == nil {
		t.Fatal("Get returned nil")
	}
	if len(*bufp) != 0 {
		t.Fatalf("Get on a new pool returned a buffer of length %d, want 0", len(*bufp))
	}

	*bufp = append(*bufp, "MACHINE-07,214.5"...)
	p.Put(bufp)

	next := p.Get()
	if len(*next) != 0 {
		t.Fatalf("Get returned a buffer of length %d holding %q.\n\n"+
			"The previous caller's bytes are visible to this one. In a service that "+
			"encodes one customer's data per buffer, this is how one customer's data "+
			"ends up in another's response. Demo 05, experiment 3.",
			len(*next), string(*next))
	}
}

func TestBufferPool_ReusesCapacity(t *testing.T) {
	p := NewBufferPool(4096)

	bufp := p.Get()
	*bufp = append(*bufp, make([]byte, 512)...)
	p.Put(bufp)

	next := p.Get()
	if cap(*next) < 512 {
		t.Fatalf("Get returned a buffer with capacity %d after a 512-byte buffer was "+
			"returned to the pool — the capacity is the whole point of pooling. "+
			"Reset the LENGTH on the way out, not the capacity.", cap(*next))
	}
}

func TestBufferPool_DropsOversizedBuffers(t *testing.T) {
	const maxCap = 1024
	p := NewBufferPool(maxCap)

	small := p.Get()
	*small = append(*small, make([]byte, 100)...)
	p.Put(small)

	// One enormous batch arrives, as it does once a day.
	big := p.Get()
	*big = append(*big, make([]byte, 8<<20)...)
	bigCap := cap(*big)
	p.Put(big)

	_, drops := p.Stats()
	if drops == 0 {
		t.Fatalf("Put retained a buffer of capacity %d when maxCap is %d, and Stats "+
			"reports 0 drops.\n\n"+
			"Every small frame from now on borrows an 8 MB buffer, and the pool holds "+
			"it alive between uses. One outlier permanently changed the service's "+
			"memory profile. Demo 05, experiment 4.", bigCap, maxCap)
	}

	for i := 0; i < 20; i++ {
		if got := p.Get(); cap(*got) > maxCap {
			t.Fatalf("Get returned a buffer with capacity %d, above the pool's maxCap "+
				"of %d — the oversized buffer is still in circulation", cap(*got), maxCap)
		}
	}
}

func TestBufferPool_PutIsSafeWithNil(t *testing.T) {
	p := NewBufferPool(1024)
	p.Put(nil)

	empty := []byte{}
	p.Put(&empty)

	if got := p.Get(); len(*got) != 0 {
		t.Fatalf("Get returned length %d after nil and empty Puts, want 0", len(*got))
	}
}

// The budget test. This is the exercise.
func TestBufferPool_SteadyStateDoesNotAllocate(t *testing.T) {
	p := NewBufferPool(4096)

	// Prime the pool so the steady-state path is what gets measured.
	for i := 0; i < 8; i++ {
		bufp := p.Get()
		*bufp = append(*bufp, make([]byte, 256)...)
		p.Put(bufp)
	}

	got := testing.AllocsPerRun(1000, func() {
		bufp := p.Get()
		*bufp = append(*bufp, "MACHINE-07,214.5,9001,ALARM\n"...)
		p.Put(bufp)
	})
	if got != 0 {
		t.Fatalf("A Get/append/Put cycle allocated %.0f time(s) — the budget is 0 once "+
			"the pool is warm.\n\n"+
			"If this says exactly 1, you are almost certainly creating a NEW pointer on "+
			"the way back in: taking the address of a local slice variable inside Put "+
			"allocates a fresh three-word header every call, which is the cost the "+
			"*[]byte signature exists to avoid. Put back the pointer the caller handed "+
			"you, rather than a pointer to a copy of it.", got)
	}
}

func TestBufferPool_ConcurrentUse(t *testing.T) {
	p := NewBufferPool(4096)

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				bufp := p.Get()
				if len(*bufp) != 0 {
					t.Errorf("Get returned a buffer of length %d under concurrent use", len(*bufp))
					return
				}
				*bufp = append(*bufp, "frame"...)
				p.Put(bufp)
			}
		}()
	}
	wg.Wait()

	news, _ := p.Stats()
	if news == 0 {
		t.Fatal("Stats reports the pool never created a buffer, which cannot be true " +
			"after 1600 Gets — the counters are not being updated")
	}
}

func TestBufferPool_RejectsUselessCapacity(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewBufferPool(0) returned normally — the contract says it panics")
		}
	}()
	NewBufferPool(0)
}
