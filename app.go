package main

import (
	"fmt"
	"net/url"
	"time"
)

// Pages of the main window.
const (
	pageToday    = "today"
	pageUpcoming = "upcoming"
	pageFocus    = "focus"
	pageStats    = "stats"
	pageSettings = "settings"
)

// note is a desktop notification; a click opens link when it has one, or
// the main window.
type note struct {
	title, body, link string
}

// hooks are what the app asks of the desktop. main wires them to MyGo;
// tests leave them recording or empty.
type hooks struct {
	notify       func(note)
	openURL      func(string)
	invalidate   func() // redraw every window
	showOverlay  func()
	closeOverlay func()
	showMain     func()
	setLogin     func(bool)
	setHotkey    func(string) error
}

// App is the state of ctrl-tower. It is read and changed on the main
// thread only: by the views, the tray and the tick.
type App struct {
	store *Store
	data  *Data
	h     hooks
	clock func() time.Time

	timer     Timer
	lastCheck time.Time

	// Main window.
	page     string
	quick    string  // the quick-add line
	preview  *Parsed // what it parses to, shown under it
	editor   editorState
	upcoming int // days the agenda shows

	// Overlay: occurrences ringing now, and those snoozed until later.
	ringing []Occurrence
	snoozed []snooze
	// Reminders held while a work session runs (Do Not Disturb).
	held []Occurrence

	// conflictSeen is the meeting whose conflict the user answered, so
	// the timer does not ask again about it.
	conflictSeen string

	// Settings page drafts.
	hotkeyDraft string
	hotkeyErr   string

	// Flyout.
	flyQuick   string
	flyPreview *Parsed

	// Dev data switch (dev builds only).
	devData   bool
	dataSet   int
	realData  *Data
	realTimer Timer
}

type snooze struct {
	occ   Occurrence
	until time.Time
}

func NewApp(store *Store, data *Data, h hooks) *App {
	a := &App{store: store, data: data, h: h, clock: time.Now, page: pageToday, upcoming: 14}
	a.timer = NewTimer(a.preset())
	a.hotkeyDraft = data.Settings.Hotkey
	a.fillHooks()
	return a
}

func (a *App) fillHooks() {
	h := &a.h
	if h.notify == nil {
		h.notify = func(note) {}
	}
	if h.openURL == nil {
		h.openURL = func(string) {}
	}
	if h.invalidate == nil {
		h.invalidate = func() {}
	}
	if h.showOverlay == nil {
		h.showOverlay = func() {}
	}
	if h.closeOverlay == nil {
		h.closeOverlay = func() {}
	}
	if h.showMain == nil {
		h.showMain = func() {}
	}
	if h.setLogin == nil {
		h.setLogin = func(bool) {}
	}
	if h.setHotkey == nil {
		h.setHotkey = func(string) error { return nil }
	}
}

func (a *App) now() time.Time { return a.clock() }

func (a *App) save() { a.store.Save(a.data) }

func (a *App) preset() Preset {
	s := a.data.Settings
	return s.Presets[clampInt(s.Preset, 0, len(s.Presets)-1)]
}

// occurrences between from and to, of the data shown.
func (a *App) occurrences(from, to time.Time) []Occurrence {
	return Occurrences(a.data.Items, from, to)
}

// nextEvent is the event under way or the next one today and tomorrow,
// with the reminders left out.
func (a *App) nextEvent(now time.Time) (Occurrence, bool) {
	for _, o := range a.occurrences(now.Add(-6*time.Hour), now.Add(48*time.Hour)) {
		if o.IsEvent() && o.End.After(now) {
			return o, true
		}
	}
	return Occurrence{}, false
}

// tick runs every second on the main thread: it ends timer phases, shows
// due alerts and wakes snoozed ones.
func (a *App) tick() {
	now := a.now()
	changed := false

	if tr, ok := a.timer.Tick(now, a.preset(), a.data.Settings.AutoStart); ok {
		a.phaseEnded(tr, now)
		changed = true
	}

	since := a.lastCheck
	if since.IsZero() {
		since = now.Add(-missedWindow)
	}
	a.lastCheck = now
	for _, al := range dueAlerts(a.data.Items, since, now, a.data.fired) {
		a.data.markFired(al.ID())
		a.alert(al)
		changed = true
	}

	kept := a.snoozed[:0]
	for _, s := range a.snoozed {
		if now.Before(s.until) {
			kept = append(kept, s)
			continue
		}
		a.ring(s.occ)
		changed = true
	}
	a.snoozed = kept

	if changed {
		a.save()
		a.h.invalidate()
	}
}

func (a *App) focusing() bool {
	return a.timer.Phase == PhaseWork && a.timer.Running
}

func (a *App) alert(al Alert) {
	o := al.Occ
	switch {
	case al.Kind == AlertLead:
		body := span(o)
		if o.Location != "" {
			body += " · " + o.Location
		}
		title := fmt.Sprintf("%s %s", o.Title, until(o.Start, a.now()))
		a.h.notify(note{title: title, body: body, link: o.URL})
	case o.IsEvent():
		a.ring(o)
	case a.data.Settings.FocusDND && a.focusing():
		a.held = append(a.held, o)
	default:
		a.remind(o)
	}
}

func (a *App) remind(o Occurrence) {
	a.h.notify(note{title: o.Title, body: "Reminder · " + clock(o.Start), link: o.URL})
}

// ring shows an event in the overlay.
func (a *App) ring(o Occurrence) {
	for _, r := range a.ringing {
		if r.ID() == o.ID() {
			return
		}
	}
	a.ringing = append(a.ringing, o)
	a.h.showOverlay()
	// The notification too, so it is in the Action Center afterwards.
	body := "Starting now"
	if o.Location != "" {
		body += " · " + o.Location
	}
	a.h.notify(note{title: o.Title, body: body, link: o.URL})
}

func (a *App) dismiss(o Occurrence) {
	out := a.ringing[:0]
	for _, r := range a.ringing {
		if r.ID() != o.ID() {
			out = append(out, r)
		}
	}
	a.ringing = out
	if len(a.ringing) == 0 {
		a.h.closeOverlay()
	}
}

func (a *App) snoozeFor(o Occurrence, d time.Duration) {
	a.dismiss(o)
	a.snoozed = append(a.snoozed, snooze{occ: o, until: a.now().Add(d)})
}

// join opens the meeting's link, or the place in Maps.
func (a *App) join(o Occurrence) {
	if o.URL != "" {
		a.h.openURL(o.URL)
	} else if o.Location != "" {
		a.h.openURL(mapsURL(o.Location))
	}
}

func (a *App) phaseEnded(tr Transition, now time.Time) {
	if tr.Session != nil {
		a.data.Sessions = append(a.data.Sessions, *tr.Session)
		a.data.useTag(tr.Session.Tag)
	}
	p := a.preset()
	switch tr.To {
	case PhaseBreak:
		a.h.notify(note{title: "Break time", body: fmt.Sprintf("Session done. %s break.", minutesLabel(p.Break))})
	case PhaseLongBreak:
		a.h.notify(note{title: "Long break", body: fmt.Sprintf("%s done. Take %s.", plural(p.LongEvery, "session", "sessions"), minutesLabel(p.LongBreak))})
	case PhaseWork:
		body := "Ready for the next session."
		if a.timer.Running {
			body = fmt.Sprintf("Next session started: %s.", minutesLabel(p.Work))
		}
		a.h.notify(note{title: "Back to focus", body: body})
	}
	if tr.From == PhaseWork {
		a.releaseHeld()
	}
	a.conflictSeen = ""
}

// releaseHeld shows the reminders held during a session, together.
func (a *App) releaseHeld() {
	switch len(a.held) {
	case 0:
		return
	case 1:
		a.remind(a.held[0])
	default:
		body := ""
		for i, o := range a.held {
			if i > 0 {
				body += " · "
			}
			body += o.Title
		}
		a.h.notify(note{title: plural(len(a.held), "reminder", "reminders") + " while you focused", body: body})
	}
	a.held = nil
}

// Timer actions.

func (a *App) startPause() {
	now := a.now()
	if a.timer.Running {
		a.timer.Pause(now)
	} else {
		a.timer.Start(now)
	}
	a.h.invalidate()
}

func (a *App) skip() {
	tr := a.timer.Skip(a.now(), a.preset(), false)
	a.phaseEnded(tr, a.now())
	a.save()
	a.h.invalidate()
}

func (a *App) resetTimer() {
	a.timer.Reset(a.preset())
	a.held = nil
	a.conflictSeen = ""
	a.h.invalidate()
}

func (a *App) choosePreset(i int) {
	if i == a.data.Settings.Preset {
		return
	}
	a.data.Settings.Preset = i
	if a.timer.Fresh() && a.timer.Phase == PhaseWork {
		a.timer.Reset(a.preset())
	}
	a.save()
}

// conflict is the meeting the current or next work session runs into,
// when the user has not answered for it yet.
func (a *App) conflict(now time.Time) (Occurrence, time.Duration, bool) {
	if !a.data.Settings.MeetingAware || a.timer.Phase != PhaseWork {
		return Occurrence{}, 0, false
	}
	ends := now.Add(a.timer.Remaining(now))
	occs := a.occurrences(now, ends.Add(time.Hour))
	o, fit, ok := meetingConflict(now, ends, occs)
	if !ok || o.ID() == a.conflictSeen {
		return Occurrence{}, 0, false
	}
	return o, fit, true
}

func (a *App) shortenFor(o Occurrence, fit time.Duration) {
	a.timer.SetLength(a.now(), a.timer.Length-a.timer.Remaining(a.now())+fit)
	a.conflictSeen = o.ID()
}

// Items.

func (a *App) addItem(it Item) *Item {
	n := it
	n.ID = newID()
	n.Created = a.now()
	n.clamp()
	a.data.Items = append(a.data.Items, &n)
	a.save()
	return &n
}

func (a *App) deleteItem(id string) *Item {
	it := a.data.removeItem(id)
	a.save()
	return it
}

func (a *App) restoreItem(it *Item) {
	if it != nil && a.data.item(it.ID) == nil {
		a.data.Items = append(a.data.Items, it)
		a.save()
	}
}

// skipOccurrence removes one occurrence of a series.
func (a *App) skipOccurrence(o Occurrence) {
	it := o.Item
	if it.Overrides == nil {
		it.Overrides = map[string]Override{}
	}
	it.Overrides[o.Key] = Override{Skip: true}
	a.save()
}

func mapsURL(place string) string {
	return "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(place)
}
