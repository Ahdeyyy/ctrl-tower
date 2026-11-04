package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// ringColor is the timer's color: the accent while focusing, green on a
// break. It fades between them as the phase changes.
func (a *App) ringColor(c *ui.Context) ui.Color {
	p := pal(c)
	if a.timer.Phase == PhaseWork {
		return p.accent
	}
	return p.breakTint
}

// ring draws the timer's progress as an arc from twelve o'clock, the
// accent brightening towards its head, where a dot marks now. progress
// eases to each new value: a second's step glides, and a reset or skip
// winds back instead of jumping.
func ring(c *ui.Context, size, stroke float32, progress float64, color ui.Color) ui.Element {
	p := pal(c)
	box := ui.Box(c).Size(size, size).Shrink(0).Center()
	frac := box.AnimateTo("progress", float32(math.Max(0, math.Min(1, progress))), ui.Animation{Duration: 600 * time.Millisecond, Ease: easeOut})
	track, head := p.track, color.Mix(ui.RGB(255, 255, 255), 0.35)
	box.Draw(func(pt *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		rad := min(r.W, r.H)/2 - stroke/2 - 3
		pt.StrokePath((&ui.Path{}).Circle(cx, cy, rad), stroke, track)
		if frac <= 0.001 {
			return
		}
		arc := &ui.Path{}
		steps := max(2, int(160*frac))
		var hx, hy float32
		for i := 0; i <= steps; i++ {
			ang := -math.Pi/2 + 2*math.Pi*float64(frac)*float64(i)/float64(steps)
			hx = cx + rad*float32(math.Cos(ang))
			hy = cy + rad*float32(math.Sin(ang))
			if i == 0 {
				arc.MoveTo(hx, hy)
			} else {
				arc.LineTo(hx, hy)
			}
		}
		pt.StrokePath(arc, stroke+8, color.Alpha(0.14)) // a soft glow under the arc
		// The gradient turns around the ring's center, which it places in
		// the arc's bounds.
		b := arc.Bounds()
		if b.W < 1 || b.H < 1 {
			pt.StrokePath(arc, stroke, color)
			return
		}
		pt.StrokePathConicGradient(arc, stroke, ui.ConicGradient{Center: ui.Point{X: (cx - b.X) / b.W, Y: (cy - b.Y) / b.H}, Oklab: true, Stops: []ui.GradientStop{
			{Color: color.Alpha(0.75), At: 0}, {Color: color, At: max(0.01, frac*0.6)}, {Color: head, At: max(0.02, frac)},
		}})
		dot := stroke * 0.32
		pt.FillPath((&ui.Path{}).Circle(hx, hy, dot), ui.RGB(255, 255, 255).Alpha(0.95))
	})
	return box
}

func (a *App) phaseLabel() string {
	t := a.timer
	switch {
	case t.Phase != PhaseWork:
		return t.Phase.String()
	case t.Fresh():
		return "Ready to focus"
	case !t.Running:
		return "Paused"
	}
	return "Focusing"
}

// focusPage is the timer, large.
func (a *App) focusPage(c *ui.Context) {
	p := pal(c)
	now := a.now()
	s := &a.data.Settings
	ui.Scroll(c).Fill().Children(func() {
		pane(c, ui.Column(c)).Padding(28, 24).Gap(18).AlignItems(ui.Center).MinHeight(560).Children(func() {
			names := make([]string, len(s.Presets))
			for i, pr := range s.Presets {
				names[i] = pr.Name
			}
			chosen := s.Preset
			if segmented(c.Key("preset"), &chosen, names...).Label("Preset").Changed() {
				a.choosePreset(chosen)
			}
			pr := a.preset()
			ui.Text(c, minutesLabel(pr.Work)+" focus · "+minutesLabel(pr.Break)+" break · long break every "+plural(pr.LongEvery, "session", "sessions")).
				FontSize(12).TextColor(p.muted).TextAlign(ui.Center)

			col := a.ringColor(c)
			holder := ui.Box(c).Size(280, 280).Shrink(0)
			col = holder.AnimateColor("ring", col, ui.Animation{Duration: 250 * time.Millisecond, Ease: ui.EaseInOut})
			// A pool of the phase's color behind the ring, brighter while it runs.
			glow := float32(0.07)
			if a.timer.Running {
				glow = 0.13
			}
			glow = holder.AnimateTo("glow", glow, ui.Animation{Duration: 400 * time.Millisecond, Ease: easeOut})
			holder.Radius(140).RadialGradient(ui.RadialGradient{Radius: 140, Stops: []ui.GradientStop{{Color: col.Alpha(glow), At: 0}, {Color: col.Alpha(glow * 0.4), At: 0.7}, {Color: col.Alpha(0), At: 1}}})
			holder.Children(func() {
				ring(c, 280, 12, a.timer.Progress(now), col).Absolute().Top(0).Left(0)
				ui.Column(c).Fill().Center().Gap(4).Children(func() {
					ui.Text(c, a.phaseLabel()).FontSize(13).FontWeight(600).TextColor(col).ContentTransition(ui.ContentTransition{})
					ui.Text(c, countdown(a.timer.Remaining(now))).FontSize(60).FontWeight(300).LetterSpacing(-1).FontFeatures("tnum").Label("Time left").
						ContentTransition(ui.ContentTransition{Numeric: true, Down: a.timer.Running})
					if a.timer.Running {
						ui.Text(c, "ends "+clock(a.timer.EndsAt)).FontSize(12).TextColor(p.muted).FontFeatures("tnum").Transition(ui.ElementTransition{Enter: &ui.Motion{Opacity: 0, Y: -4}, Exit: &ui.Motion{Opacity: 0, Y: -4}})
					}
				})
			})
			a.cycleDots(c)
			a.timerControls(c, 64)
			a.conflictBanner(c, now, false)
			a.tagField(c)
			if n := len(a.held); n > 0 {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, icMoon).Size(14, 14).TextColor(p.muted)
					ui.Text(c, plural(n, "reminder", "reminders")+" held until your break").FontSize(12.5).TextColor(p.muted)
				})
			} else if s.FocusDND && a.focusing() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, icMoon).Size(14, 14).TextColor(p.faint)
					ui.Text(c, "Do Not Disturb: reminders wait, meetings still come through").FontSize(12).TextColor(p.faint)
				})
			}
		})
	})
	if !a.editor.open {
		c.OnShortcut(0, ui.KeySpace, a.startPause)
	}
}

// cycleDots shows the sessions done before the long break: done ones
// filled, the one under way a longer pill, the rest the track's color.
func (a *App) cycleDots(c *ui.Context) {
	p := pal(c)
	every := a.preset().LongEvery
	col := a.ringColor(c)
	ui.Row(c).Gap(6).AlignItems(ui.Center).Label(fmt.Sprintf("%d of %d sessions before the long break", a.timer.Done, every)).Children(func() {
		for i := range every {
			dot := ui.Box(c.Key(i)).Height(8).Radius(4)
			w, bg := float32(8), p.track.Mix(p.faint, 0.35)
			switch {
			case i < a.timer.Done:
				bg = col
			case i == a.timer.Done && a.timer.Phase == PhaseWork:
				w, bg = 22, col.Alpha(0.45)
			}
			dot.Width(dot.AnimateTo("w", w, slide)).Background(dot.AnimateColor("bg", bg, colorAnim))
		}
	})
}

// playButton starts and pauses the timer: a disc of the phase's color
// that swells under the pointer, gives under a press, and turns its icon
// over as it changes.
func (a *App) playButton(c *ui.Context, size float32, label string) ui.Element {
	p := pal(c)
	ic := icPlay
	if a.timer.Running {
		ic = icPause
	}
	tip := label
	if size >= 48 {
		tip += " (Space)" // the large one is on the Focus page, where Space is
	}
	b := ui.ButtonBase(c).Size(size, size).Shrink(0).Radius(999).Center().Label(label).Tooltip(tip)
	col := b.AnimateColor("col", a.ringColor(c), ui.Animation{Duration: 250 * time.Millisecond, Ease: ui.EaseInOut})
	bg := col
	scale, lift := float32(1), float32(0)
	switch {
	case b.Pressed():
		bg, scale = col.Mix(ui.RGB(0, 0, 0), 0.15), 0.93
	case b.Hovered():
		bg, scale, lift = col.Mix(ui.RGB(255, 255, 255), 0.1), 1.04, 1
	}
	lift = b.AnimateTo("lift", lift, colorAnim)
	b.AnimatePose(ui.Pose{Scale: scale}, ui.Animation{Spring: ui.SpringOptions{Duration: 260 * time.Millisecond, Bounce: 0.25}})
	b.LinearGradient(ui.LinearGradient{Angle: 160, From: bg.Mix(ui.RGB(255, 255, 255), 0.14), To: bg, Oklab: true}).
		TextColor(p.accentText).Shadow(0, size*0.08+3*lift, size*0.25+8*lift, -size*0.06, col.Alpha(0.4+0.15*lift))
	b.Children(func() { ui.Icon(c, ic).Size(size*0.36, size*0.36).ContentTransition(ui.ContentTransition{}) })
	b.OnClick(a.startPause)
	return b
}

func (a *App) playLabel() string {
	switch {
	case a.timer.Running:
		return "Pause"
	case !a.timer.Fresh():
		return "Resume"
	}
	return "Start"
}

// timerControls: reset, start or pause, skip.
func (a *App) timerControls(c *ui.Context, size float32) {
	ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
		iconButton(c.Key("reset"), icReset, "Reset").Size(size*0.62, size*0.62).Radius(999).OnClick(a.resetTimer)
		a.playButton(c.Key("play"), size, a.playLabel())
		iconButton(c.Key("skip"), icSkip, "Skip to "+strings.ToLower(a.nextPhaseName())).Size(size*0.62, size*0.62).Radius(999).OnClick(a.skip)
	})
}

func (a *App) nextPhaseName() string {
	if a.timer.Phase != PhaseWork {
		return "Focus"
	}
	if a.timer.Done+1 >= a.preset().LongEvery {
		return "Long break"
	}
	return "Break"
}

// tagField names what the session is for, with recent tags a click away.
func (a *App) tagField(c *ui.Context) {
	p := pal(c)
	ui.Column(c).Gap(8).Width(360).MaxWidthPercent(100).Children(func() {
		ui.TextInput(c.Key("tag"), &a.timer.Tag).Placeholder("What are you working on? (task, book…)").Label("Session tag")
		if len(a.timer.Tag) > maxTag {
			a.timer.Tag = clampText(a.timer.Tag, maxTag)
		}
		if len(a.data.Tags) == 0 {
			return
		}
		ui.Row(c).Gap(6).Wrap().Children(func() {
			for i, tag := range a.data.Tags {
				if i >= 8 {
					break
				}
				b := ui.ButtonBase(c.Key(tag)).Row().Gap(4).Padding(3, 10).Radius(999).AlignItems(ui.Center).Label(tag).MaxWidth(180)
				bg, edge, fg := p.surface, p.rim, p.muted
				switch {
				case tag == a.timer.Tag:
					bg, edge, fg = p.accentSoft.Over(p.surface), p.accent.Alpha(0.45), p.accentInk
				case b.Hovered():
					bg, edge, fg = p.hover.Over(p.surface), p.accent.Alpha(0.25), p.text
				}
				b.Background(b.AnimateColor("bg", bg, colorAnim)).Border(1, b.AnimateColor("edge", edge, colorAnim)).
					TextColor(b.AnimateColor("fg", fg, colorAnim))
				pressScale(b, 0.95)
				b.Children(func() {
					ui.Icon(c, icTag).Size(11, 11).Shrink(0)
					ui.Text(c, tag).FontSize(12).FontWeight(500).SingleLine().Ellipsis("…").MinWidth(0)
				})
				b.OnClick(func() { a.timer.Tag = tag })
			}
		})
	})
}

// conflictBanner offers to end the session before a meeting it would run
// into.
func (a *App) conflictBanner(c *ui.Context, now time.Time, compact bool) {
	p := pal(c)
	o, fit, ok := a.conflict(now)
	if !ok {
		return
	}
	ui.Column(c.Key("conflict")).Gap(8).Padding(10, 12).Radius(12).Background(p.warn.Alpha(0.12)).
		Border(1, p.warn.Alpha(0.4)).MaxWidth(440).Transition(rowMotion).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Start).Children(func() {
			ui.Icon(c, icBell).Size(14, 14).TextColor(p.warn).Shrink(0).Margin(2, 0, 0, 0)
			ui.Text(c, "“"+clampText(o.Title, 60)+"” starts at "+clock(o.Start)+", before this session ends.").
				FontSize(12.5).FontWeight(500).MaxLines(3).MinWidth(0)
		})
		ui.Row(c).Gap(8).Wrap().Children(func() {
			if fit >= 5*time.Minute {
				pillButton(c.Key("shorten"), nil, "Shorten to "+minutesLabel(int(fit.Minutes())), true).
					OnClick(func() { a.shortenFor(o, fit) })
			}
			pillButton(c.Key("keep"), nil, "Keep full length", false).OnClick(func() { a.conflictSeen = o.ID() })
		})
	})
}

// focusCard is the timer, small, for Today.
func (a *App) focusCard(c *ui.Context, now time.Time) {
	p := pal(c)
	card(c).Children(func() {
		ui.Row(c).Children(func() {
			heading(c, "FOCUS").Grow(1)
			ui.Link(c.Key("open-focus"), "Open", "").FontSize(12).FontWeight(600).OnClick(func() { a.page = pageFocus })
		})
		ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
			holder := ui.Box(c).Size(76, 76).Shrink(0)
			ringCol := holder.AnimateColor("ring", a.ringColor(c), ui.Animation{Duration: 250 * time.Millisecond, Ease: ui.EaseInOut})
			holder.Children(func() {
				ring(c, 76, 6, a.timer.Progress(now), ringCol).Absolute().Top(0).Left(0)
				ui.Column(c).Fill().Center().Children(func() {
					ui.Text(c, countdown(a.timer.Remaining(now))).FontSize(15).FontWeight(600).FontFeatures("tnum").
						ContentTransition(ui.ContentTransition{Numeric: true, Down: a.timer.Running})
				})
			})
			ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
				ui.Text(c, a.phaseLabel()).FontWeight(600).TextColor(ringCol).ContentTransition(ui.ContentTransition{})
				sub := a.preset().Name
				if a.timer.Tag != "" {
					sub += " · " + a.timer.Tag
				}
				ui.Text(c, sub).FontSize(12).TextColor(p.muted).SingleLine().Ellipsis("…").MinWidth(0)
			})
			a.miniControls(c)
		})
		a.conflictBanner(c, now, true)
	})
}

// miniControls: start or pause, and skip, as icons.
func (a *App) miniControls(c *ui.Context) {
	ui.Row(c).Gap(4).Shrink(0).AlignItems(ui.Center).Children(func() {
		label := "Start"
		if a.timer.Running {
			label = "Pause"
		}
		a.playButton(c.Key("play"), 36, label)
		iconButton(c.Key("skip"), icSkip, "Skip").Radius(999).Size(36, 36).OnClick(a.skip)
	})
}
