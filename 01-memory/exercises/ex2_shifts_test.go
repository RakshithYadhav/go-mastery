package exercises

import (
	"reflect"
	"testing"
)

func day() []string {
	return []string{"WO-1", "WO-2", "WO-3", "WO-4"}
}

func TestSplit_SplitsCorrectly(t *testing.T) {
	morning, evening := SplitShifts(day(), 2)
	if !reflect.DeepEqual(morning, []string{"WO-1", "WO-2"}) {
		t.Fatalf("morning = %v, want [WO-1 WO-2]", morning)
	}
	if !reflect.DeepEqual(evening, []string{"WO-3", "WO-4"}) {
		t.Fatalf("evening = %v, want [WO-3 WO-4]", evening)
	}
}

// This is the failing test. It reproduces last Tuesday.
func TestSplit_EveningSurvivesMorningAppend(t *testing.T) {
	orders := day()
	morning, evening := SplitShifts(orders, 2)

	morning = AppendRush(morning, "RUSH-9")

	if want := []string{"WO-1", "WO-2", "RUSH-9"}; !reflect.DeepEqual(morning, want) {
		t.Fatalf("morning = %v, want %v", morning, want)
	}
	if want := []string{"WO-3", "WO-4"}; !reflect.DeepEqual(evening, want) {
		t.Fatalf("evening = %v, want %v.\n\n"+
			"Nobody wrote to evening. One order was appended to the MORNING shift and "+
			"the evening board changed. Work order WO-3 is now unreachable: not "+
			"cancelled, not rescheduled, not logged.\n\n"+
			"Trace it: what were len and cap of `morning` before the append? "+
			"Which index did append write into, and which element of `evening` "+
			"lives at that index? NOTES section 3.", evening, want)
	}
}

// The guard rail. It rules out "just copy both halves", which fixes the bug
// above and quietly costs 50,000 string copies on every planning request.
func TestSplit_NoAllocations(t *testing.T) {
	orders := make([]string, 50_000)
	for i := range orders {
		orders[i] = "WO-X"
	}

	got := testing.AllocsPerRun(100, func() {
		m, e := SplitShifts(orders, 25_000)
		_, _ = m, e
	})
	if got != 0 {
		t.Fatalf("SplitShifts allocated %.0f time(s) per call — the budget is 0.\n\n"+
			"Copying the input does fix the corruption, and it also copies 50,000 "+
			"strings on every planning request. There is a fix that costs nothing "+
			"and involves typing one extra number. NOTES section 1, the last row of "+
			"the table.", got)
	}
}

func TestSplit_Boundaries(t *testing.T) {
	cases := []struct {
		name                     string
		orders                   []string
		count                    int
		wantMorning, wantEvening int
	}{
		{"all evening", day(), 0, 0, 4},
		{"all morning", day(), 4, 4, 0},
		{"over-long count is clamped", day(), 99, 4, 0},
		{"negative count is clamped", day(), -3, 0, 4},
		{"empty day", []string{}, 2, 0, 0},
		{"nil day", nil, 2, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			morning, evening := SplitShifts(tc.orders, tc.count)
			if len(morning) != tc.wantMorning || len(evening) != tc.wantEvening {
				t.Fatalf("got %d morning / %d evening, want %d / %d",
					len(morning), len(evening), tc.wantMorning, tc.wantEvening)
			}
		})
	}
}

// The stretch goal has its own test. It is expected to fail until you do the
// stretch goal, and it is fine to leave it failing — say so in MISTAKES.md if
// you skip it, so the record is honest.
func TestSplit_StretchGoal_AppendRushIsSafeFromEitherSide(t *testing.T) {
	t.Skip("stretch goal — remove this line when you attempt it")

	orders := day()
	// A caller who did NOT go through SplitShifts: a raw sub-slice with spare
	// capacity, exactly what an older version of the code handed out.
	morning := orders[:2]
	evening := orders[2:]

	morning = AppendRush(morning, "RUSH-9")
	_ = morning

	if want := []string{"WO-3", "WO-4"}; !reflect.DeepEqual(evening, want) {
		t.Fatalf("evening = %v, want %v — AppendRush must defend itself when handed "+
			"a slice with spare capacity it does not own", evening, want)
	}
}
