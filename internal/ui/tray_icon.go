package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/tray-droplet.svg
var trayDropletSVG []byte

func trayIcon() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("tray-droplet.svg", trayDropletSVG))
}
