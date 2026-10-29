package wintersheavyheart

import "testing"

func TestIsStellarGlimmerActive(t *testing.T) {
	tests := []struct {
		name           string
		stellarConduct bool
		stellarSwirl   bool
		want           bool
	}{
		{name: "neither status"},
		{name: "stellar conduct", stellarConduct: true, want: true},
		{name: "stellar swirl", stellarSwirl: true, want: true},
		{name: "both statuses", stellarConduct: true, stellarSwirl: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStellarGlimmerActive(tt.stellarConduct, tt.stellarSwirl); got != tt.want {
				t.Errorf("isStellarGlimmerActive(%t, %t) = %t, want %t",
					tt.stellarConduct, tt.stellarSwirl, got, tt.want)
			}
		})
	}
}
