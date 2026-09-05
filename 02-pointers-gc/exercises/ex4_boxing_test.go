package exercises

import (
	"strings"
	"testing"
)

func TestLogger_StillLogs(t *testing.T) {
	l := NewLogger(LevelDebug)
	l.Debug("frame received", "machine", "MACHINE-07", "seq", 9001)
	l.Info("shift ended", "frames", 40000)

	lines := l.Lines()
	if len(lines) != 2 {
		t.Fatalf("wrote 2 lines at debug level, got %d: %v", len(lines), lines)
	}
	if !strings.HasPrefix(lines[0], "DEBUG ") {
		t.Fatalf("first line = %q, want it to start with %q", lines[0], "DEBUG ")
	}
	if !strings.HasPrefix(lines[1], "INFO ") {
		t.Fatalf("second line = %q, want it to start with %q", lines[1], "INFO ")
	}
}

func TestLogger_Format(t *testing.T) {
	l := NewLogger(LevelDebug)
	l.Debug("frame received", "machine", "MACHINE-07", "seq", 9001)

	got := l.Lines()[0]
	want := "DEBUG frame received machine=MACHINE-07 seq=9001"
	if got != want {
		t.Fatalf("log line = %q,\n          want %q\n\n"+
			"The output format is part of the contract. Whatever you changed to kill "+
			"the allocations, a line that IS written must look exactly the same.", got, want)
	}
}

func TestLogger_RespectsLevel(t *testing.T) {
	l := NewLogger(LevelInfo)
	l.Debug("frame received", "machine", "MACHINE-07", "seq", 9001)

	if n := len(l.Lines()); n != 0 {
		t.Fatalf("debug line was written at info level: %v", l.Lines())
	}

	l.Info("shift ended", "frames", 40000)
	if n := len(l.Lines()); n != 1 {
		t.Fatalf("info line was not written at info level, got %d lines", n)
	}
}

func TestLogger_OddNumberOfKeyValues(t *testing.T) {
	l := NewLogger(LevelDebug)
	l.Debug("frame received", "machine")

	got := l.Lines()[0]
	want := "DEBUG frame received"
	if got != want {
		t.Fatalf("log line = %q, want %q — a key with no value is dropped, not printed "+
			"half-formed", got, want)
	}
}

// The budget test. This is the exercise.
func TestLogger_DisabledLevelDoesNotAllocate(t *testing.T) {
	l := NewLogger(LevelInfo) // debug is OFF
	machine := "MACHINE-07"
	seq := 9001

	got := testing.AllocsPerRun(1000, func() {
		logDebugFrame(l, machine, seq)
	})
	if got != 0 {
		t.Fatalf("logging at a DISABLED level allocated %.0f time(s) per call — the "+
			"budget is 0. At 40,000 frames a second that is %.0f allocations a second "+
			"to build arguments for a function that throws them away.\n\n"+
			"Moving the level check earlier inside the function does not help. A "+
			"variadic ...any parameter boxes every argument at the CALL SITE, before "+
			"the function is entered, so the work is finished before any check can "+
			"run. NOTES section 4, and demo 04's first row.\n\n"+
			"The exercise comment describes the two established fixes. Pick one, and "+
			"put your reasoning in MISTAKES.md.", got, got*40000)
	}
}

// logDebugFrame is the call site under test. Change its BODY to match whichever
// fix you chose — that is allowed and expected. Its signature stays.
func logDebugFrame(l *Logger, machine string, seq int) {
	l.Debug("frame received", "machine", machine, "seq", seq)
}

func BenchmarkLogger_Disabled(b *testing.B) {
	l := NewLogger(LevelInfo)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logDebugFrame(l, "MACHINE-07", 9001)
	}
}

func BenchmarkLogger_Enabled(b *testing.B) {
	l := NewLogger(LevelDebug)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logDebugFrame(l, "MACHINE-07", 9001)
		l.lines = l.lines[:0]
	}
}
