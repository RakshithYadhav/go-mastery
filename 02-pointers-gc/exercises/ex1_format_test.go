package exercises

import (
	"strings"
	"testing"
)

func TestAppendFrame_Format(t *testing.T) {
	cases := []struct {
		name string
		f    Frame
		want string
	}{
		{"alarm", Frame{"MACHINE-07", 214.5, 9001}, "MACHINE-07,214.5,9001,ALARM\n"},
		{"ok", Frame{"PRESS-2", 88.0, 12}, "PRESS-2,88.0,12,OK\n"},
		{"exactly at the threshold is not an alarm", Frame{"OVEN-A", 200.0, 1}, "OVEN-A,200.0,1,OK\n"},
		{"just over the threshold is", Frame{"OVEN-A", 200.1, 2}, "OVEN-A,200.1,2,ALARM\n"},
		{"rounds to one decimal", Frame{"OVEN-A", 88.06, 3}, "OVEN-A,88.1,3,OK\n"},
		{"negative reading", Frame{"FREEZER-1", -18.4, 7}, "FREEZER-1,-18.4,7,OK\n"},
		{"zero", Frame{"MACHINE-07", 0, 0}, "MACHINE-07,0.0,0,OK\n"},
		{"empty machine name", Frame{"", 20.0, 5}, ",20.0,5,OK\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(AppendFrame(nil, tc.f))
			if got != tc.want {
				t.Fatalf("AppendFrame produced %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAppendFrame_AppendsRatherThanReplaces(t *testing.T) {
	dst := []byte("EXISTING\n")
	got := string(AppendFrame(dst, Frame{"PRESS-2", 88.0, 12}))
	want := "EXISTING\nPRESS-2,88.0,12,OK\n"
	if got != want {
		t.Fatalf("AppendFrame produced %q, want %q — it must APPEND to dst the way "+
			"the standard library's append-style functions do, not overwrite it", got, want)
	}
}

func TestAppendFrame_ManyFramesIntoOneBuffer(t *testing.T) {
	buf := make([]byte, 0, 4096)
	for i := 0; i < 50; i++ {
		buf = AppendFrame(buf, Frame{"MACHINE-07", float64(i), i})
	}
	lines := strings.Count(string(buf), "\n")
	if lines != 50 {
		t.Fatalf("50 frames produced %d lines, want 50 — every frame ends with exactly "+
			"one newline", lines)
	}
}

// The budget test. This is the exercise.
func TestAppendFrame_DoesNotAllocate(t *testing.T) {
	buf := make([]byte, 0, 256)
	f := Frame{"MACHINE-07", 214.5, 9001}

	got := testing.AllocsPerRun(1000, func() {
		buf = AppendFrame(buf[:0], f)
	})
	if got != 0 {
		t.Fatalf("AppendFrame allocated %.0f time(s) per call into a buffer that already "+
			"had room — the budget is 0. At 40,000 frames a second that is %.0f "+
			"allocations a second.\n\n"+
			"Every function in the fmt package takes ...any, which boxes each argument "+
			"before any formatting happens — so fmt cannot reach zero no matter how it "+
			"is called. strconv has an append-style function for each primitive type "+
			"that writes into a buffer you supply. NOTES section 4, and demo 04's "+
			"first row.", got, got*40000)
	}
}

func BenchmarkAppendFrame(b *testing.B) {
	buf := make([]byte, 0, 256)
	f := Frame{"MACHINE-07", 214.5, 9001}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendFrame(buf[:0], f)
	}
	_ = buf
}
