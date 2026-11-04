package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestSnapshots writes PNGs of every page with every data set to
// $CTRL_TOWER_SNAPSHOTS, for looking at the worst cases side by side.
func TestSnapshots(t *testing.T) {
	dir := os.Getenv("CTRL_TOWER_SNAPSHOTS")
	if dir == "" {
		t.Skip("CTRL_TOWER_SNAPSHOTS not set")
	}
	_ = os.MkdirAll(dir, 0o755)
	a, _ := testApp(t)
	a.devData = true
	save := func(tt *ui.Tester, name string) {
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_ = png.Encode(f, tt.Image())
	}
	for set, setName := range dataSets {
		a.setDataSet(set)
		for _, size := range [][2]int{{1120, 740}, {420, 720}} {
			tt := ui.NewTester(a.view, size[0], size[1])
			tt.SetDark(true)
			for _, n := range nav {
				a.page = n.id
				tt.Frame()
				tt.Advance(time.Second) // past the page's entrance
				save(tt, fmt.Sprintf("%d-%s-%s-%d", set, setName, n.id, size[0]))
			}
		}
		tt := ui.NewTester(a.flyoutView, 360, 480)
		tt.Frame()
		tt.Advance(time.Second)
		save(tt, fmt.Sprintf("%d-%s-flyout", set, setName))
	}
	// The overlay with three meetings ringing, and the editor.
	a.setDataSet(2)
	a.ringing = a.occurrences(startOfDay(testNow).Add(15*60*60e9), startOfDay(testNow).Add(15*60*60e9+1))[:3]
	tt := ui.NewTester(a.overlayView, 440, 250)
	tt.Advance(1e9)
	save(tt, "overlay-worst")
	a.page = pageToday
	tt = ui.NewTester(a.view, 1120, 740)
	tt.SetDark(false)
	occ := a.occurrences(startOfDay(testNow), startOfDay(testNow).AddDate(0, 0, 1))[0]
	a.editOccurrence(occ)
	tt.Frame()
	tt.Advance(1e9)
	save(tt, "editor-worst-light")
	a.editor.open = false
	a.page = pageToday
	tt.Frame()
	tt.Advance(time.Second)
	save(tt, "today-worst-light")
}
