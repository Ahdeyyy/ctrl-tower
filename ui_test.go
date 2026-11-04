package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func testApp(t *testing.T) (*App, *[]note) {
	t.Helper()
	var notes []note
	store := NewStore(t.TempDir())
	a := NewApp(store, newData(), hooks{notify: func(n note) { notes = append(notes, n) }})
	now := testNow
	a.clock = func() time.Time { return now }
	return a, &notes
}

// Every page and data set builds at the window's sizes, light and dark.
func TestPagesBuild(t *testing.T) {
	a, _ := testApp(t)
	for set := range dataSets {
		a.setDataSet(set)
		for _, size := range [][2]int{{1120, 740}, {380, 600}} {
			for _, dark := range []bool{false, true} {
				tt := ui.NewTester(a.view, size[0], size[1])
				tt.SetDark(dark)
				for _, n := range nav {
					a.page = n.id
					tt.Frame()
				}
			}
		}
		ui.NewTester(a.flyoutView, 360, 480).Frame()
	}
	a.setDataSet(0)
}

func TestQuickAddFlow(t *testing.T) {
	a, _ := testApp(t)
	tt := ui.NewTester(a.view, 1120, 740)
	if err := tt.Click("Quick add"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Design sync tomorrow 3pm 45m zoom.us/j/123")
	if !tt.HasText("Design sync") || !tt.HasText("Zoom") {
		t.Fatalf("preview missing: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	if len(a.data.Items) != 1 {
		t.Fatalf("%d items after Enter", len(a.data.Items))
	}
	it := a.data.Items[0]
	if it.Title != "Design sync" || it.URL != "https://zoom.us/j/123" || it.Minutes != 45 {
		t.Errorf("added %+v", it)
	}
	if a.quick != "" {
		t.Errorf("quick-add kept %q", a.quick)
	}
}

func TestEditorFlow(t *testing.T) {
	a, _ := testApp(t)
	tt := ui.NewTester(a.view, 1120, 740)
	tt.Key(ui.Ctrl, ui.KeyN)
	if !a.editor.open {
		t.Fatal("Ctrl+N did not open the editor")
	}
	tt.Click("Save")
	if a.editor.err == "" || len(a.data.Items) != 0 {
		t.Fatalf("saved without a title: %q", a.editor.err)
	}
	tt.Click("Title")
	tt.Type("Board review")
	tt.Click("Online")
	tt.Click("Meeting link")
	tt.Type("meet.google.com/abc-defg-hij")
	tt.Click("Save")
	if a.editor.open || len(a.data.Items) != 1 {
		t.Fatalf("editor open %v, %d items, err %q", a.editor.open, len(a.data.Items), a.editor.err)
	}
	if it := a.data.Items[0]; it.URL != "https://meet.google.com/abc-defg-hij" || it.Lead != 10 {
		t.Errorf("saved %+v", it)
	}
}

func TestMeetingAlertsAndOverlay(t *testing.T) {
	a, notes := testApp(t)
	now := testNow
	a.clock = func() time.Time { return now }
	sync := &Item{ID: "m", Kind: KindEvent, Title: "Sync", Start: now.Add(10 * time.Minute), Minutes: 30, Lead: 10, URL: "https://zoom.us/j/1"}
	a.data.Items = []*Item{sync}
	a.tick()
	if len(*notes) != 1 {
		t.Fatalf("%d notifications at the lead time", len(*notes))
	}
	now = now.Add(10 * time.Minute)
	a.tick()
	if len(a.ringing) != 1 {
		t.Fatalf("%d ringing at the start", len(a.ringing))
	}
	var opened []string
	a.h.openURL = func(u string) { opened = append(opened, u) }
	tt := ui.NewTester(a.overlayView, 440, 250)
	tt.Advance(time.Second)
	if err := tt.Click("Join Zoom"); err != nil {
		t.Fatal(err)
	}
	if len(opened) != 1 || len(a.ringing) != 0 {
		t.Errorf("opened %v, ringing %d", opened, len(a.ringing))
	}
}

func TestFocusDNDHoldsReminders(t *testing.T) {
	a, notes := testApp(t)
	now := testNow
	a.clock = func() time.Time { return now }
	a.data.Items = []*Item{{ID: "w", Kind: KindReminder, Title: "Water", Start: now.Add(5 * time.Minute)}}
	a.startPause()
	a.tick()
	now = now.Add(5 * time.Minute)
	a.tick()
	if len(*notes) != 0 || len(a.held) != 1 {
		t.Fatalf("during focus: %d notes, %d held", len(*notes), len(a.held))
	}
	a.skip()
	if len(a.held) != 0 || len(*notes) < 1 {
		t.Errorf("after the session: %d held, notes %+v", len(a.held), *notes)
	}
}
