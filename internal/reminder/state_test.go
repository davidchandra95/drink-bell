package reminder

import (
	"testing"
	"time"
)

type memoryPrefs struct {
	values map[string]int
}

func newMemoryPrefs(values map[string]int) *memoryPrefs {
	if values == nil {
		values = map[string]int{}
	}
	return &memoryPrefs{values: values}
}

func (m *memoryPrefs) IntWithFallback(key string, fallback int) int {
	value, ok := m.values[key]
	if !ok {
		return fallback
	}
	return value
}

func (m *memoryPrefs) SetInt(key string, value int) {
	m.values[key] = value
}

func (m *memoryPrefs) RemoveValue(key string) {
	delete(m.values, key)
}

func TestLoadStateDefaultsMissingFrequency(t *testing.T) {
	prefs := newMemoryPrefs(nil)
	state := LoadState(prefs, time.Now())

	if state.FrequencyMinutes != 30 {
		t.Fatalf("FrequencyMinutes = %d, want 30", state.FrequencyMinutes)
	}
	if prefs.values[PreferenceKeyFrequencyMinutes] != 30 {
		t.Fatalf("stored frequency = %d, want 30", prefs.values[PreferenceKeyFrequencyMinutes])
	}
}

func TestLoadStateCorrectsInvalidFrequency(t *testing.T) {
	prefs := newMemoryPrefs(map[string]int{
		PreferenceKeyFrequencyMinutes: 45,
	})
	state := LoadState(prefs, time.Now())

	if state.FrequencyMinutes != 30 {
		t.Fatalf("FrequencyMinutes = %d, want 30", state.FrequencyMinutes)
	}
	if prefs.values[PreferenceKeyFrequencyMinutes] != 30 {
		t.Fatalf("stored frequency = %d, want 30", prefs.values[PreferenceKeyFrequencyMinutes])
	}
}

func TestLoadStateHonorsFuturePause(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	prefs := newMemoryPrefs(map[string]int{
		PreferenceKeyFrequencyMinutes: 15,
		PreferenceKeyPauseUntilUnix:   int(future.Unix()),
	})

	state := LoadState(prefs, now)
	if state.PauseUntil == nil {
		t.Fatal("PauseUntil is nil, want future pause")
	}
	if !state.PauseUntil.Equal(future) {
		t.Fatalf("PauseUntil = %s, want %s", state.PauseUntil, future)
	}
}

func TestLoadStateClearsPastPause(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	prefs := newMemoryPrefs(map[string]int{
		PreferenceKeyFrequencyMinutes: 30,
		PreferenceKeyPauseUntilUnix:   int(past.Unix()),
	})

	state := LoadState(prefs, now)
	if state.PauseUntil != nil {
		t.Fatalf("PauseUntil = %s, want nil", state.PauseUntil)
	}
	if _, ok := prefs.values[PreferenceKeyPauseUntilUnix]; ok {
		t.Fatal("past pause timestamp was not cleared")
	}
}

func TestSaveFrequencyNormalizesInput(t *testing.T) {
	prefs := newMemoryPrefs(nil)
	got := SaveFrequency(prefs, 60)
	if got != 60 {
		t.Fatalf("SaveFrequency returned %d, want 60", got)
	}
	if prefs.values[PreferenceKeyFrequencyMinutes] != 60 {
		t.Fatalf("stored frequency = %d, want 60", prefs.values[PreferenceKeyFrequencyMinutes])
	}

	got = SaveFrequency(prefs, 45)
	if got != 30 {
		t.Fatalf("SaveFrequency returned %d, want 30", got)
	}
}
