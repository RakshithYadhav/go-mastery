package exercises

import (
	"reflect"
	"testing"
)

func mkReadings(n int) []Reading {
	out := make([]Reading, n)
	for i := range out {
		out[i] = Reading{MachineID: "MACHINE-07", Celsius: float64(20 + i), Seq: i}
	}
	return out
}

func seqs(rs []Reading) []int {
	out := make([]int, len(rs))
	for i, r := range rs {
		out[i] = r.Seq
	}
	return out
}

func TestRing_KeepsOrderWhenNotFull(t *testing.T) {
	r := NewRing(5)
	for _, rd := range mkReadings(3) {
		r.Push(rd)
	}
	got := seqs(r.Snapshot(nil))
	want := []int{0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Snapshot returned %v, want %v — readings must come back oldest first", got, want)
	}
}

func TestRing_OverwritesOldest(t *testing.T) {
	r := NewRing(3)
	for _, rd := range mkReadings(7) { // seq 0..6 into a 3-slot ring
		r.Push(rd)
	}
	got := seqs(r.Snapshot(nil))
	want := []int{4, 5, 6}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("after pushing 7 readings into a 3-slot ring, Snapshot returned %v, want %v — "+
			"a full ring must overwrite the OLDEST reading and still report oldest-to-newest", got, want)
	}
}

func TestRing_LenCapsAtCapacity(t *testing.T) {
	r := NewRing(4)
	if r.Len() != 0 {
		t.Fatalf("a fresh ring reported Len() = %d, want 0", r.Len())
	}
	for i, rd := range mkReadings(10) {
		r.Push(rd)
		want := min(i+1, 4)
		if r.Len() != want {
			t.Fatalf("after %d pushes into a 4-slot ring, Len() = %d, want %d", i+1, r.Len(), want)
		}
	}
}

func TestRing_PushDoesNotAllocate(t *testing.T) {
	r := NewRing(256)
	rd := Reading{MachineID: "PRESS-2", Celsius: 41.5, Seq: 1}

	got := testing.AllocsPerRun(1000, func() {
		r.Push(rd)
	})
	if got != 0 {
		t.Fatalf("Push allocated %.0f time(s) per call — the budget is 0. "+
			"At 40,000 readings a second that is %.0f allocations a second for data "+
			"whose size you knew at construction. NewRing is the only function allowed "+
			"to allocate; see NOTES section 2.", got, got*40000)
	}
}

func TestRing_SnapshotDoesNotAllocateIntoSizedBuffer(t *testing.T) {
	r := NewRing(64)
	for _, rd := range mkReadings(64) {
		r.Push(rd)
	}
	dst := make([]Reading, 0, 64)

	got := testing.AllocsPerRun(1000, func() {
		dst = r.Snapshot(dst[:0])
	})
	if got != 0 {
		t.Fatalf("Snapshot allocated %.0f time(s) per call into a buffer that already had "+
			"room for every reading — the budget is 0. Appending into a slice with spare "+
			"capacity writes in place; see NOTES section 2.", got)
	}
	if len(dst) != 64 {
		t.Fatalf("Snapshot returned %d readings, want 64", len(dst))
	}
}

func TestRing_SnapshotSurvivesLaterPushes(t *testing.T) {
	r := NewRing(3)
	for _, rd := range mkReadings(3) { // seq 0,1,2
		r.Push(rd)
	}
	snap := r.Snapshot(nil)
	before := seqs(snap)

	// The operator is still looking at this snapshot while new readings arrive.
	for _, rd := range mkReadings(3) {
		rd.Seq += 100
		r.Push(rd)
	}

	after := seqs(snap)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("a snapshot taken earlier changed from %v to %v after three more pushes — "+
			"Snapshot handed the caller a window onto the ring's own array instead of a copy. "+
			"NOTES section 3 is about exactly this.", before, after)
	}
}

func TestRing_BoringCases(t *testing.T) {
	t.Run("empty ring", func(t *testing.T) {
		r := NewRing(4)
		if got := r.Snapshot(nil); len(got) != 0 {
			t.Fatalf("Snapshot of an empty ring returned %d readings, want 0", len(got))
		}
	})

	t.Run("capacity one", func(t *testing.T) {
		r := NewRing(1)
		for _, rd := range mkReadings(5) {
			r.Push(rd)
		}
		got := seqs(r.Snapshot(nil))
		if !reflect.DeepEqual(got, []int{4}) {
			t.Fatalf("a 1-slot ring after 5 pushes returned %v, want [4]", got)
		}
	})

	t.Run("appends to a non-empty dst", func(t *testing.T) {
		r := NewRing(2)
		r.Push(Reading{Seq: 8})
		r.Push(Reading{Seq: 9})
		dst := []Reading{{Seq: 1}}
		got := seqs(r.Snapshot(dst))
		if !reflect.DeepEqual(got, []int{1, 8, 9}) {
			t.Fatalf("Snapshot(dst) returned %v, want [1 8 9] — it must APPEND to dst, "+
				"the way the standard library's append-style functions do", got)
		}
	})

	t.Run("rejects a useless capacity", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("NewRing(0) returned normally — the contract says it panics")
			}
		}()
		NewRing(0)
	})
}
