package exercises

import (
	"reflect"
	"testing"
)

func mkOrders(spec ...string) []*WorkOrder {
	out := make([]*WorkOrder, len(spec))
	for i, s := range spec {
		out[i] = &WorkOrder{
			ID:     s,
			Status: map[bool]string{true: "ACTIVE", false: "DONE"}[s[0] == 'A'],
			Notes:  make([]byte, 4096), // fat on purpose: see TestCompact_DropsTheTail
		}
	}
	return out
}

func ids(orders []*WorkOrder) []string {
	out := make([]string, len(orders))
	for i, o := range orders {
		out[i] = o.ID
	}
	return out
}

func TestCompact_KeepsActiveInOrder(t *testing.T) {
	orders := mkOrders("A1", "D1", "A2", "D2", "A3")
	got := ids(CompactActive(orders))
	want := []string{"A1", "A2", "A3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CompactActive returned %v, want %v — keep the ACTIVE orders in their "+
			"original relative order", got, want)
	}
}

func TestCompact_DoesNotAllocate(t *testing.T) {
	orders := mkOrders("A1", "D1", "A2", "D2", "A3", "D3", "A4", "D4")
	backup := make([]*WorkOrder, len(orders))
	copy(backup, orders)

	got := testing.AllocsPerRun(100, func() {
		copy(orders, backup) // restore between runs; this is the test's cost, not yours
		_ = CompactActive(orders)
	})
	if got != 0 {
		t.Fatalf("CompactActive allocated %.0f time(s) per call — the budget is 0. "+
			"Building a second slice with append is the shape this exercise exists to "+
			"replace: filter in place, two indexes walking one array.", got)
	}
}

// The leak test. This is the one the exercise is really about.
func TestCompact_DropsTheTail(t *testing.T) {
	orders := mkOrders("A1", "D1", "A2", "D2", "A3")
	full := orders[:len(orders):len(orders)] // remember the whole array

	got := CompactActive(orders)
	if len(got) != 3 {
		t.Fatalf("CompactActive returned %d orders, want 3", len(got))
	}

	tail := full[len(got):]
	for i, o := range tail {
		if o != nil {
			t.Fatalf("backing-array slot %d (past the returned length) still points at "+
				"work order %q.\n\n"+
				"The slice you returned is short, but the array behind it is not. Every "+
				"dropped *WorkOrder is still reachable through that array, so the garbage "+
				"collector keeps it — and each one holds 4 KB of notes. Filter a 50,000-order "+
				"slice a few times a second and you have a leak that no amount of shrinking "+
				"fixes.\n\n"+
				"NOTES section 4, second half: `items = items[:0]` releases nothing.",
				len(got)+i, o.ID)
		}
	}
}

func TestCompact_BoringCases(t *testing.T) {
	cases := []struct {
		name  string
		input []*WorkOrder
		want  []string
	}{
		{"nil", nil, []string{}},
		{"empty", []*WorkOrder{}, []string{}},
		{"all active", mkOrders("A1", "A2"), []string{"A1", "A2"}},
		{"none active", mkOrders("D1", "D2"), []string{}},
		{"single active", mkOrders("A1"), []string{"A1"}},
		{"single done", mkOrders("D1"), []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ids(CompactActive(tc.input))
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
