package reminder

import (
	"fmt"
	"time"
)

const DefaultFrequencyMinutes = 30

var ValidFrequencyMinutes = []int{15, 30, 60}

type PauseOption string

const (
	Pause30Minutes     PauseOption = "30m"
	Pause1Hour         PauseOption = "1h"
	Pause3Hours        PauseOption = "3h"
	PauseUntilTomorrow PauseOption = "tomorrow"
)

var PauseOptions = []PauseOption{
	Pause30Minutes,
	Pause1Hour,
	Pause3Hours,
	PauseUntilTomorrow,
}

func NormalizeFrequencyMinutes(value int) (minutes int, corrected bool) {
	for _, valid := range ValidFrequencyMinutes {
		if value == valid {
			return value, false
		}
	}
	return DefaultFrequencyMinutes, true
}

func FrequencyDuration(minutes int) time.Duration {
	normalized, _ := NormalizeFrequencyMinutes(minutes)
	return time.Duration(normalized) * time.Minute
}

func FrequencyLabel(minutes int) string {
	normalized, _ := NormalizeFrequencyMinutes(minutes)
	return fmt.Sprintf("%d mins", normalized)
}

func PauseUntil(option PauseOption, now time.Time) (time.Time, error) {
	switch option {
	case Pause30Minutes:
		return now.Add(30 * time.Minute), nil
	case Pause1Hour:
		return now.Add(time.Hour), nil
	case Pause3Hours:
		return now.Add(3 * time.Hour), nil
	case PauseUntilTomorrow:
		tomorrow := now.AddDate(0, 0, 1)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, now.Location()), nil
	default:
		return time.Time{}, fmt.Errorf("unknown pause option %q", option)
	}
}

func PauseLabel(option PauseOption) string {
	switch option {
	case Pause30Minutes:
		return "30 mins"
	case Pause1Hour:
		return "1 hour"
	case Pause3Hours:
		return "3 hours"
	case PauseUntilTomorrow:
		return "Until tomorrow"
	default:
		return string(option)
	}
}
