package ui

import (
	"fmt"
	"runtime"
	"time"

	"github.com/davidchandra95/drink-bell/internal/reminder"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

type trayDesktop interface {
	SetSystemTrayIcon(fyne.Resource)
	SetSystemTrayMenu(*fyne.Menu)
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
	desk, ok := a.fyneApp.(trayDesktop)
	if !ok {
		return
	}

	installTrayMenu(desk, trayIcon(), a.buildTrayMenu())
}

func installTrayMenu(desk trayDesktop, icon fyne.Resource, menu *fyne.Menu) {
	desk.SetSystemTrayIcon(icon)
	desk.SetSystemTrayMenu(menu)
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
