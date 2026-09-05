package exercises

import (
	"reflect"
	"testing"
)

func TestTagCounter_Counts(t *testing.T) {
	c := NewTagCounter()
	for i := 0; i < 3; i++ {
		c.Add([]byte("MACHINE-07"))
	}
	c.Add([]byte("PRESS-2"))

	if got := c.Count("MACHINE-07"); got != 3 {
		t.Fatalf("Count(MACHINE-07) = %d, want 3", got)
	}
	if got := c.Count("PRESS-2"); got != 1 {
		t.Fatalf("Count(PRESS-2) = %d, want 1", got)
	}
	if got := c.Count("OVEN-A"); got != 0 {
		t.Fatalf("Count of a tag never seen = %d, want 0", got)
	}
}

// The budget test. This is the exercise.
func TestTagCounter_KnownTagDoesNotAllocate(t *testing.T) {
	c := NewTagCounter()
	tag := []byte("MACHINE-07")
	c.Add(tag) // the tag is now known; every later Add is the hot path

	got := testing.AllocsPerRun(1000, func() {
		c.Add(tag)
	})
	if got != 0 {
		t.Fatalf("Add allocated %.0f time(s) per call for a tag that was already known — "+
			"the budget is 0. At 40,000 frames a second that is %.0f allocations a second "+
			"to look up a key created months ago.\n\n"+
			"`c.counts[string(tag)]++` is a map WRITE, and a write has to be able to store "+
			"the key, so it copies every time. Reading the map with the same expression is "+
			"free. Restructure so the hot path only reads. NOTES section 5, the last "+
			"subsection, has the measured numbers.", got, got*40000)
	}
}

// The collector reuses one read buffer for every frame off the socket.
func TestTagCounter_DoesNotRetainCallerBuffer(t *testing.T) {
	c := NewTagCounter()
	buf := make([]byte, 0, 64)

	buf = append(buf[:0], "MACHINE-07"...)
	c.Add(buf)
	buf = append(buf[:0], "PRESS-2"...)
	c.Add(buf)
	buf = append(buf[:0], "OVEN-A"...)
	c.Add(buf)

	// Scribble over the buffer the way the next socket read would.
	buf = append(buf[:0], "ZZZZZZZZZZ"...)
	_ = buf

	for _, tag := range []string{"MACHINE-07", "PRESS-2", "OVEN-A"} {
		if got := c.Count(tag); got != 1 {
			t.Fatalf("Count(%s) = %d, want 1 — the caller reuses one buffer for every "+
				"frame, so anything you kept a reference to has since been overwritten. "+
				"The tag has to become yours at the moment you store it.", tag, got)
		}
	}
}

func TestTagCounter_TagsAreSorted(t *testing.T) {
	c := NewTagCounter()
	for _, tag := range []string{"OVEN-A", "MACHINE-07", "PRESS-2", "MACHINE-07"} {
		c.Add([]byte(tag))
	}

	want := []string{"MACHINE-07", "OVEN-A", "PRESS-2"}
	for i := 0; i < 5; i++ { // ranging a map gives a different order every time
		got := c.Tags()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Tags() = %v, want %v — the dashboard renders these in a table and "+
				"the order must be identical between refreshes. Map iteration order is "+
				"randomised on every range; NOTES section 5, consequence 2.", got, want)
		}
	}
}

func TestTagCounter_BoringCases(t *testing.T) {
	t.Run("empty counter", func(t *testing.T) {
		c := NewTagCounter()
		if got := c.Count("anything"); got != 0 {
			t.Fatalf("Count on an empty counter = %d, want 0", got)
		}
		if got := c.Tags(); len(got) != 0 {
			t.Fatalf("Tags() on an empty counter = %v, want empty", got)
		}
	})

	t.Run("zero-length tag", func(t *testing.T) {
		c := NewTagCounter()
		c.Add([]byte{})
		if got := c.Count(""); got != 1 {
			t.Fatalf("Count(\"\") = %d, want 1 — an empty tag is still a tag", got)
		}
	})

	t.Run("same tag many times", func(t *testing.T) {
		c := NewTagCounter()
		for i := 0; i < 40_000; i++ {
			c.Add([]byte("MACHINE-07"))
		}
		if got := c.Count("MACHINE-07"); got != 40_000 {
			t.Fatalf("Count after 40,000 events = %d, want 40000", got)
		}
	})
}
