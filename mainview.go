package main

import (
	"time"

	"github.com/egoist/mygo/ui"
)

type navItem struct {
	id    string
	icon  *ui.SVG
	label string
}

var nav = []navItem{
	{pageToday, icToday, "Today"},
	{pageUpcoming, icUpcoming, "Upcoming"},
	{pageFocus, icFocus, "Focus"},
	{pageStats, icStats, "Stats"},
	{pageSettings, icSettings, "Settings"},
}

// view is the main window.
func (a *App) view(c *ui.Context) {
	appTheme(c)
	backdrop(c)
	a.redrawSoon(c)
	w, _ := c.Size()
	compact := w < 860

	a.shortcuts(c)
	ui.Column(c).Fill().Children(func() {
		a.titleBar(c, compact)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Padding(0, 12, 12, 12).Gap(12).Children(func() {
			a.navRail(c, compact)
			page := ui.Column(c.Key(a.page)).Grow(1).MinWidth(0)
			riseIn(page, 10, 260*time.Millisecond, 0)
			page.Children(func() {
				switch a.page {
				case pageUpcoming:
					a.upcomingPage(c)
				case pageFocus:
					a.focusPage(c)
				case pageStats:
					a.statsPage(c)
				case pageSettings:
					a.settingsPage(c)
				default:
					a.todayPage(c, compact)
				}
			})
		})
	})
	a.editorView(c)
	a.devSwitch(c)
}

// redrawSoon asks for the frame that next changes what shows: the next
// second while the timer runs, otherwise the next minute, for the
// countdowns to meetings.
func (a *App) redrawSoon(c *ui.Context) {
	now := a.now()
	if a.timer.Running {
		c.After(time.Second - time.Duration(now.Nanosecond()) + time.Millisecond)
	} else {
		c.After(time.Minute - time.Duration(now.Second())*time.Second - time.Duration(now.Nanosecond()) + time.Millisecond)
	}
}

// shortcuts: Ctrl+N new event, Ctrl+Shift+N new reminder, Ctrl+1…5
// pages, Space starts and pauses the timer on the Focus page.
func (a *App) shortcuts(c *ui.Context) {
	if a.editor.open {
		return
	}
	c.OnShortcut(ui.Ctrl, ui.KeyN, func() { a.newItem(KindEvent, a.nextSlot()) })
	c.OnShortcut(ui.Ctrl|ui.Shift, ui.KeyN, func() { a.newItem(KindReminder, a.nextSlot()) })
	for i, n := range nav {
		id := n.id
		c.OnShortcut(ui.Ctrl, ui.Key1+ui.Key(i), func() { a.page = id })
	}
}

// nextSlot is the next half hour, for new items.
func (a *App) nextSlot() time.Time {
	now := a.now()
	return now.Truncate(30 * time.Minute).Add(30 * time.Minute)
}

func (a *App) titleBar(c *ui.Context, compact bool) {
	bar := c.TitleBar()
	h := max(bar.Height, 52)
	ui.Row(c).Height(h).Shrink(0).Padding(0, bar.Right+12, 0, bar.Left+18).Gap(12).AlignItems(ui.Center).DragWindow().Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Shrink(0).Children(func() {
			ui.Image(c, logoMark).Size(26, 26).Shrink(0).Radius(6).Shadow(0, 2, 6, -2, ui.Hex("#1e40af").Alpha(0.45))
			if !compact {
				ui.Text(c, "ctrl-tower").FontWeight(700).FontSize(14).LetterSpacing(-0.2).SingleLine()
			}
		})
		ui.Box(c).Grow(1).MinWidth(0).MaxWidth(560).Margin(0, ui.Auto).Children(func() {
			a.quickAdd(c, &a.quick, &a.preview, "main")
		})
	})
}

// quickAdd is the line that turns plain English into an item, with what
// it understood shown under it before Enter adds it.
func (a *App) quickAdd(c *ui.Context, line *string, preview **Parsed, key string) {
	p := pal(c)
	in := ui.TextInput(c.Key("quick-"+key), line).
		Placeholder("Add: “Design sync tomorrow 3pm 45m zoom.us/j/123”").Label("Quick add").
		Padding(8, 12).Radius(10).Background(p.surface).Shadow(0, 1, 3, 0, p.shadow)
	if key == "fly" {
		in.AutoFocus() // the panel opens to type into
	}
	if in.Changed() {
		if *line == "" {
			*preview = nil
		} else {
			pr := ParseQuick(*line, a.now(), a.data.Settings.DefaultLead)
			*preview = &pr
		}
	}
	if in.Submitted() && *preview != nil && (*preview).Item.Title != "" {
		it := a.addItem((*preview).Item)
		c.ToastAction("Added “"+clampText(it.Title, 40)+"”", "Undo", func() { a.deleteItem(it.ID) })
		*line, *preview = "", nil
	}
	if *preview == nil {
		return
	}
	pr := **preview
	ui.Overlay(c, func() {
		box := pane(c, ui.Column(c.Key("preview-"+key))).Background(p.raised).Padding(12, 14).Gap(8).Radius(12).
			AttachTo(in, ui.AnchorBottomLeft, ui.AnchorTopLeft).Margin(6, 0, 0, 0).Width(in.Bounds().W).
			Shadow(0, 12, 32, -10, p.lift)
		riseIn(box, -4, 180*time.Millisecond, 0)
		box.Children(func() { a.previewBody(c, pr, func() { a.newFromParsed(pr); *line, *preview = "", nil }) })
	})
	if in.OverlayShortcut(0, ui.KeyEscape) {
		*line, *preview = "", nil
	}
}

// previewBody shows a parsed item as it will be added.
func (a *App) previewBody(c *ui.Context, pr Parsed, details func()) {
	p := pal(c)
	it := pr.Item
	now := a.now()
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ic, tint := icToday, p.accentInk
		if it.Kind == KindReminder {
			ic, tint = icBell, p.warn
		}
		ui.Icon(c, ic).Size(15, 15).TextColor(tint)
		title := it.Title
		if title == "" {
			title = "Untitled"
		}
		ui.Text(c, title).FontWeight(600).MaxLines(2).Grow(1).MinWidth(0)
	})
	ui.Row(c).Gap(6).Wrap().Children(func() {
		when := dayLabel(it.Start, now) + " · " + clock(it.Start)
		chip(c, nil, when, p.accentInk)
		if it.Kind == KindEvent {
			chip(c, nil, minutesLabel(it.Minutes), p.muted)
		} else {
			chip(c, icBell, "Reminder", p.warn)
		}
		if it.URL != "" {
			name := meetingService(it.URL)
			if name == "" {
				name = "Link"
			}
			chip(c, icVideo, name, p.accentInk)
		}
		if it.Location != "" {
			chip(c, icPin, clampText(it.Location, 40), p.muted)
		}
		if it.Repeat.Freq != FreqNone {
			chip(c, icRepeat, repeatLabel(it.Repeat, it.Start), p.muted)
		}
	})
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "Enter to add · Esc to clear").FontSize(11.5).TextColor(p.faint).Grow(1)
		ui.Link(c.Key("details"), "More details…", "").FontSize(12).OnClick(details)
	})
}

func (a *App) navRail(c *ui.Context, compact bool) {
	p := pal(c)
	width := float32(196)
	if compact {
		width = 56
	}
	rail := pane(c, ui.Column(c)).Width(width).Shrink(0).Padding(8).Gap(2)
	// The chosen page's pill is painted under the buttons, where the
	// chosen one was last frame, and slides there when the page changes.
	var y, h float32
	rail.Children(func() {
		for _, n := range nav {
			chosen := a.page == n.id
			b := ui.ButtonBase(c.Key(n.id)).Row().Gap(10).Padding(9, 10).Radius(10).AlignItems(ui.Center).
				Label(n.label).Selected(chosen)
			if compact {
				b.Tooltip(n.label).Justify(ui.Center)
			}
			bg, fg, icFg := p.hover.Alpha(0), p.muted, p.muted
			switch {
			case chosen:
				fg, icFg = p.text, p.accentInk
				bb, rb := b.LayoutBounds(), rail.LayoutBounds()
				y, h = bb.Y-rb.Y, bb.H
			case b.Hovered():
				bg, fg, icFg = p.hover, p.text, p.text
			}
			b.Background(b.AnimateColor("bg", bg, colorAnim)).TextColor(b.AnimateColor("fg", fg, colorAnim))
			pressScale(b, 0.97)
			icFg = b.AnimateColor("ic", icFg, colorAnim)
			b.Children(func() {
				ui.Icon(c, n.icon).Size(17, 17).TextColor(icFg)
				if !compact {
					weight := 500
					if chosen {
						weight = 600
					}
					ui.Text(c, n.label).FontSize(13.5).FontWeight(weight).SingleLine().Grow(1)
					if n.id == pageFocus && a.timer.Running {
						ui.Text(c, countdown(a.timer.Remaining(a.now()))).FontSize(11.5).FontWeight(600).FontFeatures("tnum").
							TextColor(a.ringColor(c)).ContentTransition(ui.ContentTransition{Numeric: true, Down: true})
					}
				}
			})
			if b.Clicked() {
				a.page = n.id
			}
		}
		ui.Spacer(c)
		var add ui.Element
		if compact {
			add = iconButton(c.Key("new-compact"), icPlus, "New event").Background(p.accent).TextColor(p.accentText).Size(40, 40).Radius(12).
				Shadow(0, 3, 10, -3, p.accent.Alpha(0.45))
		} else {
			add = pillButton(c.Key("new"), icPlus, "New", true)
		}
		add.Tooltip("New event (Ctrl+N)")
		add.OnClick(func() { a.newItem(KindEvent, a.nextSlot()) })
	})
	if h > 0 {
		y, h = rail.AnimateTo("y", y, slide), rail.AnimateTo("h", h, slide)
		soft, bar := p.accentSoft, p.accent
		rail.Draw(func(pt *ui.Painter, r ui.Rect) {
			pill := ui.Rect{X: r.X + 8, Y: r.Y + y, W: r.W - 16, H: h}
			pt.Fill(pill, soft, 10)
			pt.Fill(ui.Rect{X: pill.X, Y: pill.Y + h/2 - 8, W: 3, H: 16}, bar, 1.5)
		})
	}
}

// Today.

func (a *App) todayPage(c *ui.Context, compact bool) {
	now := a.now()
	day := startOfDay(now)
	occs := a.occurrences(day, day.AddDate(0, 0, 1))
	side := func() {
		a.nextUpCard(c, now)
		a.focusCard(c, now)
	}
	if compact {
		ui.Scroll(c).Fill().Gap(12).Children(func() {
			side()
			a.timeline(c, now, occs, false)
		})
		return
	}
	ui.Row(c).Fill().Gap(12).AlignItems(ui.Stretch).Children(func() {
		pane(c, ui.Column(c)).Grow(1).MinWidth(0).Children(func() {
			ui.Scroll(c).Fill().Padding(18, 18, 18, 18).Children(func() { a.timeline(c, now, occs, true) })
		})
		ui.Scroll(c).Width(340).Shrink(0).Gap(12).Children(side)
	})
}

func (a *App) timeline(c *ui.Context, now time.Time, occs []Occurrence, header bool) {
	p := pal(c)
	col := ui.Column(c).Gap(4)
	if !header {
		pane(c, col).Padding(16)
	}
	col.Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Baseline).Padding(0, 0, 10, 0).Children(func() {
			ui.Text(c, now.Format("Monday")).FontSize(24).Bold().LetterSpacing(-0.4)
			ui.Text(c, now.Format("2 January")).FontSize(15).FontWeight(500).TextColor(p.muted).Grow(1)
			if len(occs) > 0 {
				chip(c, nil, plural(len(occs), "item", "items"), p.accentInk)
			}
		})
		if len(occs) == 0 {
			a.emptyDay(c)
			return
		}
		nowShown := false
		for i, o := range occs {
			if !nowShown && !o.Start.Before(now) {
				nowLine(c, now)
				nowShown = true
			}
			a.occRow(c, o, now, i)
		}
		if !nowShown {
			nowLine(c, now)
		}
	})
}

// emptyState is what a list shows with nothing in it: an icon in a soft
// accent disc, a line, and what to do about it.
func emptyState(c *ui.Context, ic *ui.SVG, title, hint string) {
	p := pal(c)
	col := ui.Column(c).Padding(40, 12).Gap(8).AlignItems(ui.Center)
	riseIn(col, 8, 320*time.Millisecond, 60*time.Millisecond)
	col.Children(func() {
		ui.Box(c).Size(56, 56).Radius(28).Center().Margin(0, 0, 6, 0).
			RadialGradient(ui.RadialGradient{Stops: []ui.GradientStop{{Color: p.accent.Alpha(0.22), At: 0}, {Color: p.accent.Alpha(0.08), At: 1}}}).
			Children(func() { ui.Icon(c, ic).Size(26, 26).TextColor(p.accentInk) })
		ui.Text(c, title).FontWeight(600).FontSize(15)
		if hint != "" {
			ui.Text(c, hint).FontSize(13).TextColor(p.muted).TextAlign(ui.Center).MaxWidth(360)
		}
	})
}

func (a *App) emptyDay(c *ui.Context) {
	emptyState(c, icSparkle, "Nothing scheduled today",
		"Type into the bar above, e.g. “Call mum at 6pm” or “Standup every weekday 9:30”.")
}

func nowLine(c *ui.Context, now time.Time) {
	p := pal(c)
	ui.Row(c.Key("now")).Gap(8).AlignItems(ui.Center).Padding(4, 0).Transition(ui.ElementTransition{Position: true, Animation: &slide}).Children(func() {
		ui.Text(c, clock(now)).FontSize(11).FontWeight(700).TextColor(p.danger).Width(64).TextAlign(ui.End).FontFeatures("tnum").Shrink(0)
		ui.Box(c).Size(14, 14).Radius(7).Center().Background(p.danger.Alpha(0.2)).Shrink(0).Children(func() {
			ui.Box(c).Size(8, 8).Radius(4).Background(p.danger)
		})
		ui.Box(c).Height(2).Grow(1).Radius(1).LinearGradient(ui.LinearGradient{Angle: 90, From: p.danger, To: p.danger.Alpha(0.15)})
	})
}

// occRow is one occurrence in a list, the ith: its time, a card with its
// title and where, and Join while joining makes sense. Rows rise in one
// after another as the list first shows.
func (a *App) occRow(c *ui.Context, o Occurrence, now time.Time, i int) {
	p := pal(c)
	past := !o.End.After(now) && !(o.End.Equal(o.Start) && o.Start.After(now.Add(-time.Minute)))
	live := o.IsEvent() && !o.Start.After(now) && o.End.After(now)
	line := ui.Row(c.Key(o.ID())).Gap(12).AlignItems(ui.Start).Transition(rowMotion)
	riseIn(line, 8, 280*time.Millisecond, stagger(i))
	line.Children(func() {
		timeFg := p.muted
		if live {
			timeFg = p.accentInk
		}
		ui.Column(c).Width(64).Shrink(0).Padding(11, 0, 0, 0).AlignItems(ui.End).Children(func() {
			ui.Text(c, clock(o.Start)).FontSize(12.5).FontWeight(600).FontFeatures("tnum").SingleLine().TextColor(timeFg)
			if o.IsEvent() {
				ui.Text(c, minutesLabel(int(o.End.Sub(o.Start).Minutes()))).FontSize(11).TextColor(p.faint).SingleLine()
			}
		})
		row := ui.ButtonBase(c).Row().Grow(1).MinWidth(0).Padding(10, 12).Gap(10).Radius(12).AlignItems(ui.Start).
			Label(o.Title).Description(span(o))
		tint := p.accent
		if !o.IsEvent() {
			tint = p.warn
		}
		bg, edge := p.hover.Alpha(0), p.rim.Alpha(0)
		if live {
			bg, edge = p.accentSoft, p.accent.Alpha(0.35)
		}
		if row.Hovered() {
			bg, edge = tint.Alpha(0.09), tint.Alpha(0.22)
			if live {
				bg = tint.Alpha(0.22)
			}
		}
		row.Background(row.AnimateColor("bg", bg, colorAnim)).Border(1, row.AnimateColor("edge", edge, colorAnim))
		pressScale(row, 0.99)
		if past {
			row.Opacity(0.6)
		}
		row.OnClick(func() { a.editOccurrence(o) })
		row.Children(func() {
			ui.Box(c).Width(3).Height(36).Radius(2).Background(tint).Shrink(0)
			ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
				ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
					if !o.IsEvent() {
						ui.Icon(c, icBell).Size(13, 13).TextColor(p.warn).Shrink(0).Margin(3, 0, 0, 0)
					}
					ui.Text(c, o.Title).FontWeight(600).FontSize(14).MaxLines(2).MinWidth(0).Grow(1)
					if live {
						chip(c, nil, "Now", p.accentInk)
					}
				})
				a.occMeta(c, o)
			})
			a.occAction(c, o, now)
		})
	})
}

// occMeta is the line under a title: the times, where, and how it repeats.
func (a *App) occMeta(c *ui.Context, o Occurrence) {
	p := pal(c)
	ui.Row(c).Gap(6).Wrap().AlignItems(ui.Center).Children(func() {
		ui.Text(c, span(o)).FontSize(12).TextColor(p.muted).FontFeatures("tnum").SingleLine()
		if o.URL != "" {
			name := meetingService(o.URL)
			if name == "" {
				name = hostOf(o.URL)
			}
			ui.Row(c).Gap(4).AlignItems(ui.Center).MinWidth(0).Shrink(1).Children(func() {
				ui.Icon(c, icVideo).Size(12, 12).TextColor(p.muted).Shrink(0)
				ui.Text(c, name).FontSize(12).TextColor(p.muted).SingleLine().Ellipsis("…").MinWidth(0).Tooltip(o.URL)
			})
		}
		if o.Location != "" {
			ui.Row(c).Gap(4).AlignItems(ui.Center).MinWidth(0).Shrink(1).Children(func() {
				ui.Icon(c, icPin).Size(12, 12).TextColor(p.muted).Shrink(0)
				ui.Text(c, o.Location).FontSize(12).TextColor(p.muted).SingleLine().Ellipsis("…").MinWidth(0).Tooltip(o.Location)
			})
		}
		if o.Item.Repeat.Freq != FreqNone {
			ui.Icon(c, icRepeat).Size(12, 12).TextColor(p.faint).Shrink(0).Tooltip(repeatLabel(o.Item.Repeat, o.Item.Start))
		}
	})
}

// occAction is Join from ten minutes before an online meeting until it
// ends, or Directions for one in person.
func (a *App) occAction(c *ui.Context, o Occurrence, now time.Time) {
	if !o.IsEvent() || !o.End.After(now) {
		return
	}
	soon := !o.Start.After(now.Add(10 * time.Minute))
	switch {
	case o.URL != "" && soon:
		pillButton(c.Key("join"), icVideo, "Join", true).OnClick(func() { a.join(o) })
	case o.URL != "":
		iconButton(c.Key("open"), icExternal, "Open meeting link").OnClick(func() { a.join(o) })
	case o.Location != "":
		iconButton(c.Key("map"), icPin, "Directions").OnClick(func() { a.join(o) })
	}
}

// nextUpCard shows the event under way or next, large, with Join.
func (a *App) nextUpCard(c *ui.Context, now time.Time) {
	p := pal(c)
	o, ok := a.nextEvent(now)
	box := card(c)
	if ok {
		// The card that matters most carries the accent: a wash of it
		// from the top corner, fading into the pane.
		box.LinearGradient(ui.LinearGradient{Angle: 160, From: p.accent.Alpha(0.14).Over(p.surface), To: p.surface, End: 0.7, Oklab: true}).
			Border(1, p.accent.Alpha(0.22))
	}
	box.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			heading(c, "NEXT UP").Grow(1)
			if ok && !o.Start.After(now) {
				chip(c, nil, "Live", p.danger)
			}
		})
		if !ok {
			ui.Text(c, "No meetings in the next two days.").TextColor(p.muted).FontSize(13)
			return
		}
		live := !o.Start.After(now)
		ui.Text(c, o.Title).FontSize(19).Bold().LetterSpacing(-0.2).MaxLines(3).ContentTransition(ui.ContentTransition{})
		when := until(o.Start, now)
		if live {
			when = "Started " + until(o.Start, now) + " · ends " + clock(o.End)
		} else {
			when = dayLabel(o.Start, now) + " · " + span(o) + " · " + when
		}
		ui.Text(c, when).FontSize(12.5).TextColor(p.muted).FontFeatures("tnum")
		if o.Location != "" {
			ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
				ui.Icon(c, icPin).Size(13, 13).TextColor(p.muted).Shrink(0).Margin(2, 0, 0, 0)
				ui.Text(c, o.Location).FontSize(13).MaxLines(2).MinWidth(0)
			})
		}
		ui.Row(c).Gap(8).Wrap().Padding(4, 0, 0, 0).Children(func() {
			switch {
			case o.URL != "":
				label := "Join"
				if s := meetingService(o.URL); s != "" {
					label = "Join " + s
				}
				pillButton(c.Key("join"), icVideo, label, true).OnClick(func() { a.join(o) })
			case o.Location != "":
				pillButton(c.Key("dir"), icPin, "Directions", true).OnClick(func() { a.join(o) })
			}
			pillButton(c.Key("edit"), nil, "Details", false).OnClick(func() { a.editOccurrence(o) })
		})
	})
}
