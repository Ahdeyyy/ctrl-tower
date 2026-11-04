package main

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// editorState is the event editor's draft.
type editorState struct {
	open   bool
	id     string // the item edited, "" for a new one
	occKey string // the occurrence edited, of a repeating item
	occDay time.Time
	scope  int // 0: this occurrence, 1: the whole series

	kind     int // 0 event, 1 reminder
	title    string
	start    time.Time
	length   string
	where    int // 0 in person, 1 online
	location string
	link     string
	lead     string
	repeat   string
	days     [7]bool
	ends     bool
	until    time.Time
	notes    string
	err      string
}

var (
	lengthOptions = []string{"15 min", "30 min", "45 min", "1 h", "1 h 30 min", "2 h", "3 h", "4 h", "All day (8 h)"}
	lengthMinutes = []int{15, 30, 45, 60, 90, 120, 180, 240, 480}
	leadOptions   = []string{"Only when it starts", "5 min before", "10 min before", "15 min before", "30 min before", "1 h before", "1 day before"}
	leadMinutes   = []int{0, 5, 10, 15, 30, 60, 1440}
	repeatOptions = []string{"Doesn't repeat", "Every day", "Every weekday", "Weekly", "Monthly"}
	repeatFreqs   = []Freq{FreqNone, FreqDaily, FreqWeekdays, FreqWeekly, FreqMonthly}
)

func optionFor(options []string, values []int, v int) string {
	for i, x := range values {
		if x == v {
			return options[i]
		}
	}
	return ""
}

func valueFor(options []string, values []int, s string, def int) int {
	for i, o := range options {
		if o == s {
			return values[i]
		}
	}
	return def
}

// newItem opens the editor on a new item at start.
func (a *App) newItem(kind Kind, start time.Time) {
	e := editorState{open: true, start: start, length: "30 min", repeat: repeatOptions[0], until: start.AddDate(0, 1, 0)}
	e.lead = optionFor(leadOptions, leadMinutes, a.data.Settings.DefaultLead)
	if e.lead == "" {
		e.lead = leadOptions[2]
	}
	if kind == KindReminder {
		e.kind = 1
		e.lead = leadOptions[0]
	}
	a.editor = e
}

// newFromParsed opens the editor on what quick-add understood.
func (a *App) newFromParsed(p Parsed) {
	a.newItem(p.Item.Kind, p.Item.Start)
	a.fillEditor(&p.Item, p.Item.Start)
	a.editor.id = ""
}

// editOccurrence opens the editor on an occurrence; for a repeating item
// it edits that occurrence unless the user chooses the series.
func (a *App) editOccurrence(o Occurrence) {
	a.newItem(o.Item.Kind, o.Start)
	a.fillEditor(o.Item, o.Start)
	e := &a.editor
	e.id = o.Item.ID
	e.title, e.location, e.link, e.notes = o.Title, o.Location, o.URL, o.Notes
	if o.URL != "" {
		e.where = 1
	}
	if m := int(o.End.Sub(o.Start).Minutes()); m > 0 {
		e.length = lengthLabel(m)
	}
	if o.Item.Repeat.Freq != FreqNone {
		e.occKey = o.Key
		e.occDay, _ = time.ParseInLocation("2006-01-02", o.Key, o.Start.Location())
	}
}

func lengthLabel(m int) string {
	if s := optionFor(lengthOptions, lengthMinutes, m); s != "" {
		return s
	}
	return minutesLabel(m)
}

func (a *App) fillEditor(it *Item, start time.Time) {
	e := &a.editor
	e.kind = 0
	if it.Kind == KindReminder {
		e.kind = 1
	}
	e.title, e.start = it.Title, start
	e.location, e.link, e.notes = it.Location, it.URL, it.Notes
	if it.URL != "" {
		e.where = 1
	}
	if it.Minutes > 0 {
		e.length = lengthLabel(it.Minutes)
	}
	if s := optionFor(leadOptions, leadMinutes, it.Lead); s != "" {
		e.lead = s
	} else if it.Lead > 0 {
		e.lead = minutesLabel(it.Lead) + " before"
	}
	for i, f := range repeatFreqs {
		if f == it.Repeat.Freq {
			e.repeat = repeatOptions[i]
		}
	}
	e.days = [7]bool{}
	for _, d := range it.Repeat.Days {
		e.days[d] = true
	}
	if !it.Repeat.Until.IsZero() {
		e.ends, e.until = true, it.Repeat.Until
	}
}

// draftMinutes reads the length chosen, including one not in the list.
func draftMinutes(s string) int {
	if v := valueFor(lengthOptions, lengthMinutes, s, -1); v > 0 {
		return v
	}
	if d, n := parseDuration(compactDuration(s), 0); n > 0 {
		return int(d.Minutes())
	}
	return 30
}

func draftLead(s string) int {
	if v := valueFor(leadOptions, leadMinutes, s, -1); v >= 0 {
		return v
	}
	if d, n := parseDuration(compactDuration(strings.TrimSuffix(s, " before")), 0); n > 0 {
		return int(d.Minutes())
	}
	return 10
}

// compactDuration turns minutesLabel's "1 h 30 min" back into "1h30m".
func compactDuration(s string) []string {
	s = strings.ReplaceAll(s, " h", "h")
	s = strings.ReplaceAll(s, " min", "m")
	return []string{strings.ReplaceAll(s, " ", "")}
}

// saveEditor checks the draft and writes it, returning false with
// e.err set when it cannot.
func (a *App) saveEditor() bool {
	e := &a.editor
	e.err = ""
	title := strings.TrimSpace(e.title)
	if title == "" {
		e.err = "Give it a title."
		return false
	}
	link := ""
	if e.kind == 0 && e.where == 1 && strings.TrimSpace(e.link) != "" {
		link = normalizeURL(e.link)
		if link == "" {
			e.err = "That meeting link doesn't look like a web address."
			return false
		}
	}
	location := ""
	if e.kind == 0 && e.where == 0 {
		location = strings.TrimSpace(e.location)
	}
	minutes := 0
	if e.kind == 0 {
		minutes = draftMinutes(e.length)
	}
	start := e.start.Truncate(time.Minute)

	var rep Recurrence
	for i, o := range repeatOptions {
		if o == e.repeat {
			rep.Freq = repeatFreqs[i]
		}
	}
	if rep.Freq == FreqWeekly {
		for d := range 7 {
			if e.days[d] {
				rep.Days = append(rep.Days, time.Weekday(d))
			}
		}
	}
	if rep.Freq != FreqNone && e.ends {
		if startOfDay(e.until).Before(startOfDay(start)) {
			e.err = "The series ends before it starts."
			return false
		}
		rep.Until = startOfDay(e.until)
	}
	kind := KindEvent
	lead := draftLead(e.lead)
	if e.kind == 1 {
		kind = KindReminder
	}

	if e.id == "" {
		a.addItem(Item{Kind: kind, Title: title, Start: start, Minutes: minutes, Location: location, URL: link, Notes: strings.TrimSpace(e.notes), Lead: lead, Repeat: rep})
		e.open = false
		return true
	}
	it := a.data.item(e.id)
	if it == nil {
		e.open = false
		return true
	}
	if e.occKey != "" && e.scope == 0 {
		// This occurrence only: an override over the series.
		if it.Overrides == nil {
			it.Overrides = map[string]Override{}
		}
		ov := Override{Title: title, Location: location, URL: link, Notes: strings.TrimSpace(e.notes), Minutes: minutes}
		if !start.Equal(at(e.occDay, it.Start)) {
			ov.Start = start
		}
		it.Overrides[e.occKey] = ov
	} else {
		if e.occKey != "" {
			// The series moves by as many days as this occurrence did, to
			// the new time of day.
			shift := int(startOfDay(start).Sub(e.occDay).Hours()/24 + 0.5)
			start = at(startOfDay(it.Start).AddDate(0, 0, shift), start)
		}
		it.Kind, it.Title, it.Start, it.Minutes = kind, title, start, minutes
		it.Location, it.URL, it.Notes, it.Lead, it.Repeat = location, link, strings.TrimSpace(e.notes), lead, rep
		it.clamp()
	}
	a.save()
	e.open = false
	return true
}

// editorView builds the editor dialog while it is open.
func (a *App) editorView(c *ui.Context) {
	if !a.editor.open {
		return
	}
	p := pal(c)
	w, h := c.Size()
	e := &a.editor
	ui.DialogBase(c, &e.open, func(backdrop, panel ui.Element) {
		// The backdrop dims in as the panel pops up.
		dim := float32(0.32)
		if p.dark {
			dim = 0.55
		}
		fade := backdrop.Keyframes("dim", []ui.Keyframe{{Value: 0}, {At: 180 * time.Millisecond, Value: 1, Ease: easeOut}})
		backdrop.Background(ui.Hex("#05060a").Alpha(dim * fade))
		pane(c, panel).Background(p.raised).Width(min(520, w-24)).MaxHeight(h-32).Radius(18).
			Shadow(0, 24, 64, -12, p.lift)
		popIn(panel)
		panel.Children(func() {
			title := "New event"
			switch {
			case e.id != "":
				title = "Edit"
			case e.kind == 1:
				title = "New reminder"
			}
			ic, tint := icToday, p.accent
			if e.kind == 1 {
				ic, tint = icBell, p.warn
			}
			ui.Row(c).Padding(16, 16, 8, 20).Gap(10).AlignItems(ui.Center).Children(func() {
				tile := ui.Box(c).Size(30, 30).Radius(9).Center().Shrink(0)
				tint = tile.AnimateColor("tint", tint, colorAnim)
				tile.Background(tint.Alpha(0.15)).Children(func() {
					ui.Icon(c, ic).Size(16, 16).TextColor(tint).ContentTransition(ui.ContentTransition{})
				})
				ui.Text(c, title).FontSize(17).Bold().Grow(1).ContentTransition(ui.ContentTransition{})
				iconButton(c.Key("close"), icClose, "Close").OnClick(func() { e.open = false })
			})
			ui.Scroll(c).Grow(1).Shrink(1).Padding(4, 20, 8, 20).Gap(14).Children(func() { a.editorFields(c) })
			a.editorFooter(c)
		})
	})
}

func (a *App) editorFields(c *ui.Context) {
	p := pal(c)
	e := &a.editor
	if e.id == "" {
		segmented(c.Key("kind"), &e.kind, "Event", "Reminder").Label("Kind")
	}
	if e.occKey != "" {
		ui.Column(c).Gap(6).Children(func() {
			heading(c, "CHANGE")
			segmented(c.Key("scope"), &e.scope, "This one", "All in the series").Label("Change")
		})
	}
	ui.Field(c, "Title", func() {
		ui.TextInput(c.Key("title"), &e.title).Placeholder("What is it?").AutoFocus().OnSubmit(func() {
			if a.saveEditor() {
				a.h.invalidate()
			}
		})
	})
	ui.Row(c).Gap(10).Wrap().AlignItems(ui.End).Children(func() {
		ui.Field(c, "Date", func() { ui.DateInput(c.Key("date"), &e.start) }).Grow(1).MinWidth(150)
		ui.Field(c, "Time", func() { ui.TimeInput(c.Key("time"), &e.start) }).MinWidth(120)
		if e.kind == 0 {
			opts := lengthOptions
			if valueFor(lengthOptions, lengthMinutes, e.length, -1) < 0 {
				opts = append([]string{e.length}, lengthOptions...)
			}
			ui.Field(c, "Length", func() { ui.Select(c.Key("length"), &e.length, opts) }).MinWidth(130)
		}
	})
	if e.kind == 0 {
		ui.Column(c).Gap(8).Children(func() {
			segmented(c.Key("where"), &e.where, "In person", "Online").Label("Where")
			if e.where == 0 {
				ui.TextInput(c.Key("loc"), &e.location).Placeholder("Address or room").Label("Location")
			} else {
				in := ui.TextInput(c.Key("link"), &e.link).Placeholder("Paste the meeting link (Zoom, Meet, Teams…)").Label("Meeting link")
				if e.link != "" && normalizeURL(e.link) == "" && !in.Focused() {
					ui.Text(c, "Not a web address yet.").FontSize(12).TextColor(p.warn)
				}
			}
		})
	}
	ui.Row(c).Gap(10).Wrap().Children(func() {
		opts := leadOptions
		if valueFor(leadOptions, leadMinutes, e.lead, -1) < 0 {
			opts = append([]string{e.lead}, leadOptions...)
		}
		ui.Field(c, "Alert", func() { ui.Select(c.Key("lead"), &e.lead, opts) }).Grow(1).MinWidth(170)
		ui.Field(c, "Repeat", func() { ui.Select(c.Key("repeat"), &e.repeat, repeatOptions) }).Grow(1).MinWidth(170)
	})
	if e.repeat == "Weekly" {
		ui.Row(c).Gap(6).Wrap().Children(func() {
			for _, d := range sortedDays([]time.Weekday{0, 1, 2, 3, 4, 5, 6}) {
				on := e.days[d]
				if !anyDay(e.days) && d == e.start.Weekday() {
					on = true
				}
				t := ui.Toggle(c.Key(d), &on, d.String()[:3])
				if t.Changed() {
					if !anyDay(e.days) {
						e.days[e.start.Weekday()] = true
					}
					e.days[d] = on
				}
			}
		})
	}
	if e.repeat != repeatOptions[0] {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Checkbox(c.Key("ends"), &e.ends, "Ends on")
			if e.ends {
				ui.DateInput(c.Key("until"), &e.until).Label("Ends on")
			}
		})
	}
	ui.Field(c, "Notes", func() {
		ui.TextArea(c.Key("notes"), &e.notes).Lines(2, 6).Placeholder("Agenda, dial-in, who's coming…")
	})
	if e.err != "" {
		ui.Text(c, e.err).TextColor(p.danger).FontSize(13)
	}
}

func anyDay(d [7]bool) bool {
	for _, v := range d {
		if v {
			return true
		}
	}
	return false
}

func (a *App) editorFooter(c *ui.Context) {
	p := pal(c)
	e := &a.editor
	ui.Row(c).Padding(12, 20, 16, 20).Gap(8).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(p.divider).Children(func() {
		if e.id != "" {
			label := "Delete"
			if e.occKey != "" && e.scope == 0 {
				label = "Delete this one"
			}
			del := pillButton(c.Key("delete"), icTrash, label, false).TextColor(p.danger)
			del.OnClick(func() { a.deleteFromEditor(c) })
		}
		ui.Spacer(c)
		pillButton(c.Key("cancel"), nil, "Cancel", false).OnClick(func() { e.open = false })
		pillButton(c.Key("save"), icCheck, "Save", true).OnClick(func() { a.saveEditor() })
	})
}

func (a *App) deleteFromEditor(c *ui.Context) {
	e := &a.editor
	it := a.data.item(e.id)
	e.open = false
	if it == nil {
		return
	}
	if e.occKey != "" && e.scope == 0 {
		prev, had := it.Overrides[e.occKey]
		key := e.occKey
		a.skipOccurrence(Occurrence{Item: it, Key: key})
		c.ToastAction("Removed from the series", "Undo", func() {
			if had {
				it.Overrides[key] = prev
			} else {
				delete(it.Overrides, key)
			}
			a.save()
		})
		return
	}
	removed := a.deleteItem(it.ID)
	c.ToastAction("Deleted “"+clampText(removed.Title, 40)+"”", "Undo", func() { a.restoreItem(removed) })
}
