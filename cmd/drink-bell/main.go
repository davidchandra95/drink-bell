package main

import (
	"github.com/davidchandra95/drink-bell/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	fyneApp := app.NewWithID("dev.slowtyper.drinkbell")
	ui.New(fyneApp).Run()
}
