# Drink Bell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. This project is small and tightly coupled, so inline execution is the default recommendation.

**Goal:** Build a cross-platform Go tray/menu-bar water reminder app with test reminder, fixed frequency options, fixed pause options, native OS notifications, persisted settings, and a guaranteed Linux fallback window.

**Architecture:** The app uses Fyne as the only desktop toolkit. Pure reminder behavior lives in `internal/reminder` and has no Fyne dependency. Fyne-specific tray, notification, preference, and fallback-window wiring lives in `internal/ui`, with `cmd/drink-bell` as the binary entrypoint.

**Tech Stack:** Go 1.26.3 locally, Fyne v2.7.4, standard library `testing`, standard library `log/slog`, native Fyne preferences and notifications.

## Global Constraints

- Do not stage, commit, push, create branches, or open pull requests unless David explicitly asks.
- Review work with tests and working-tree evidence; this folder is currently not a git repo.
- App is implemented in Go using Fyne.
- macOS and Windows expose tray/menu-bar controls.
- Linux exposes a visible fallback control window by default.
- Frequency options are exactly 15, 30, and 60 minutes for v1.
- Default frequency is 30 minutes.
- Pause options are exactly 30 minutes, 1 hour, 3 hours, and until tomorrow for v1.
- `Test Reminder` sends an immediate native OS notification.
- Selected frequency persists across restarts.
- Future pause state survives restart; expired pause state is cleared.
- No backend, account, analytics, premium, onboarding, sound effects, launch-at-login, or custom popup reminder window.
- `until tomorrow` means the next local calendar day at 09:00.

---

## File Structure

```text
drink-bell/
  go.mod
  go.sum
  cmd/
    drink-bell/
      main.go
  internal/
    reminder/
      options.go
      options_test.go
      state.go
      state_test.go
      controller.go
      controller_test.go
    ui/
      app.go
  docs/
    superpowers/
      specs/
        2026-07-07-water-drink-reminder-design.md
      plans/
        2026-07-07-drink-bell-implementation.md
```

Ownership:

```text
cmd/drink-bell/main.go
  -> creates the Fyne app and hands control to internal/ui

internal/reminder/options.go
  -> fixed frequency and pause option definitions

internal/reminder/state.go
  -> preference keys, state loading, state persistence validation

internal/reminder/controller.go
  -> command handling, scheduling decisions, notification trigger points

internal/ui/app.go
  -> Fyne tray/menu, fallback window, Fyne preference adapter, Fyne notification adapter
```

The dependency direction must stay one-way:

```text
cmd/drink-bell
  -> internal/ui
      -> internal/reminder

internal/reminder
  -> standard library only
```

---

### Task 1: Initialize Go Module And Fyne Dependency

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/go.mod`
- Create: `/Users/slowtyper/code/drink-bell/go.sum`

**Interfaces:**
- Produces: Go module path `drink-bell`
- Produces: Fyne dependency `fyne.io/fyne/v2 v2.7.4`

- [ ] **Step 1: Initialize the module**

Run:

```bash
go mod init drink-bell
```

Expected:

```text
go: creating new go.mod: module drink-bell
```

- [ ] **Step 2: Add Fyne**

Run:

```bash
go get fyne.io/fyne/v2@v2.7.4
```

Expected:

```text
go: added fyne.io/fyne/v2 v2.7.4
```

The exact output may include additional indirect modules. That is normal for Fyne.

- [ ] **Step 3: Tidy module state**

Run:

```bash
go mod tidy
```

Expected:

```text
go: warning: "all" matched no packages
```

This warning is expected because no Go packages exist yet.

- [ ] **Step 4: Verify module file**

`/Users/slowtyper/code/drink-bell/go.mod` should contain at least:

```go
module drink-bell

go 1.26.3
```

If `go mod init` writes `go 1.26`, keep the generated value. Do not manually force a patch version if the tool omits it.

- [ ] **Step 5: Review checkpoint**

Run:

```bash
go env GOMOD
```

Expected:

```text
/Users/slowtyper/code/drink-bell/go.mod
```

---

### Task 2: Add Frequency And Pause Option Domain Logic

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/internal/reminder/options.go`
- Create: `/Users/slowtyper/code/drink-bell/internal/reminder/options_test.go`

**Interfaces:**
- Produces: `const DefaultFrequencyMinutes = 30`
- Produces: `var ValidFrequencyMinutes = []int{15, 30, 60}`
- Produces: `func NormalizeFrequencyMinutes(value int) (minutes int, corrected bool)`
- Produces: `type PauseOption string`
- Produces: `func PauseUntil(option PauseOption, now time.Time) (time.Time, error)`
- Produces: `func FrequencyLabel(minutes int) string`
- Produces: `func PauseLabel(option PauseOption) string`

- [ ] **Step 1: Write failing option tests**

Create `/Users/slowtyper/code/drink-bell/internal/reminder/options_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/reminder
```

Expected:

```text
undefined: NormalizeFrequencyMinutes
```

The compiler may list more missing names from the same test file.

- [ ] **Step 3: Implement option logic**

Create `/Users/slowtyper/code/drink-bell/internal/reminder/options.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/reminder
```

Expected:

```text
ok  	drink-bell/internal/reminder
```

---

### Task 3: Add Preference-Backed State Loading

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/internal/reminder/state.go`
- Create: `/Users/slowtyper/code/drink-bell/internal/reminder/state_test.go`

**Interfaces:**
- Consumes: `NormalizeFrequencyMinutes`, `DefaultFrequencyMinutes`
- Produces: `type PreferenceStore interface`
- Produces: `type State struct`
- Produces: `func LoadState(store PreferenceStore, now time.Time) State`
- Produces: `func SaveFrequency(store PreferenceStore, minutes int) int`
- Produces: `func SavePauseUntil(store PreferenceStore, until time.Time)`
- Produces: `func ClearPauseUntil(store PreferenceStore)`

- [ ] **Step 1: Write failing state tests**

Create `/Users/slowtyper/code/drink-bell/internal/reminder/state_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/reminder
```

Expected:

```text
undefined: LoadState
```

- [ ] **Step 3: Implement preference-backed state**

Create `/Users/slowtyper/code/drink-bell/internal/reminder/state.go`:

```go
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
	frequency, corrected := NormalizeFrequencyMinutes(rawFrequency)
	if corrected {
		store.SetInt(PreferenceKeyFrequencyMinutes, frequency)
	}

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
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/reminder
```

Expected:

```text
ok  	drink-bell/internal/reminder
```

---

### Task 4: Add Controller And Scheduler Decisions

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/internal/reminder/controller.go`
- Create: `/Users/slowtyper/code/drink-bell/internal/reminder/controller_test.go`

**Interfaces:**
- Consumes: `State`, `PreferenceStore`, `PauseOption`, `PauseUntil`, `FrequencyDuration`
- Produces: `type Clock interface`
- Produces: `type NotificationSender interface`
- Produces: `type Controller struct`
- Produces: `func NewController(store PreferenceStore, notifier NotificationSender, clock Clock, onChange func(Snapshot)) *Controller`
- Produces: `func (c *Controller) SetFrequency(minutes int) time.Duration`
- Produces: `func (c *Controller) Pause(option PauseOption) (time.Duration, error)`
- Produces: `func (c *Controller) TestReminder() error`
- Produces: `func (c *Controller) HandleTimer() (time.Duration, error)`
- Produces: `func (c *Controller) NextDelay() time.Duration`
- Produces: `type Scheduler struct`
- Produces: `func NewScheduler(controller *Controller) *Scheduler`

- [ ] **Step 1: Write failing controller tests**

Create `/Users/slowtyper/code/drink-bell/internal/reminder/controller_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/reminder
```

Expected:

```text
undefined: NewController
```

- [ ] **Step 3: Implement controller and scheduler**

Create `/Users/slowtyper/code/drink-bell/internal/reminder/controller.go`:

```go
package reminder

import (
	"log/slog"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type NotificationSender interface {
	SendReminder() error
}

type Snapshot struct {
	FrequencyMinutes int
	PauseUntil       *time.Time
	Status           string
	NextDelay        time.Duration
}

type Controller struct {
	mu       sync.Mutex
	store    PreferenceStore
	notifier NotificationSender
	clock    Clock
	state    State
	status   string
	onChange func(Snapshot)
}

func NewController(store PreferenceStore, notifier NotificationSender, clock Clock, onChange func(Snapshot)) *Controller {
	if clock == nil {
		clock = RealClock{}
	}

	controller := &Controller{
		store:    store,
		notifier: notifier,
		clock:    clock,
		state:    LoadState(store, clock.Now()),
		status:   "Ready",
		onChange: onChange,
	}
	controller.emitChangeLocked()
	return controller
}

func (c *Controller) SetFrequency(minutes int) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	normalized := SaveFrequency(c.store, minutes)
	c.state.FrequencyMinutes = normalized
	c.status = "Frequency set to " + FrequencyLabel(normalized)
	slog.Info("frequency changed", "minutes", normalized)
	c.emitChangeLocked()
	return FrequencyDuration(normalized)
}

func (c *Controller) Pause(option PauseOption) (time.Duration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	until, err := PauseUntil(option, c.clock.Now())
	if err != nil {
		return c.nextDelayLocked(), err
	}

	c.state.PauseUntil = &until
	SavePauseUntil(c.store, until)
	c.status = "Paused until " + until.Format("Jan 2 15:04")
	slog.Info("pause selected", "option", option, "pause_until", until)
	c.emitChangeLocked()
	return c.nextDelayLocked(), nil
}

func (c *Controller) TestReminder() error {
	err := c.notifier.SendReminder()

	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.status = "Reminder attempted. Check OS notification permissions."
	} else {
		c.status = "Reminder sent"
	}
	c.emitChangeLocked()
	return err
}

func (c *Controller) HandleTimer() (time.Duration, error) {
	c.mu.Lock()
	if c.isPausedLocked() {
		next := c.nextDelayLocked()
		c.emitChangeLocked()
		c.mu.Unlock()
		return next, nil
	}
	c.clearExpiredPauseLocked()
	c.mu.Unlock()

	err := c.notifier.SendReminder()

	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.status = "Reminder attempted. Check OS notification permissions."
	} else {
		c.status = "Reminder sent"
	}
	next := FrequencyDuration(c.state.FrequencyMinutes)
	c.emitChangeLocked()
	return next, err
}

func (c *Controller) NextDelay() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nextDelayLocked()
}

func (c *Controller) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snapshotLocked()
}

func (c *Controller) isPausedLocked() bool {
	return c.state.PauseUntil != nil && c.state.PauseUntil.After(c.clock.Now())
}

func (c *Controller) clearExpiredPauseLocked() {
	if c.state.PauseUntil == nil {
		return
	}
	if c.state.PauseUntil.After(c.clock.Now()) {
		return
	}
	c.state.PauseUntil = nil
	ClearPauseUntil(c.store)
}

func (c *Controller) nextDelayLocked() time.Duration {
	c.clearExpiredPauseLocked()
	if c.state.PauseUntil != nil {
		return c.state.PauseUntil.Sub(c.clock.Now())
	}
	return FrequencyDuration(c.state.FrequencyMinutes)
}

func (c *Controller) snapshotLocked() Snapshot {
	return Snapshot{
		FrequencyMinutes: c.state.FrequencyMinutes,
		PauseUntil:       c.state.PauseUntil,
		Status:           c.status,
		NextDelay:        c.nextDelayLocked(),
	}
}

func (c *Controller) emitChangeLocked() {
	if c.onChange == nil {
		return
	}
	c.onChange(c.snapshotLocked())
}

type Scheduler struct {
	controller *Controller
	resetCh    chan struct{}
	stopCh     chan struct{}
	doneCh     chan struct{}
	startOnce  sync.Once
	stopOnce   sync.Once
}

func NewScheduler(controller *Controller) *Scheduler {
	return &Scheduler{
		controller: controller,
		resetCh:    make(chan struct{}, 1),
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.startOnce.Do(func() {
		go s.loop()
	})
}

func (s *Scheduler) Reset() {
	select {
	case s.resetCh <- struct{}{}:
	default:
	}
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		<-s.doneCh
	})
}

func (s *Scheduler) loop() {
	defer close(s.doneCh)

	timer := time.NewTimer(clampDelay(s.controller.NextDelay()))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			next, err := s.controller.HandleTimer()
			if err != nil {
				slog.Warn("reminder notification failed", "error", err)
			}
			resetTimer(timer, clampDelay(next))
		case <-s.resetCh:
			resetTimer(timer, clampDelay(s.controller.NextDelay()))
		case <-s.stopCh:
			return
		}
	}
}

func resetTimer(timer *time.Timer, delay time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(delay)
}

func clampDelay(delay time.Duration) time.Duration {
	if delay < time.Second {
		return time.Second
	}
	return delay
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/reminder
```

Expected:

```text
ok  	drink-bell/internal/reminder
```

---

### Task 5: Add Fyne Tray, Fallback Window, And Entrypoint

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/internal/ui/app.go`
- Create: `/Users/slowtyper/code/drink-bell/cmd/drink-bell/main.go`

**Interfaces:**
- Consumes: `reminder.Controller`, `reminder.Scheduler`, `reminder.PreferenceStore`, `reminder.NotificationSender`
- Produces: `func New(fyneApp fyne.App) *App`
- Produces: `func (a *App) Run()`
- Produces: runnable app command `go run ./cmd/drink-bell`

- [ ] **Step 1: Create the Fyne UI wrapper**

Create `/Users/slowtyper/code/drink-bell/internal/ui/app.go`:

```go
package ui

import (
	"fmt"
	"runtime"
	"time"

	"drink-bell/internal/reminder"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	controller *reminder.Controller
	scheduler  *reminder.Scheduler

	statusLabel *widget.Label
	frequency   *widget.Select
	pause       *widget.Select
	updatingUI  bool
}

type notifier struct {
	app fyne.App
}

func (n notifier) SendReminder() error {
	n.app.SendNotification(fyne.NewNotification("Take a sip now!", "Stay hydrated!"))
	return nil
}

func New(fyneApp fyne.App) *App {
	a := &App{
		fyneApp: fyneApp,
	}

	a.controller = reminder.NewController(
		fyneApp.Preferences(),
		notifier{app: fyneApp},
		reminder.RealClock{},
		a.applySnapshot,
	)
	a.scheduler = reminder.NewScheduler(a.controller)
	a.window = fyneApp.NewWindow("Drink Bell")
	a.window.Resize(fyne.NewSize(360, 260))
	a.window.SetCloseIntercept(func() {
		a.window.Hide()
	})
	a.window.SetContent(a.buildWindowContent())
	a.installTray()
	a.applySnapshot(a.controller.Snapshot())

	return a
}

func (a *App) Run() {
	a.scheduler.Start()
	if runtime.GOOS == "linux" {
		a.window.Show()
	} else {
		a.window.Hide()
	}
	a.fyneApp.Run()
	a.scheduler.Stop()
}

func (a *App) buildWindowContent() fyne.CanvasObject {
	a.statusLabel = widget.NewLabel("Ready")
	a.statusLabel.Wrapping = fyne.TextWrapWord

	a.frequency = widget.NewSelect(frequencyLabels(), func(label string) {
		if a.updatingUI {
			return
		}
		minutes := minutesFromFrequencyLabel(label)
		a.controller.SetFrequency(minutes)
		a.scheduler.Reset()
		a.installTray()
	})

	a.pause = widget.NewSelect(pauseLabels(), func(label string) {
		if a.updatingUI {
			return
		}
		option := pauseOptionFromLabel(label)
		if _, err := a.controller.Pause(option); err != nil {
			a.statusLabel.SetText("Could not pause reminders")
			return
		}
		a.scheduler.Reset()
		a.installTray()
	})
	a.pause.PlaceHolder = "Choose pause duration"

	testButton := widget.NewButtonWithIcon("Test Reminder", theme.MailSendIcon(), func() {
		_ = a.controller.TestReminder()
	})

	quitButton := widget.NewButtonWithIcon("Quit", theme.CancelIcon(), func() {
		a.scheduler.Stop()
		a.fyneApp.Quit()
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Drink Bell", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.statusLabel,
		widget.NewSeparator(),
		widget.NewLabel("Frequency"),
		a.frequency,
		widget.NewLabel("Pause reminders"),
		a.pause,
		widget.NewSeparator(),
		container.NewHBox(testButton, quitButton),
	)
}

func (a *App) installTray() {
	desk, ok := a.fyneApp.(desktop.App)
	if !ok {
		return
	}

	desk.SetSystemTrayIcon(theme.InfoIcon())
	desk.SetSystemTrayWindow(a.window)
	desk.SetSystemTrayMenu(a.buildTrayMenu())
}

func (a *App) buildTrayMenu() *fyne.Menu {
	current := a.controller.Snapshot()

	frequencyMenu := fyne.NewMenu("Frequency")
	for _, minutes := range reminder.ValidFrequencyMinutes {
		minutes := minutes
		item := fyne.NewMenuItem(reminder.FrequencyLabel(minutes), func() {
			a.controller.SetFrequency(minutes)
			a.scheduler.Reset()
			a.installTray()
		})
		item.Checked = current.FrequencyMinutes == minutes
		frequencyMenu.Items = append(frequencyMenu.Items, item)
	}

	pauseMenu := fyne.NewMenu("Pause Reminders")
	for _, option := range reminder.PauseOptions {
		option := option
		pauseMenu.Items = append(pauseMenu.Items, fyne.NewMenuItem(reminder.PauseLabel(option), func() {
			if _, err := a.controller.Pause(option); err != nil {
				a.statusLabel.SetText("Could not pause reminders")
				return
			}
			a.scheduler.Reset()
			a.installTray()
		}))
	}

	return fyne.NewMenu("Drink Bell",
		fyne.NewMenuItem("Show Controls", func() {
			a.window.Show()
			a.window.RequestFocus()
		}),
		fyne.NewMenuItem("Test Reminder", func() {
			_ = a.controller.TestReminder()
		}),
		fyne.NewMenuItemSeparator(),
		&fyne.MenuItem{Label: fmt.Sprintf("Frequency: %s", reminder.FrequencyLabel(current.FrequencyMinutes)), ChildMenu: frequencyMenu},
		&fyne.MenuItem{Label: "Pause Reminders", ChildMenu: pauseMenu},
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			a.scheduler.Stop()
			a.fyneApp.Quit()
		}),
	)
}

func (a *App) applySnapshot(snapshot reminder.Snapshot) {
	fyne.Do(func() {
		a.updatingUI = true
		defer func() {
			a.updatingUI = false
		}()

		if a.statusLabel != nil {
			a.statusLabel.SetText(statusText(snapshot))
		}
		if a.frequency != nil {
			a.frequency.SetSelected(reminder.FrequencyLabel(snapshot.FrequencyMinutes))
		}
	})
}

func statusText(snapshot reminder.Snapshot) string {
	if snapshot.PauseUntil != nil && snapshot.PauseUntil.After(time.Now()) {
		return "Paused until " + snapshot.PauseUntil.Format("Jan 2 15:04")
	}
	if snapshot.Status != "" {
		return snapshot.Status
	}
	return "Next reminder in " + snapshot.NextDelay.Round(time.Second).String()
}

func frequencyLabels() []string {
	labels := make([]string, 0, len(reminder.ValidFrequencyMinutes))
	for _, minutes := range reminder.ValidFrequencyMinutes {
		labels = append(labels, reminder.FrequencyLabel(minutes))
	}
	return labels
}

func minutesFromFrequencyLabel(label string) int {
	for _, minutes := range reminder.ValidFrequencyMinutes {
		if reminder.FrequencyLabel(minutes) == label {
			return minutes
		}
	}
	return reminder.DefaultFrequencyMinutes
}

func pauseLabels() []string {
	labels := make([]string, 0, len(reminder.PauseOptions))
	for _, option := range reminder.PauseOptions {
		labels = append(labels, reminder.PauseLabel(option))
	}
	return labels
}

func pauseOptionFromLabel(label string) reminder.PauseOption {
	for _, option := range reminder.PauseOptions {
		if reminder.PauseLabel(option) == label {
			return option
		}
	}
	return reminder.Pause30Minutes
}
```

- [ ] **Step 2: Create the binary entrypoint**

Create `/Users/slowtyper/code/drink-bell/cmd/drink-bell/main.go`:

```go
package main

import (
	"drink-bell/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	fyneApp := app.NewWithID("dev.slowtyper.drinkbell")
	ui.New(fyneApp).Run()
}
```

- [ ] **Step 3: Tidy module state**

Run:

```bash
go mod tidy
```

Expected:

```text
```

No output is expected on success. Dependency downloads may print module download lines on first run.

- [ ] **Step 4: Run all tests**

Run:

```bash
go test ./...
```

Expected:

```text
ok  	drink-bell/internal/reminder
```

Packages without tests may show `?`.

- [ ] **Step 5: Build the binary**

Run:

```bash
go build ./cmd/drink-bell
```

Expected:

```text
```

No output is expected on success. The build creates `/Users/slowtyper/code/drink-bell/drink-bell`.

- [ ] **Step 6: Manual app verification**

Run:

```bash
go run ./cmd/drink-bell
```

Expected manual behavior on macOS:

```text
- The app starts without terminal errors.
- A tray/menu-bar icon is available.
- The tray menu includes Show Controls, Test Reminder, Frequency, Pause Reminders, and Quit.
- Test Reminder triggers a native notification or creates an OS permission prompt.
- Frequency shows 30 mins as the default selected option.
- Selecting 15 mins updates the fallback window status and tray checked state.
- Selecting Pause Reminders -> 30 mins suppresses normal reminder delivery until the pause expires.
- Quit exits the process cleanly.
```

Linux manual behavior to verify on a Linux desktop:

```text
- The fallback window is visible by default.
- The same controls are reachable from the window even when no tray icon is visible.
```

---

### Task 6: Add Minimal Project Documentation

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/README.md`

**Interfaces:**
- Consumes: app command `go run ./cmd/drink-bell`
- Produces: local run instructions and platform caveats

- [ ] **Step 1: Write README**

Create `/Users/slowtyper/code/drink-bell/README.md`:

```markdown
# Drink Bell

Drink Bell is a small Go desktop app that reminds you to drink water.

## Features

- Test reminder.
- Fixed reminder frequencies: 15 mins, 30 mins, 60 mins.
- Fixed pause options: 30 mins, 1 hour, 3 hours, until tomorrow.
- Native OS notifications.
- Tray/menu-bar controls on macOS, Windows, and supported Linux desktops.
- Visible Linux fallback window so the app stays controllable without tray support.

## Run Locally

```bash
go run ./cmd/drink-bell
```

## Test

```bash
go test ./...
```

## Build

```bash
go build ./cmd/drink-bell
```

## Linux Tray Caveat

Linux desktop tray behavior varies by desktop environment and installed extensions.
Drink Bell still attempts to register a tray icon, but the fallback window is visible
by default on Linux and is the guaranteed control surface.

## Notification Permissions

Drink Bell uses native OS notifications. If `Test Reminder` appears to do nothing,
check the operating system notification settings for the app or terminal used to run it.
```

- [ ] **Step 2: Run formatting and tests**

Run:

```bash
gofmt -w cmd internal
go test ./...
```

Expected:

```text
ok  	drink-bell/internal/reminder
```

- [ ] **Step 3: Final working-tree review**

Run:

```bash
find . -maxdepth 4 -type f | sort
```

Expected file list includes:

```text
./README.md
./cmd/drink-bell/main.go
./docs/superpowers/plans/2026-07-07-drink-bell-implementation.md
./docs/superpowers/specs/2026-07-07-water-drink-reminder-design.md
./go.mod
./go.sum
./internal/reminder/controller.go
./internal/reminder/controller_test.go
./internal/reminder/options.go
./internal/reminder/options_test.go
./internal/reminder/state.go
./internal/reminder/state_test.go
./internal/ui/app.go
```

If David initializes git before execution, also run:

```bash
git status --short
git diff --stat HEAD
git diff HEAD
```

Do not stage or commit unless David explicitly asks.

---

## Self-Review Checklist

- Spec coverage: all v1 features map to tasks 2 through 6.
- Tray/menu-bar primary UX: task 5.
- Native OS notification: task 5 `notifier.SendReminder`.
- Frequency options and default: tasks 2, 3, 5.
- Pause options and `until tomorrow` at 09:00: tasks 2, 4, 5.
- Persisted frequency and future pause: task 3.
- Expired pause clearing: task 3.
- Linux fallback window: task 5.
- Unit tests for scheduler decisions and preferences: tasks 3 and 4.
- No backend, sync, accounts, analytics, custom popup, premium, sounds, or login item: preserved by file structure and README.
- Git policy: every task avoids staging and committing.
