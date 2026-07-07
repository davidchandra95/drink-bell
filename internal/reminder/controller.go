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
