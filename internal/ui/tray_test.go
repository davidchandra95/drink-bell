package ui

import (
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

type fakeTrayDesktop struct {
	icon  fyne.Resource
	menu  *fyne.Menu
	calls []string
}

func (f *fakeTrayDesktop) SetSystemTrayIcon(icon fyne.Resource) {
	f.icon = icon
	f.calls = append(f.calls, "icon")
}

func (f *fakeTrayDesktop) SetSystemTrayMenu(menu *fyne.Menu) {
	f.menu = menu
	f.calls = append(f.calls, "menu")
}

func TestInstallTrayMenuConfiguresMenuOnly(t *testing.T) {
	desk := &fakeTrayDesktop{}
	icon := fyne.NewStaticResource("icon.png", []byte{1})
	menu := fyne.NewMenu("Drink Bell")

	installTrayMenu(desk, icon, menu)

	if desk.icon == nil || desk.icon.Name() != icon.Name() {
		t.Fatalf("tray icon = %v, want %s", desk.icon, icon.Name())
	}
	if desk.menu != menu {
		t.Fatalf("tray menu = %p, want %p", desk.menu, menu)
	}
	if want := []string{"icon", "menu"}; !reflect.DeepEqual(desk.calls, want) {
		t.Fatalf("tray calls = %v, want %v", desk.calls, want)
	}
}

func TestTrayIconUsesDropletResource(t *testing.T) {
	icon := trayIcon()
	if icon == nil {
		t.Fatal("tray icon is nil")
	}
	if !strings.Contains(icon.Name(), "tray-droplet.svg") {
		t.Fatalf("tray icon name = %q, want droplet resource", icon.Name())
	}
}
