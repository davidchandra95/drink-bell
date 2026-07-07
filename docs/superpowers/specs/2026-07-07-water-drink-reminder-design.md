# Drink Bell Water Reminder Design

Date: 2026-07-07

## Summary

Drink Bell is a small cross-platform desktop utility written in Go. It reminds the user to drink water at a selected interval and exposes all controls from the system tray or menu bar. The app intentionally keeps the first version narrow:

- Send a test reminder.
- Select reminder frequency from fixed options.
- Pause reminders from fixed options.
- Quit the app.

The app targets macOS, Windows, and Linux. macOS and Windows should behave primarily as tray/menu-bar background apps. Linux should still register a tray icon when the desktop environment supports it, but it must also show a regular fallback control window so the app stays reachable even on desktops where tray/AppIndicator support is missing or hidden.

## Goals

- Provide a low-friction water reminder that runs in the background.
- Keep the interface small enough to fit naturally in a tray/menu-bar menu.
- Use native OS notifications for reliability and platform consistency.
- Persist the selected reminder frequency across restarts.
- Persist an active pause-until timestamp only while it is still in the future.
- Keep scheduling logic separate from UI code so the important behavior can be unit-tested.
- Use one Go desktop stack where possible instead of stitching together unrelated tray, notification, and fallback-window libraries.

## Non-Goals

- No accounts, backend, sync, cloud storage, or analytics.
- No custom reminder popup window in v1.
- No premium features, rate-app prompts, onboarding, randomized colors, sound effects, or launch-at-login controls.
- No user-defined custom interval input in v1.
- No mobile support.
- No complicated calendar-aware quiet hours.

## Technical Direction

Use Fyne as the single desktop application toolkit.

Rationale:

- Fyne supports Go-native desktop apps across macOS, Windows, and Linux.
- Fyne provides system tray/menu capabilities suitable for this app's primary control surface.
- Fyne provides native notification APIs, avoiding a separate notification dependency for v1.
- Fyne provides normal windows, which gives the Linux fallback UI without adding a second GUI stack.
- Fyne preferences give a platform-appropriate persistence mechanism without hand-rolled config paths.

Alternatives considered:

| Option | Pros | Cons | Decision |
|---|---|---|---|
| Fyne-only | One Go stack for tray, notifications, preferences, and fallback window. | Requires GUI build dependencies and cgo/toolchain setup on each OS. | Chosen for v1. |
| Wails | Strong tray/window story and future web UI flexibility. | Heavier than needed because this feature set does not need a web frontend. | Defer unless richer UI becomes important. |
| systray + beeep + separate window toolkit | Small focused libraries and direct control. | More integration risk and more ownership boundaries for a tiny app. | Avoid for v1. |

## User Experience

The primary interface is a tray/menu-bar menu similar to the reference app, but reduced to only the required controls.

Menu structure:

```text
Drink Bell
  Test Reminder
  ----------------
  Frequency
    15 mins
    30 mins   [default]
    60 mins
  ----------------
  Pause Reminders
    30 mins
    1 hour
    3 hours
    Until tomorrow
  ----------------
  Quit
```

The selected frequency should be visibly checked in the menu when the platform supports checkable menu items. If the toolkit or platform does not expose a reliable checked state in some context, the menu label can include the current value, for example `Frequency: 30 mins`.

Reminder notification copy:

```text
Take a sip now!
Stay hydrated!
```

`Test Reminder` sends the notification immediately. It must not reset or delay the normal reminder schedule.

The fallback window contains the same controls as the tray menu:

- Test reminder button.
- Frequency selector with 15, 30, and 60 minute options.
- Pause selector with 30 minutes, 1 hour, 3 hours, and until tomorrow.
- Current status text.
- Quit button.

On Linux, the fallback window should be visible by default. This is the conservative choice because tray visibility is not reliable across all Linux desktop environments. On macOS and Windows, the app can start hidden in the tray/menu bar by default.

## State Model

The app state is intentionally small:

```text
State:
  frequency       duration, default 30m, persisted
  pauseUntil      optional timestamp, persisted only if still future
  nextReminderAt  derived runtime value, not persisted
```

Persistence keys:

```text
drinkbell.frequency_minutes = 15 | 30 | 60
drinkbell.pause_until_unix   = unix timestamp seconds, optional
```

Validation rules:

- Missing frequency defaults to 30 minutes.
- Invalid frequency values default to 30 minutes and should be overwritten with the default.
- Missing pause timestamp means reminders are active.
- Past pause timestamps are ignored and cleared.
- Future pause timestamps are honored after restart.

## Runtime Architecture

```text
                    +-------------------+
                    |     Fyne App      |
                    +---------+---------+
                              |
           +------------------+------------------+
           |                                     |
+----------v-----------+             +-----------v----------+
| Tray/Menu Controller |             | Fallback Window      |
|                      |             | Controller           |
| - Test Reminder      |             |                      |
| - Frequency options  |             | - Same commands      |
| - Pause options      |             | - Current status     |
| - Quit               |             | - Linux visibility   |
+----------+-----------+             +-----------+----------+
           |                                     |
           +------------------+------------------+
                              |
                    +---------v---------+
                    | App State Manager |
                    |                  |
                    | - frequency      |
                    | - pauseUntil     |
                    | - persistence    |
                    +---------+---------+
                              |
                    +---------v---------+
                    | Reminder Scheduler|
                    |                  |
                    | - timer lifecycle|
                    | - pause handling |
                    | - next reminder  |
                    +---------+---------+
                              |
                    +---------v---------+
                    | Notification     |
                    | Sender           |
                    +------------------+
```

Ownership boundaries:

- UI controllers own rendering menu/window controls and translating user actions into commands.
- The app state manager owns validation, persistence, and the canonical current settings.
- The scheduler owns timers and decides when a reminder should fire.
- The notification sender owns the OS notification call.

The scheduler must not directly depend on tray or window widgets. This keeps reminder behavior testable without creating a GUI.

## Data Flows

### Startup

```text
App starts
  -> load preferences
  -> validate frequency
  -> validate pauseUntil
  -> create scheduler state
  -> create tray/menu
  -> create fallback window
  -> Linux: show fallback window
  -> macOS/Windows: keep fallback window hidden unless opened by user
  -> schedule next reminder
```

### Frequency Change

```text
User selects frequency
  -> validate selected option
  -> update state.frequency
  -> persist frequency
  -> clear current timer
  -> schedule next reminder at now + frequency
  -> refresh menu/window selected state
```

Changing frequency should restart the schedule from the time of change. This is predictable and avoids surprising immediate reminders after a user shortens the interval.

### Pause Reminders

```text
User selects pause option
  -> compute pauseUntil
  -> update state.pauseUntil
  -> persist pauseUntil
  -> clear current timer
  -> schedule next evaluation at pauseUntil
  -> refresh status
```

Pause options:

| Option | Behavior |
|---|---|
| 30 mins | Pause until now + 30 minutes. |
| 1 hour | Pause until now + 1 hour. |
| 3 hours | Pause until now + 3 hours. |
| Until tomorrow | Pause until the next local calendar day at 09:00. |

The `until tomorrow` rule needs to be explicit. Using 09:00 local time is better than midnight because midnight can create an unexpected reminder while the user is sleeping.

### Timer Fires

```text
Timer fires
  -> if pauseUntil is future:
       schedule next evaluation at pauseUntil
       do not notify
  -> else:
       clear expired pauseUntil
       send notification
       schedule next reminder at now + frequency
```

### Test Reminder

```text
User selects Test Reminder
  -> send notification immediately
  -> update status text
  -> leave scheduler unchanged
```

Test reminders are diagnostic. They should not affect the real interval timer.

## Error Handling

The app should fail soft. A water reminder utility should not crash because of a blocked notification or invalid persisted setting.

Expected handling:

- Invalid persisted frequency: use 30 minutes and save the corrected value.
- Expired pause timestamp: clear it and continue normally.
- Notification blocked by OS settings: keep app running and update fallback window status if possible.
- Tray not visible on Linux: fallback window remains visible and usable.
- Timer reset race during rapid menu clicks: stop old timer before scheduling the next one; route scheduler changes through one owner to avoid overlapping timers.

If Fyne's notification API does not expose detailed delivery failures on a platform, the notification sender should return best-effort status. The UI can display neutral text such as:

```text
Reminder sent. If you did not see it, check OS notification permissions.
```

## Observability

This app does not need heavy observability. Still, basic structured-ish logs are useful during local debugging and issue reports.

Log events:

- App startup with OS and selected frequency.
- Frequency changed.
- Pause selected with pauseUntil.
- Reminder notification attempted.
- Scheduler rescheduled.
- Invalid preference corrected.

Logs should avoid noisy per-second timer output.

## Packaging And Platform Notes

The first implementation should make the app runnable locally with:

```text
go run ./cmd/drink-bell
```

Packaging should be documented separately after the app works locally:

- macOS app bundle.
- Windows executable.
- Linux desktop app packaging.

Linux caveat:

```text
Linux desktop environment
  |
  +-- Tray/AppIndicator supported
  |     -> tray icon may be visible
  |     -> fallback window still available
  |
  +-- Tray/AppIndicator unsupported or hidden
        -> fallback window is the guaranteed control surface
```

Because Linux tray behavior varies by desktop environment and installed extensions, the fallback window is part of the product requirement, not a debugging workaround.

## Test Plan

Unit tests should focus on behavior that can break silently:

| Area | Test case | Expected behavior |
|---|---|---|
| Frequency validation | Missing value | Defaults to 30 minutes. |
| Frequency validation | Valid 15/30/60 | Accepted without modification. |
| Frequency validation | Invalid value | Defaults to 30 minutes and reports correction. |
| Pause calculation | 30 mins / 1 hour / 3 hours | Computes the right future timestamp. |
| Pause calculation | Until tomorrow | Computes next local calendar day at 09:00. |
| Pause validation | Past timestamp | Clears pause and resumes active schedule. |
| Scheduler | Active timer fires | Sends one notification and schedules next interval. |
| Scheduler | Paused timer fires | Does not notify; schedules evaluation at pauseUntil. |
| Scheduler | Test reminder | Sends immediate notification and leaves schedule unchanged. |

Manual verification:

- Run locally on macOS.
- Confirm tray/menu-bar icon appears.
- Confirm `Test Reminder` triggers a native notification.
- Confirm frequency selection changes checked/current state.
- Confirm pause changes status and suppresses normal reminders.
- Confirm app quits cleanly.
- On Linux, confirm fallback window is visible even if tray visibility is unavailable.

## Acceptance Criteria

- App is implemented in Go using Fyne.
- macOS and Windows expose tray/menu-bar controls.
- Linux exposes a visible fallback control window by default.
- Frequency options are exactly 15, 30, and 60 minutes for v1.
- Default frequency is 30 minutes.
- Pause options are exactly 30 minutes, 1 hour, 3 hours, and until tomorrow for v1.
- `Test Reminder` sends an immediate native OS notification.
- Selected frequency persists across restarts.
- Future pause state survives restart; expired pause state is cleared.
- Scheduler and preference behavior have unit tests.
- No backend, account, analytics, or extra product features are added.

## Open Risks

- Native notification behavior can depend on OS permissions and desktop notification settings.
- Linux tray visibility cannot be guaranteed across all desktop environments; the fallback window satisfies the control-surface requirement.
- Cross-platform packaging may expose toolchain-specific issues and should be treated as a separate hardening task after local app behavior is correct.
