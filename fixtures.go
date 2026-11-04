package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

// The dev data switch shows the views with fixtures in place of the
// user's data, to see what realistic worst cases do to them. It only shows
// in development builds with CTRL_TOWER_DEV=1, and fixtures are never
// saved.

var dataSets = []string{"Real", "Demo", "Worst case", "Empty", "One", "1,000 rows"}

func (a *App) setDataSet(i int) {
	if a.realData == nil {
		a.realData, a.realTimer = a.data, a.timer
	}
	a.dataSet = i
	now := a.now()
	switch i {
	case 0:
		a.data, a.timer = a.realData, a.realTimer
		a.store.Disabled = false
		a.ringing = nil
		return
	case 1:
		a.data = demoData(now)
	case 2:
		a.data = worstData(now)
	case 3:
		a.data = newData()
	case 4:
		a.data = oneData(now)
	case 5:
		a.data = hugeData(now)
	}
	a.store.Disabled = true
	a.timer = NewTimer(a.preset())
	if i == 2 {
		a.timer.Tag = a.data.Tags[0]
		a.timer.Start(now.Add(-17 * time.Minute))
		a.held = make([]Occurrence, 3)
	}
}

func (a *App) devSwitch(c *ui.Context) {
	if !a.devData {
		return
	}
	ui.Overlay(c, func() {
		chosen := a.dataSet
		seg := ui.SegmentedBase(c.Key("dev-data"), &chosen, len(dataSets))
		seg.Track.Attach(ui.AnchorBottom, ui.AnchorBottom).Bottom(12).Row().Padding(3).Gap(2).Radius(8).
			Background(ui.RGB(229, 231, 235)).Border(1, ui.RGB(209, 213, 219)).Label("Data").Children(func() {
			for i, name := range dataSets {
				s := seg.Segment(i).Padding(4, 10).Radius(6).Label(name)
				if i == a.dataSet {
					s.Background(ui.RGB(255, 255, 255))
				}
				s.Children(func() { ui.Text(c, name).FontSize(12).TextColor(ui.RGB(31, 41, 55)).SingleLine() })
			}
		})
		if chosen != a.dataSet {
			a.setDataSet(chosen)
		}
	})
}

func fixtureItem(kind Kind, title string, start time.Time, minutes int) *Item {
	return &Item{ID: newID(), Kind: kind, Title: title, Start: start, Minutes: minutes, Lead: 10, Created: start}
}

func demoData(now time.Time) *Data {
	d := newData()
	day := startOfDay(now)
	at := func(dd, h, m int) time.Time { return day.AddDate(0, 0, dd).Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	standup := fixtureItem(KindEvent, "Standup", at(-7, 9, 30), 15)
	standup.Repeat.Freq, standup.URL = FreqWeekdays, "https://meet.google.com/abc-defg-hij"
	sync := fixtureItem(KindEvent, "Design sync", at(0, 15, 0), 45)
	sync.URL = "https://zoom.us/j/1234567890"
	dentist := fixtureItem(KindEvent, "Dentist", at(1, 10, 0), 30)
	dentist.Location = "12 Allen Avenue, Ikeja"
	bank := fixtureItem(KindReminder, "Call the bank", at(0, 16, 0), 0)
	bank.Lead = 0
	one := fixtureItem(KindEvent, "1:1 with Tolu", at(2, 11, 0), 30)
	one.URL = "https://teams.microsoft.com/l/meetup-join/abc"
	d.Items = []*Item{standup, sync, dentist, bank, one}
	tags := []string{"Thesis", "Atomic Habits", "Client work"}
	for i := 0; i < 14; i++ {
		for k := 0; k < i%4+1; k++ {
			d.Sessions = append(d.Sessions, Session{Start: at(-13+i, 9+k, 0), Minutes: 50, Tag: tags[(i+k)%3], Preset: "Deep work", Full: true})
		}
	}
	d.Tags = tags
	return d
}

func worstData(now time.Time) *Data {
	d := newData()
	day := startOfDay(now)
	at := func(dd, h, m int) time.Time { return day.AddDate(0, 0, dd).Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	ev := func(title string, start time.Time, minutes int) *Item {
		it := fixtureItem(KindEvent, title, start, minutes)
		d.Items = append(d.Items, it)
		return it
	}
	rem := func(title string, start time.Time) *Item {
		it := fixtureItem(KindReminder, title, start, 0)
		d.Items = append(d.Items, it)
		return it
	}

	// The longest title the editor takes, written as people write them.
	long := ev("Q4 Cross-Functional Roadmap Alignment & Budget Reconciliation Review — Engineering, Product, Finance, Legal and Customer Success (mandatory, cameras on, please read the pre-read deck beforehand)", at(0, 8, 0), 90)
	long.URL = "https://teams.microsoft.com/l/meetup-join/19%3ameeting_NjA2ZTk3YjQtZDM1Ny00YjU4LWE4ZjItYzQ2ODE0NzQ5ZTIx%40thread.v2/0?context=%7b%22Tid%22%3a%2272f988bf-86f1-41af-91ab-2d7cd011db47%22%2c%22Oid%22%3a%22c1d4e8a0-3c9e-4d1b-9a5e-2f6b7c8d9e0f%22%7d"
	long.Location = ""
	ev("Jo", at(0, 9, 45), 15).URL = "https://whereby.com/jo"
	off := ev("Offsite planning", at(0, 10, 0), 60)
	off.Location = "Northwind Industries Holdings HQ, Conference Room 14B (Wiśniewska-Kowalczyk Suite), 1600 Amphitheatre Parkway Extension, Building C, 4th Floor, Ikeja GRA, Lagos"
	ev("Re:Re:Fwd:Fwd:CustomerEscalation#48213-Northwind-Industries-Holdings-Priority-P0", at(0, 11, 0), 30).URL = "https://bartholomew.fitzgerald.northwind-industries-holdings.example.com/meetings/escalation-48213"
	ev("🎂 Jo's birthday lunch", at(0, 12, 30), 60).Location = "Café"
	ev("Đánh giá thiết kế với nhóm Việt Nam", at(0, 13, 0), 30)
	ev("会議：四半期レビュー（全社）", at(0, 13, 0), 30)
	ev("اجتماع الفريق الأسبوعي", at(0, 13, 0), 30)
	// A double-booked afternoon.
	for i := range 6 {
		ev(fmt.Sprintf("Interview loop — candidate %d", i+1), at(0, 15, 0), 45).URL = "https://zoom.us/j/98765432" + fmt.Sprint(i)
	}
	ev("Zero-length marker", at(0, 17, 0), 0)
	ev("Launch night", at(0, 23, 30), 120).URL = "https://meet.google.com/xyz-abcd-efg"
	rem("Take the bins out (it's Wednesday, they come early)", at(0, 23, 59))
	rem("x", at(0, 18, 0))
	daily := ev("Standup", at(-30, 9, 30), 15)
	daily.Repeat = Recurrence{Freq: FreqWeekly, Days: []time.Weekday{1, 2, 3, 4, 5, 6, 0}}
	ev("All-day workshop", at(1, 9, 0), 480).Location = "Lagos Business School"
	month := rem("Rent", time.Date(now.Year(), 1, 31, 9, 0, 0, 0, now.Location()))
	month.Repeat.Freq = FreqMonthly

	d.Settings.Presets[0].Name = "Deep work — thesis chapter 3 drafts"
	d.Settings.Presets = append(d.Settings.Presets, Preset{Name: "Ultra", Work: 180, Break: 60, LongBreak: 90, LongEvery: 12})
	d.Tags = []string{"Atomic Habits: An Easy & Proven Way to Build Good Habits & Br", "Thesis", "x"}
	for i := range 40 {
		d.Tags = append(d.Tags, fmt.Sprintf("Client %d — Northwind", i))
	}
	for i := 0; i < 1284; i++ {
		d.Sessions = append(d.Sessions, Session{
			Start: day.AddDate(0, 0, -i%365).Add(time.Duration(6+i%14) * time.Hour), Minutes: 25 + float64(i%160),
			Tag: d.Tags[i%len(d.Tags)], Preset: "Deep work", Full: i%3 != 0,
		})
	}
	return d
}

func oneData(now time.Time) *Data {
	d := newData()
	it := fixtureItem(KindEvent, "Design sync", startOfDay(now).Add(15*time.Hour), 1)
	d.Items = []*Item{it}
	d.Sessions = []Session{{Start: now.Add(-2 * time.Hour), Minutes: 1, Tag: "Thesis", Preset: "Deep work", Full: true}}
	d.Tags = []string{"Thesis"}
	d.Settings.Presets[0] = Preset{Name: "Deep work", Work: 1, Break: 1, LongBreak: 1, LongEvery: 1}
	return d
}

func hugeData(now time.Time) *Data {
	d := newData()
	day := startOfDay(now)
	for i := range 1000 {
		it := fixtureItem(KindEvent, fmt.Sprintf("Imported event %s", thousands(i+1)), day.Add(time.Duration(i)*86*time.Second), 15)
		d.Items = append(d.Items, it)
	}
	return d
}
