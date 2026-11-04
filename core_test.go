package main

import (
	"testing"
	"time"
)

// A Wednesday afternoon, the clock of every test here.
var testNow = time.Date(2026, 11, 4, 14, 20, 0, 0, time.Local)

func day(y int, m time.Month, d, h, min int) time.Time {
	return time.Date(y, m, d, h, min, 0, 0, time.Local)
}

func TestRecurrence(t *testing.T) {
	standup := &Item{ID: "s", Kind: KindEvent, Title: "Standup", Start: day(2026, 11, 2, 9, 30), Minutes: 15, Repeat: Recurrence{Freq: FreqWeekdays}}
	rent := &Item{ID: "r", Kind: KindReminder, Title: "Rent", Start: day(2026, 1, 31, 9, 0), Repeat: Recurrence{Freq: FreqMonthly}}
	gym := &Item{ID: "g", Kind: KindEvent, Title: "Gym", Start: day(2026, 11, 2, 18, 0), Minutes: 60, Repeat: Recurrence{Freq: FreqWeekly, Days: []time.Weekday{time.Monday, time.Thursday}}}
	items := []*Item{standup, rent, gym}

	week := Occurrences(items, day(2026, 11, 2, 0, 0), day(2026, 11, 9, 0, 0))
	count := map[string]int{}
	for _, o := range week {
		count[o.Item.ID]++
	}
	if count["s"] != 5 || count["g"] != 2 || count["r"] != 0 {
		t.Errorf("week of Nov 2: %v", count)
	}

	// Monthly on the 31st falls on the last day of shorter months.
	feb := Occurrences(items, day(2026, 2, 1, 0, 0), day(2026, 3, 1, 0, 0))
	var rentDays []int
	for _, o := range feb {
		if o.Item.ID == "r" {
			rentDays = append(rentDays, o.Start.Day())
		}
	}
	if len(rentDays) != 1 || rentDays[0] != 28 {
		t.Errorf("rent in February on %v, want [28]", rentDays)
	}

	// One occurrence skipped, another moved to the afternoon.
	standup.Overrides = map[string]Override{
		"2026-11-03": {Skip: true},
		"2026-11-04": {Start: day(2026, 11, 4, 15, 0), Title: "Standup (late)"},
	}
	week = Occurrences([]*Item{standup}, day(2026, 11, 2, 0, 0), day(2026, 11, 9, 0, 0))
	if len(week) != 4 {
		t.Fatalf("%d standups after a skip, want 4", len(week))
	}
	if o := week[1]; o.Start.Hour() != 15 || o.Title != "Standup (late)" || o.Key != "2026-11-04" {
		t.Errorf("moved standup: %v %q %s", o.Start, o.Title, o.Key)
	}

	// Until ends the series on its day.
	gym.Repeat.Until = day(2026, 11, 5, 0, 0)
	if n := len(Occurrences([]*Item{gym}, day(2026, 11, 1, 0, 0), day(2026, 12, 1, 0, 0))); n != 2 {
		t.Errorf("%d gym sessions until Nov 5, want 2", n)
	}
}

func TestOccurrencesSpanningMidnight(t *testing.T) {
	late := &Item{ID: "l", Kind: KindEvent, Title: "Launch", Start: day(2026, 11, 4, 23, 30), Minutes: 90}
	if n := len(Occurrences([]*Item{late}, day(2026, 11, 5, 0, 0), day(2026, 11, 6, 0, 0))); n != 1 {
		t.Errorf("an event running past midnight shows on the next day %d times, want 1", n)
	}
}

func TestQuickAdd(t *testing.T) {
	cases := []struct {
		in       string
		kind     Kind
		title    string
		start    time.Time
		minutes  int
		url      string
		location string
		freq     Freq
	}{
		{"Design sync tomorrow 3pm 45m zoom.us/j/123", KindEvent, "Design sync", day(2026, 11, 5, 15, 0), 45, "https://zoom.us/j/123", "", ""},
		{"remind me to call the bank at 4", KindReminder, "Call the bank", day(2026, 11, 4, 16, 0), 0, "", "", ""},
		{"Standup every weekday 9:30 for 15 min", KindEvent, "Standup", day(2026, 11, 5, 9, 30), 15, "", "", FreqWeekdays},
		{"Dentist fri 10am @ 12 Allen Avenue, Ikeja", KindEvent, "Dentist", day(2026, 11, 6, 10, 0), 30, "", "12 Allen Avenue, Ikeja", ""},
		{"1:1 with Tolu next wednesday 11:00 1h https://meet.google.com/abc-defg-hij", KindEvent, "1:1 with Tolu", day(2026, 11, 11, 11, 0), 60, "https://meet.google.com/abc-defg-hij", "", ""},
		{"Read chapter 3 in 20m", KindEvent, "Read chapter 3", day(2026, 11, 4, 14, 40), 30, "", "", ""},
		{"Pay rent nov 30 noon", KindEvent, "Pay rent", day(2026, 11, 30, 12, 0), 30, "", "", ""},
		{"Gym every mon and thu 6pm 1h30", KindEvent, "Gym", day(2026, 11, 5, 18, 0), 90, "", "", FreqWeekly},
	}
	for _, c := range cases {
		p := ParseQuick(c.in, testNow, 10)
		it := p.Item
		if it.Kind != c.kind || it.Title != c.title || !it.Start.Equal(c.start) || it.Minutes != c.minutes ||
			it.URL != c.url || it.Location != c.location || it.Repeat.Freq != c.freq {
			t.Errorf("%q:\n got  %s %q %v %d %q %q %q\n want %s %q %v %d %q %q %q", c.in,
				it.Kind, it.Title, it.Start, it.Minutes, it.URL, it.Location, it.Repeat.Freq,
				c.kind, c.title, c.start, c.minutes, c.url, c.location, c.freq)
		}
	}
}

func TestTimer(t *testing.T) {
	p := Preset{Name: "Test", Work: 25, Break: 5, LongBreak: 15, LongEvery: 2}
	tm := NewTimer(p)
	now := testNow
	tm.Start(now)
	now = now.Add(10 * time.Minute)
	tm.Pause(now)
	now = now.Add(time.Hour) // paused time does not count
	if got := tm.Remaining(now); got != 15*time.Minute {
		t.Fatalf("remaining %v after 10 min of work, want 15m", got)
	}
	tm.Start(now)
	now = now.Add(15*time.Minute + 3*time.Second)
	tr, ok := tm.Tick(now, p, true)
	if !ok || tr.To != PhaseBreak || tr.Session == nil || int(tr.Session.Minutes) != 25 || !tr.Session.Full {
		t.Fatalf("after a full session: %+v %v", tr, ok)
	}
	// Auto-start chains the break from when the session ended.
	if !tm.Running || tm.Remaining(now) != 5*time.Minute-3*time.Second {
		t.Fatalf("break running %v with %v left", tm.Running, tm.Remaining(now))
	}
	now = now.Add(5 * time.Minute)
	tm.Tick(now, p, true)
	now = now.Add(25 * time.Minute)
	if tr, _ := tm.Tick(now, p, true); tr.To != PhaseLongBreak {
		t.Fatalf("after the second session: %v, want a long break", tr.To)
	}

	// A session skipped after 30 seconds records nothing.
	tm.Reset(p)
	tm.Start(now)
	if tr := tm.Skip(now.Add(30*time.Second), p, false); tr.Session != nil || tm.Running {
		t.Errorf("a 30-second session: %+v, running %v", tr.Session, tm.Running)
	}
}

func TestMeetingConflict(t *testing.T) {
	sync := &Item{ID: "m", Kind: KindEvent, Title: "Sync", Start: testNow.Add(40 * time.Minute), Minutes: 30}
	note := &Item{ID: "n", Kind: KindReminder, Title: "Water", Start: testNow.Add(10 * time.Minute)}
	occs := Occurrences([]*Item{sync, note}, testNow, testNow.Add(3*time.Hour))
	o, fit, ok := meetingConflict(testNow, testNow.Add(50*time.Minute), occs)
	if !ok || o.Title != "Sync" || fit != 38*time.Minute {
		t.Errorf("conflict %q %v %v, want Sync, 38m", o.Title, fit, ok)
	}
	if _, _, ok := meetingConflict(testNow, testNow.Add(25*time.Minute), occs); ok {
		t.Error("a session ending 15 minutes before the meeting conflicts")
	}
}

func TestDueAlerts(t *testing.T) {
	sync := &Item{ID: "m", Kind: KindEvent, Title: "Sync", Start: testNow.Add(10 * time.Minute), Minutes: 30, Lead: 10}
	fired := map[string]bool{}
	isFired := func(id string) bool { return fired[id] }
	got := dueAlerts([]*Item{sync}, testNow.Add(-time.Second), testNow, isFired)
	if len(got) != 1 || got[0].Kind != AlertLead {
		t.Fatalf("at the lead time: %+v", got)
	}
	fired[got[0].ID()] = true
	if got := dueAlerts([]*Item{sync}, testNow.Add(-time.Second), testNow, isFired); len(got) != 0 {
		t.Errorf("fired twice: %+v", got)
	}
	// Waking up 5 minutes after the start shows the start, not the lead.
	got = dueAlerts([]*Item{sync}, testNow.Add(-2*time.Hour), testNow.Add(15*time.Minute), func(string) bool { return false })
	if len(got) != 1 || got[0].Kind != AlertStart {
		t.Errorf("after sleeping through it: %+v", got)
	}
	// An hour late is too late.
	if got := dueAlerts([]*Item{sync}, testNow, testNow.Add(70*time.Minute), func(string) bool { return false }); len(got) != 0 {
		t.Errorf("an hour late: %+v", got)
	}
}

func TestFormat(t *testing.T) {
	if s := plural(1, "session", "sessions"); s != "1 session" {
		t.Error(s)
	}
	if s := plural(1284, "member", "members"); s != "1,284 members" {
		t.Error(s)
	}
	if s := countdown(25*time.Minute - 400*time.Millisecond); s != "25:00" {
		t.Error(s)
	}
	if s := until(testNow.Add(90*time.Second), testNow); s != "in 2 min" {
		t.Error(s)
	}
}
