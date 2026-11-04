package main

import (
	_ "embed"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/icon.png
var appIcon []byte

// trayIcon is the tray's cut of the mark, 32 px for 16 points at 2x.
//
//go:embed resources/tray.png
var trayIcon []byte

// desktop holds the windows and the tray, on the main thread.
type desktop struct {
	app      *App
	main     *mygo.Window
	flyout   *mygo.Window
	overlay  *mygo.Window
	tray     *mygo.Tray
	quitting bool
	hotkey   string
}

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return // the running one shows itself (OnSecondInstance below)
	}
	store := NewStore(dataDir())
	data, err := store.Load()
	if err != nil {
		log.Println("ctrl-tower: loading data:", err)
	}
	d := &desktop{}
	a := NewApp(store, data, hooks{
		notify:       d.notify,
		openURL:      func(u string) { _ = mygo.Shell.OpenExternal(u) },
		invalidate:   d.invalidate,
		showOverlay:  d.showOverlay,
		closeOverlay: d.closeOverlay,
		showMain:     d.showMain,
		setLogin: func(on bool) {
			if err := mygo.App.SetOpenAtLogin(on); err != nil {
				log.Println("ctrl-tower: start at login:", err)
			}
		},
		setHotkey: d.setHotkey,
	})
	a.devData = mygo.IsDev() && os.Getenv("CTRL_TOWER_DEV") == "1"
	d.app = a

	mygo.App.SetName("ctrl-tower")
	// The tray keeps the app running with every window closed.
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.WhenReady(d.ready)
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) {
		d.quitting = true
		_ = store.SaveNow(a.data)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func (d *desktop) ready() {
	a := d.app
	hour24 = mygo.App.Region().HourCycle == 24

	d.main = mygo.NewWindow(mygo.WindowOptions{
		Title:         "ctrl-tower",
		Width:         1120,
		Height:        740,
		MinWidth:      380,
		MinHeight:     520,
		TitleBarStyle: mygo.TitleBarHidden,
		StateKey:      "main",
		Hidden:        mygo.App.WasOpenedAtLogin(),
		Content:       ui.View(a.view),
	})
	_ = d.main.SetIcon(appIcon)
	// Closing the window hides it to the tray; Quit is in the tray menu.
	d.main.OnClose(func(e *mygo.CloseEvent) {
		if !d.quitting {
			e.PreventDefault()
			d.main.Hide()
		}
	})

	tray, err := mygo.NewTray(mygo.TrayOptions{Icon: trayIcon, ToolTip: "ctrl-tower"})
	if err != nil {
		log.Println("ctrl-tower: tray:", err)
	} else {
		d.tray = tray
		tray.OnClick(func() { d.toggleFlyout() })
		menu := mygo.NewMenu([]*mygo.MenuItem{
			{Label: "Open ctrl-tower", Click: func(*mygo.MenuItem, *mygo.Window) { d.showMain() }},
			{Label: "Start / pause timer", Click: func(*mygo.MenuItem, *mygo.Window) { a.startPause() }},
			mygo.Separator(),
			{Label: "Quit", Click: func(*mygo.MenuItem, *mygo.Window) { d.quitting = true; mygo.App.Quit() }},
		})
		tray.OnRightClick(func() {
			if d.flyout != nil {
				d.flyout.Close()
			}
			tray.SetMenu(menu)
			defer tray.SetMenu(nil)
			tray.PopUpMenu()
		})
		mygo.App.OnWillQuit(func(*mygo.QuitEvent) { tray.Destroy() })
	}

	// Another app may hold the shortcut: try a few others, and say so in
	// Settings.
	want := a.data.Settings.Hotkey
	if err := d.setHotkey(want); err != nil {
		a.hotkeyErr = want + " is taken by another app."
		for _, alt := range []string{"Ctrl+Alt+Space", "Ctrl+Shift+Space", "Ctrl+Alt+T", "Ctrl+Alt+K"} {
			if alt != want && d.setHotkey(alt) == nil {
				a.data.Settings.Hotkey, a.hotkeyDraft = alt, alt
				a.hotkeyErr += " Using " + alt + " instead."
				a.save()
				break
			}
		}
		log.Println("ctrl-tower:", a.hotkeyErr)
	}
	if a.data.Settings.OpenAtLogin && !mygo.App.OpenAtLogin() && !mygo.IsDev() {
		_ = mygo.App.SetOpenAtLogin(true)
	}
	mygo.App.OnSecondInstance(func([]string, string) { d.showMain() })

	a.tick()
	go func() {
		for range time.Tick(time.Second) {
			mygo.RunOnMain(a.tick)
		}
	}()
}

func (d *desktop) showMain() {
	if d.flyout != nil {
		d.flyout.Close()
	}
	d.main.Show()
	d.main.Restore()
	d.main.Focus()
}

func (d *desktop) toggleFlyout() {
	if d.flyout != nil {
		d.flyout.Close()
		return
	}
	if d.tray == nil {
		d.showMain()
		return
	}
	d.flyout = mygo.NewFlyout(mygo.FlyoutOptions{
		Tray:         d.tray,
		Placement:    mygo.PlacementTop,
		Gap:          8,
		Width:        360,
		Height:       480,
		Focusable:    true,
		Shadow:       true,
		CornerRadius: 12,
		Content:      ui.View(d.app.flyoutView),
	})
	d.flyout.OnClosed(func() {
		d.flyout = nil
		d.app.flyQuick, d.app.flyPreview = "", nil
	})
}

func (d *desktop) setHotkey(acc string) error {
	if d.hotkey != "" {
		mygo.GlobalShortcut.Unregister(d.hotkey)
	}
	if err := mygo.GlobalShortcut.Register(acc, func() { d.toggleFlyout() }); err != nil {
		if d.hotkey != "" && d.hotkey != acc {
			_ = mygo.GlobalShortcut.Register(d.hotkey, func() { d.toggleFlyout() })
		}
		return err
	}
	d.hotkey = acc
	return nil
}

// showOverlay brings up the meeting overlay at the top of the primary
// display, over every app, without taking the focus from what the user is
// typing into.
func (d *desktop) showOverlay() {
	if d.overlay != nil {
		d.overlay.Invalidate()
		return
	}
	const w, h = 440, 250
	x, y := 0, 64
	disp := mygo.Screen.PrimaryDisplay()
	x = disp.WorkArea.X + (disp.WorkArea.Width-w)/2
	y = disp.WorkArea.Y + 48
	d.overlay = mygo.NewWindow(mygo.WindowOptions{
		Title:        "Meeting starting",
		Width:        w,
		Height:       h,
		X:            x,
		Y:            y,
		Frameless:    true,
		AlwaysOnTop:  true,
		SkipTaskbar:  true,
		CornerRadius: 16,
		Hidden:       true,
		Content:      ui.View(d.app.overlayView),
	})
	d.overlay.ShowInactive()
	d.overlay.OnClosed(func() {
		d.overlay = nil
		d.app.ringing = nil
	})
}

func (d *desktop) closeOverlay() {
	if d.overlay != nil {
		d.overlay.Close()
	}
}

func (d *desktop) invalidate() {
	for _, w := range []*mygo.Window{d.main, d.flyout, d.overlay} {
		if w != nil {
			w.Invalidate()
		}
	}
}

func (d *desktop) notify(n note) {
	if runtime.GOOS == "darwin" && !mygo.NotificationsSupported() {
		return
	}
	nt := mygo.NewNotification(mygo.NotificationOptions{Title: n.title, Body: n.body, Silent: !d.app.data.Settings.Sound})
	nt.OnClick(func() {
		if n.link != "" {
			_ = mygo.Shell.OpenExternal(n.link)
		} else {
			d.showMain()
		}
		nt.Close()
	})
	go func() {
		if err := nt.Show(); err != nil {
			log.Println("ctrl-tower: notification:", err)
		}
	}()
}
