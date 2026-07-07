package reminder

import (
	"errors"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	return f.now
}

type fakeNotifier struct {
	count int
	err   error
}

func (f *fakeNotifier) SendReminder() error {
	f.count++
	return f.err
}

func TestControllerTestReminderDoesNotChangeSchedule(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	notifier := &fakeNotifier{}
	controller := NewController(newMemoryPrefs(nil), notifier, clock, nil)

	before := controller.NextDelay()
	if err := controller.TestReminder(); err != nil {
		t.Fatalf("TestReminder returned error: %v", err)
	}
	after := controller.NextDelay()

	if notifier.count != 1 {
		t.Fatalf("notification count = %d, want 1", notifier.count)
	}
	if after != before {
		t.Fatalf("NextDelay changed from %s to %s", before, after)
	}
}

func TestControllerPauseSuppressesTimerNotification(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	notifier := &fakeNotifier{}
	controller := NewController(newMemoryPrefs(nil), notifier, clock, nil)

	delay, err := controller.Pause(Pause30Minutes)
	if err != nil {
		t.Fatalf("Pause returned error: %v", err)
	}
	if delay != 30*time.Minute {
		t.Fatalf("pause delay = %s, want 30m", delay)
	}

	next, err := controller.HandleTimer()
	if err != nil {
		t.Fatalf("HandleTimer returned error: %v", err)
	}
	if notifier.count != 0 {
		t.Fatalf("notification count = %d, want 0", notifier.count)
	}
	if next != 30*time.Minute {
		t.Fatalf("next delay = %s, want 30m", next)
	}
}

func TestControllerTimerSendsNotificationWhenActive(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	notifier := &fakeNotifier{}
	controller := NewController(newMemoryPrefs(nil), notifier, clock, nil)

	next, err := controller.HandleTimer()
	if err != nil {
		t.Fatalf("HandleTimer returned error: %v", err)
	}
	if notifier.count != 1 {
		t.Fatalf("notification count = %d, want 1", notifier.count)
	}
	if next != 30*time.Minute {
		t.Fatalf("next delay = %s, want 30m", next)
	}
}

func TestControllerTimerReturnsNotificationError(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	notifier := &fakeNotifier{err: errors.New("blocked")}
	controller := NewController(newMemoryPrefs(nil), notifier, clock, nil)

	next, err := controller.HandleTimer()
	if err == nil {
		t.Fatal("expected notification error")
	}
	if next != 30*time.Minute {
		t.Fatalf("next delay = %s, want 30m", next)
	}
}

func TestControllerSetFrequencyPersistsAndResetsDelay(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	prefs := newMemoryPrefs(nil)
	controller := NewController(prefs, &fakeNotifier{}, &fakeClock{now: now}, nil)

	delay := controller.SetFrequency(15)
	if delay != 15*time.Minute {
		t.Fatalf("delay = %s, want 15m", delay)
	}
	if prefs.values[PreferenceKeyFrequencyMinutes] != 15 {
		t.Fatalf("stored frequency = %d, want 15", prefs.values[PreferenceKeyFrequencyMinutes])
	}
}

func TestClampDelay(t *testing.T) {
	if got := clampDelay(0); got != time.Second {
		t.Fatalf("clampDelay(0) = %s, want 1s", got)
	}
	if got := clampDelay(5 * time.Minute); got != 5*time.Minute {
		t.Fatalf("clampDelay(5m) = %s, want 5m", got)
	}
}
