package exercises

import "testing"

func feed(s *ShiftStats) {
	s.Observe("MACHINE-07", 180.0, 1)
	s.Observe("PRESS-2", 214.5, 2)
	s.Observe("MACHINE-07", 195.25, 3)
	s.Observe("OVEN-A", 88.0, 4)
	s.Observe("PRESS-2", 201.0, 5)
}

func TestStats_StillCorrect(t *testing.T) {
	s := NewShiftStats()
	feed(s)

	if s.Count != 5 {
		t.Fatalf("Count = %d, want 5", s.Count)
	}
	if s.Alarms != 2 {
		t.Fatalf("Alarms = %d, want 2 (only readings strictly above 200)", s.Alarms)
	}
	if s.Max != 214.5 {
		t.Fatalf("Max = %v, want 214.5", s.Max)
	}
	if want := 180.0 + 214.5 + 195.25 + 88.0 + 201.0; s.Sum != want {
		t.Fatalf("Sum = %v, want %v", s.Sum, want)
	}
}

func TestStats_ReturnsAlarmFlag(t *testing.T) {
	s := NewShiftStats()
	if got := s.Observe("MACHINE-07", 199.9, 1); got {
		t.Fatal("Observe reported an alarm for 199.9; the threshold is strictly above 200")
	}
	if got := s.Observe("MACHINE-07", 200.0, 2); got {
		t.Fatal("Observe reported an alarm for exactly 200.0; the threshold is strictly above")
	}
	if got := s.Observe("MACHINE-07", 200.1, 3); !got {
		t.Fatal("Observe did not report an alarm for 200.1")
	}
}

func TestStats_SummaryUnchanged(t *testing.T) {
	s := NewShiftStats()
	feed(s)

	got := s.Summary()
	want := "frames=5 mean=175.75 max=214.5 alarms=2 hottest=PRESS-2@214.5"
	if got != want {
		t.Fatalf("Summary() = %q,\n                want %q\n\n"+
			"The report text is part of the contract. Whatever you changed inside "+
			"Observe, the end-of-shift output must be identical.", got, want)
	}
}

func TestStats_EmptyShift(t *testing.T) {
	s := NewShiftStats()
	if got := s.Summary(); got != "no frames" {
		t.Fatalf("Summary() on an empty shift = %q, want %q", got, "no frames")
	}
}

func TestStats_PerMachineCounts(t *testing.T) {
	s := NewShiftStats()
	feed(s)

	// seq%4 picks the bank: seq 1 -> bank 1, seq 3 -> bank 3.
	if got := s.MachineCount("MACHINE-07", 1); got != 1 {
		t.Fatalf("MachineCount(MACHINE-07, 1) = %d, want 1", got)
	}
	if got := s.MachineCount("MACHINE-07", 3); got != 1 {
		t.Fatalf("MachineCount(MACHINE-07, 3) = %d, want 1", got)
	}
	if got := s.MachineCount("OVEN-A", 0); got != 1 {
		t.Fatalf("MachineCount(OVEN-A, 0) = %d, want 1", got)
	}
	if got := s.MachineCount("NEVER-SEEN", 0); got != 0 {
		t.Fatalf("MachineCount for an unseen machine = %d, want 0", got)
	}
}

// The budget test. This is the exercise.
func TestStats_ObserveDoesNotAllocate(t *testing.T) {
	s := NewShiftStats()
	// Warm up: every machine and bank this test uses must already exist, so the
	// only allocations left are the per-call ones you are hunting.
	for i := 0; i < 8; i++ {
		s.Observe("MACHINE-07", float64(100+i), i)
	}

	got := testing.AllocsPerRun(1000, func() {
		s.Observe("MACHINE-07", 150.0, 2)
	})
	if got != 0 {
		t.Fatalf("Observe allocated %.0f time(s) per call — the budget is 0. At 40,000 "+
			"frames a second that is %.0f allocations a second on the hottest path in "+
			"the service.\n\n"+
			"There are three causes and they are three different kinds of problem:\n"+
			"  - one is boxing values into an interface,\n"+
			"  - one is formatting a string that nobody reads until the shift ends,\n"+
			"  - one is a map WRITE with a []byte-to-string key, which Module 1 "+
			"section 5 measured at one allocation every time.\n\n"+
			"Run: go build -gcflags='-m -l' ./02-pointers-gc/exercises\n"+
			"Two of the three appear in that output. The third does not, and working "+
			"out why is the point.", got, got*40000)
	}
}

func BenchmarkStatsObserve(b *testing.B) {
	s := NewShiftStats()
	s.Observe("MACHINE-07", 150.0, 2)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Observe("MACHINE-07", 150.0, 2)
	}
}
