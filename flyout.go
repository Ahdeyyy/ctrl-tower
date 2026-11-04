package main

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// flyoutView is the tray panel: the next meeting, the timer and
// quick-add. It opens many times a day, so nothing in it animates in.
func (a *App) flyoutView(c *ui.Context) {
	appTheme(c)
	backdrop(c)
	a.redrawSoon(c)
	p := pal(c)
	now := a.now()
	ui.Column(c).Fill().Padding(14).Gap(12).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Image(c, logoMark).Size(22, 22).Shrink(0).Radius(5)
			ui.Text(c, now.Format("Monday 2 January")).FontWeight(600).FontSize(13).Grow(1)
			ui.Text(c, clock(now)).FontSize(12.5).FontWeight(500).TextColor(p.muted).FontFeatures("tnum").ContentTransition(ui.ContentTransition{Numeric: true})
		})
		ui.Box(c).Children(func() {
			in := &a.flyQuick
			a.quickAdd(c, in, &a.flyPreview, "fly")
		})
		a.nextUpCard(c, now)
		card(c).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				holder := ui.Box(c).Size(56, 56).Shrink(0)
				col := holder.AnimateColor("ring", a.ringColor(c), ui.Animation{Duration: 250 * time.Millisecond, Ease: ui.EaseInOut})
				holder.Children(func() {
					ring(c, 56, 5, a.timer.Progress(now), col).Absolute().Top(0).Left(0)
				})
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					ui.Text(c, countdown(a.timer.Remaining(now))).FontSize(22).FontWeight(600).FontFeatures("tnum").
						ContentTransition(ui.ContentTransition{Numeric: true, Down: a.timer.Running})
					ui.Text(c, a.phaseLabel()+" · "+a.preset().Name).FontSize(12).TextColor(p.muted).SingleLine().Ellipsis("…").MinWidth(0)
				})
				a.miniControls(c)
			})
			a.conflictBanner(c, now, true)
		})
		ui.Spacer(c)
		ui.Row(c).Gap(8).Children(func() {
			pillButton(c.Key("open"), nil, "Open ctrl-tower", false).OnClick(a.h.showMain)
			ui.Spacer(c)
			pillButton(c.Key("new"), icPlus, "New", true).OnClick(func() {
				a.h.showMain()
				a.newItem(KindEvent, a.nextSlot())
			})
		})
	})
}

// overlayView is the panel that comes up over everything as a meeting
// starts: Join, Snooze and Dismiss, for each meeting starting.

func (a *App) overlayView(c *ui.Context) {
	appTheme(c)
	backdrop(c)
	a.redrawSoon(c)
	p := pal(c)
	now := a.now()
	root := ui.Column(c).Fill().Padding(16).Gap(12)
	popIn(root)
	root.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).DragWindow().Children(func() {
			livePulse(c, p.danger)
			label := "Starting now"
			if len(a.ringing) > 1 {
				label = plural(len(a.ringing), "meeting", "meetings") + " starting"
			}
			ui.Text(c, strings.ToUpper(label)).FontSize(11.5).FontWeight(700).TextColor(p.danger).LetterSpacing(0.7).Grow(1)
		})
		ui.Scroll(c).Grow(1).Gap(10).Children(func() {
			for _, o := range a.ringing {
				ui.Column(c.Key(o.ID())).Gap(6).Transition(rowMotion).Children(func() {
					ui.Text(c, o.Title).FontSize(19).Bold().LetterSpacing(-0.2).MaxLines(2)
					when := span(o)
					if o.Start.Before(now.Add(-time.Minute)) {
						when += " · started " + until(o.Start, now)
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
							pillButton(c.Key("join"), icVideo, label, true).OnClick(func() { a.join(o); a.dismiss(o) })
						case o.Location != "":
							pillButton(c.Key("dir"), icPin, "Directions", true).OnClick(func() { a.join(o); a.dismiss(o) })
						}
						pillButton(c.Key("snooze"), nil, "Snooze 5 min", false).OnClick(func() { a.snoozeFor(o, 5*time.Minute) })
						pillButton(c.Key("dismiss"), nil, "Dismiss", false).OnClick(func() { a.dismiss(o) })
					})
				})
			}
		})
	})
	c.OnShortcut(0, ui.KeyEscape, func() {
		for _, o := range append([]Occurrence(nil), a.ringing...) {
			a.dismiss(o)
		}
	})
}

// livePulse is a dot that sends out a ring every second and a half, to
// say something is happening now. Only the overlay, up for a minute or
// two, keeps one going.
func livePulse(c *ui.Context, color ui.Color) {
	box := ui.Box(c).Size(14, 14).Center().Shrink(0)
	v := box.Keyframes("pulse", []ui.Keyframe{{Value: 0}, {At: 1100 * time.Millisecond, Value: 1, Ease: ui.EaseOut}},
		ui.KeyframeOptions{Repeat: ui.RepeatForever, RepeatDelay: 400 * time.Millisecond})
	box.Draw(func(pt *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		pt.FillPath((&ui.Path{}).Circle(cx, cy, 3+4*v), color.Alpha(0.35*(1-v)))
	})
	box.Children(func() { ui.Box(c).Size(8, 8).Radius(4).Background(color) })
}
