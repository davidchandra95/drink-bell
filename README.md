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

## Releases

Release artifacts are built by GitHub Actions when a version tag is pushed.

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The release workflow publishes unsigned desktop packages:

- `DrinkBell-darwin-arm64.zip`
- `DrinkBell-darwin-amd64.zip`
- `DrinkBell-windows-amd64.zip`
- `DrinkBell-linux-amd64.tar.xz`

These artifacts are unsigned for the first release setup. macOS may show Gatekeeper warnings and Windows may show SmartScreen warnings.

## Linux Tray Caveat

Linux desktop tray behavior varies by desktop environment and installed extensions.
Drink Bell still attempts to register a tray icon, but the fallback window is visible
by default on Linux and is the guaranteed control surface.

## Notification Permissions

Drink Bell uses native OS notifications. If `Test Reminder` appears to do nothing,
check the operating system notification settings for the app or terminal used to run it.
