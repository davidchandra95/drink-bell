package reminder

import "time"

const (
	PreferenceKeyFrequencyMinutes = "drinkbell.frequency_minutes"
	PreferenceKeyPauseUntilUnix   = "drinkbell.pause_until_unix"
)

type PreferenceStore interface {
	IntWithFallback(key string, fallback int) int
	SetInt(key string, value int)
	RemoveValue(key string)
}

type State struct {
	FrequencyMinutes int
	PauseUntil       *time.Time
}

func LoadState(store PreferenceStore, now time.Time) State {
	rawFrequency := store.IntWithFallback(PreferenceKeyFrequencyMinutes, DefaultFrequencyMinutes)
	frequency, _ := NormalizeFrequencyMinutes(rawFrequency)
	store.SetInt(PreferenceKeyFrequencyMinutes, frequency)

	rawPauseUntil := store.IntWithFallback(PreferenceKeyPauseUntilUnix, 0)
	var pauseUntil *time.Time
	if rawPauseUntil > 0 {
		loaded := time.Unix(int64(rawPauseUntil), 0)
		if loaded.After(now) {
			pauseUntil = &loaded
		} else {
			store.RemoveValue(PreferenceKeyPauseUntilUnix)
		}
	}

	return State{
		FrequencyMinutes: frequency,
		PauseUntil:       pauseUntil,
	}
}

func SaveFrequency(store PreferenceStore, minutes int) int {
	normalized, _ := NormalizeFrequencyMinutes(minutes)
	store.SetInt(PreferenceKeyFrequencyMinutes, normalized)
	return normalized
}

func SavePauseUntil(store PreferenceStore, until time.Time) {
	store.SetInt(PreferenceKeyPauseUntilUnix, int(until.Unix()))
}

func ClearPauseUntil(store PreferenceStore) {
	store.RemoveValue(PreferenceKeyPauseUntilUnix)
}
