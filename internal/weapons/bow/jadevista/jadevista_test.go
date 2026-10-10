package jadevista

import "testing"

func TestStackBonuses(t *testing.T) {
	tests := []struct {
		refine  int
		wantEM  float64
		wantATK float64
	}{
		{refine: 1, wantEM: 64, wantATK: 0.12},
		{refine: 5, wantEM: 128, wantATK: 0.24},
	}

	for _, tt := range tests {
		em, atk := stackBonuses(tt.refine)
		if em != tt.wantEM || atk != tt.wantATK {
			t.Errorf("stackBonuses(%d) = (%v, %v), want (%v, %v)", tt.refine, em, atk, tt.wantEM, tt.wantATK)
		}
	}
}

func TestStackCountsPrioritizesSameElementAndCapsTotal(t *testing.T) {
	tests := []struct {
		name                    string
		same, different         int
		wantSame, wantDifferent int
	}{
		{name: "same element fills all stacks", same: 4, different: 2, wantSame: 3},
		{name: "different element uses remaining stacks", same: 1, different: 3, wantSame: 1, wantDifferent: 2},
		{name: "different element fills all stacks", different: 4, wantDifferent: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			same, different := stackCounts(tt.same, tt.different)
			if same != tt.wantSame || different != tt.wantDifferent {
				t.Fatalf("stackCounts() = (%d, %d), want (%d, %d)", same, different, tt.wantSame, tt.wantDifferent)
			}
			if same+different > 3 {
				t.Fatalf("stackCounts() returned %d total stacks, want at most 3", same+different)
			}
		})
	}
}
