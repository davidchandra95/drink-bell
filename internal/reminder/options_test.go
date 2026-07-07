package reminder

import (
	"testing"
	"time"
)

func TestNormalizeFrequencyMinutes(t *testing.T) {
	tests := []struct {
		name          string
		input         int
		wantMinutes   int
		wantCorrected bool
	}{
		{name: "missing zero", input: 0, wantMinutes: 30, wantCorrected: true},
		{name: "valid 15", input: 15, wantMinutes: 15, wantCorrected: false},
		{name: "valid 30", input: 30, wantMinutes: 30, wantCorrected: false},
		{name: "valid 60", input: 60, wantMinutes: 60, wantCorrected: false},
		{name: "invalid 45", input: 45, wantMinutes: 30, wantCorrected: true},
		{name: "negative", input: -1, wantMinutes: 30, wantCorrected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMinutes, gotCorrected := NormalizeFrequencyMinutes(tt.input)
			if gotMinutes != tt.wantMinutes {
				t.Fatalf("minutes = %d, want %d", gotMinutes, tt.wantMinutes)
			}
			if gotCorrected != tt.wantCorrected {
				t.Fatalf("corrected = %v, want %v", gotCorrected, tt.wantCorrected)
			}
		})
	}
}

func TestPauseUntil(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))

	tests := []struct {
		name   string
		option PauseOption
		want   time.Time
	}{
		{name: "30 minutes", option: Pause30Minutes, want: now.Add(30 * time.Minute)},
		{name: "1 hour", option: Pause1Hour, want: now.Add(time.Hour)},
		{name: "3 hours", option: Pause3Hours, want: now.Add(3 * time.Hour)},
		{
			name:   "until tomorrow",
			option: PauseUntilTomorrow,
			want:   time.Date(2026, 7, 8, 9, 0, 0, 0, now.Location()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PauseUntil(tt.option, now)
			if err != nil {
				t.Fatalf("PauseUntil returned error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("pauseUntil = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPauseUntilRejectsUnknownOption(t *testing.T) {
	_, err := PauseUntil(PauseOption("bad"), time.Now())
	if err == nil {
		t.Fatal("expected error for unknown pause option")
	}
}

func TestLabels(t *testing.T) {
	if got := FrequencyLabel(30); got != "30 mins" {
		t.Fatalf("FrequencyLabel(30) = %q, want %q", got, "30 mins")
	}
	if got := PauseLabel(Pause1Hour); got != "1 hour" {
		t.Fatalf("PauseLabel(Pause1Hour) = %q, want %q", got, "1 hour")
	}
}
