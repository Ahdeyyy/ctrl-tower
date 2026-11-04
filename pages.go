package main

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Upcoming: the agenda of the days ahead, grouped by day.

func (a *App) upcomingPage(c *ui.Context) {
	p := pal(c)
	now := a.now()
	from := startOfDay(now)
	to := from.AddDate(0, 0, a.upcoming)
	occs := a.occurrences(now, to)
	ui.Scroll(c).Fill().Children(func() {
		pane(c, ui.Column(c)).Padding(18).Gap(4).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Baseline).Padding(0, 0, 8, 0).Children(func() {
				ui.Text(c, "Upcoming").FontSize(24).Bold().LetterSpacing(-0.4).Grow(1)
				ui.Text(c, fmt.Sprintf("Next %d days · %s", a.upcoming, plural(len(occs), "item", "items"))).
					FontSize(12.5).TextColor(p.muted).Shrink(0)
			})
			if len(occs) == 0 {
				emptyState(c, icUpcoming, "Your next "+fmt.Sprint(a.upcoming)+" days are clear", "")
			}
			// A long agenda builds only the first few hundred rows; the
			// rest is a click away.
			const cap = 300
			var last time.Time
			for i, o := range occs {
				if i == cap {
					ui.Text(c, fmt.Sprintf("%s more not shown", thousands(len(occs)-cap))).FontSize(12.5).TextColor(p.muted).Padding(8, 0)
					break
				}
				if i == 0 || !sameDay(o.Start, last) {
					head := ui.Row(c.Key("day-"+dateKey(o.Start))).Padding(16, 0, 6, 0).Gap(8).AlignItems(ui.Center)
					riseIn(head, 8, 280*time.Millisecond, stagger(i))
					head.Children(func() {
						l := dayLabel(o.Start, now)
						fg := p.text
						if l == "Today" {
							fg = p.accentInk
						}
						ui.Text(c, l).FontWeight(700).FontSize(14).TextColor(fg).SingleLine()
						if l == "Today" || l == "Tomorrow" || len(l) < 10 {
							ui.Text(c, o.Start.Format("2 Jan")).FontSize(12).FontWeight(500).TextColor(p.muted).SingleLine()
						}
						ui.Box(c).Height(1).Grow(1).Background(p.divider)
					})
					last = o.Start
				}
				a.occRow(c, o, now, i)
			}
			ui.Row(c).Padding(12, 0, 0, 0).Justify(ui.Center).Children(func() {
				pillButton(c.Key("more"), nil, "Show 14 more days", false).OnClick(func() { a.upcoming = min(a.upcoming+14, 365) })
			})
		})
	})
}

// Stats: focus time today and over the last two weeks, and by tag.

func (a *App) statsPage(c *ui.Context) {
	p := pal(c)
	now := a.now()
	ss := a.data.Sessions
	days := minutesOnDays(ss, now, 14)
	week := 0.0
	for _, m := range days[7:] {
		week += m
	}
	todayCount := 0
	for _, s := range ss {
		if sameDay(s.Start.In(now.Location()), now) {
			todayCount++
		}
	}
	ui.Scroll(c).Fill().Gap(12).Children(func() {
		ui.Row(c).Gap(12).Wrap().Children(func() {
			tile(c, icFocus, p.accent, "TODAY", hoursLabel(days[13]), plural(todayCount, "session", "sessions"))
			tile(c, icStats, p.breakTint, "LAST 7 DAYS", hoursLabel(week), fmt.Sprintf("%s a day on average", hoursLabel(week/7)))
			n := streak(ss, now)
			tile(c, icSparkle, p.warn, "STREAK", plural(n, "day", "days"), "in a row with focus")
		})
		card(c).Children(func() {
			heading(c, "LAST 14 DAYS")
			barChart(c, days, now)
		})
		card(c).Children(func() {
			heading(c, "BY TAG · LAST 7 DAYS")
			tags := minutesByTag(ss, startOfDay(now).AddDate(0, 0, -6))
			if len(tags) == 0 {
				ui.Text(c, "Finish a focus session to see where your time goes.").FontSize(13).TextColor(p.muted)
				return
			}
			top := tags[0].Minutes
			for i, t := range tags {
				if i >= 12 {
					ui.Text(c, plural(len(tags)-12, "more tag", "more tags")).FontSize(12).TextColor(p.muted)
					break
				}
				name := t.Tag
				if name == "" {
					name = "Untagged"
				}
				ui.Row(c.Key(name)).Gap(10).AlignItems(ui.Center).Children(func() {
					ui.Text(c, name).FontSize(13).FontWeight(500).Width(160).Shrink(0).SingleLine().Ellipsis("…").Tooltip(name)
					ui.Box(c).Grow(1).Height(10).Radius(5).Background(p.track).Children(func() {
						bar := ui.Box(c).FillHeight().Radius(5)
						bar.WidthPercent(float32(100*t.Minutes/top) * grow(bar, stagger(i))).
							LinearGradient(ui.LinearGradient{Angle: 90, From: p.accent.Alpha(0.7), To: p.accent, Oklab: true})
					})
					ui.Text(c, hoursLabel(t.Minutes)).FontSize(12.5).FontFeatures("tnum").Width(88).TextAlign(ui.End).Shrink(0).TextColor(p.muted)
				})
			}
		})
	})
}

func tile(c *ui.Context, ic *ui.SVG, tint ui.Color, label, value, sub string) {
	p := pal(c)
	card(c).Grow(1).Basis(180).Gap(4).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Padding(0, 0, 4, 0).Children(func() {
			ui.Box(c).Size(26, 26).Radius(8).Center().Shrink(0).Background(tint.Alpha(0.15)).Children(func() {
				ui.Icon(c, ic).Size(14, 14).TextColor(tint)
			})
			heading(c, label)
		})
		ui.Text(c, value).FontSize(28).FontWeight(650).LetterSpacing(-0.5).FontFeatures("tnum").ContentTransition(ui.ContentTransition{Numeric: true})
		ui.Text(c, sub).FontSize(12).TextColor(p.muted)
	})
}

// barChart draws a bar for each day, today last and in the accent.
func barChart(c *ui.Context, days []float64, now time.Time) {
	p := pal(c)
	top := 60.0
	for _, m := range days {
		top = math.Max(top, m)
	}
	ui.Row(c).Height(180).Gap(6).AlignItems(ui.Stretch).Children(func() {
		for i, m := range days {
			day := startOfDay(now).AddDate(0, 0, i-len(days)+1)
			ui.Column(c.Key(i)).Grow(1).Basis(0).MinWidth(0).Gap(4).Justify(ui.End).AlignItems(ui.Center).
				Tooltip(day.Format("Mon 2 Jan") + ": " + hoursLabel(m)).Children(func() {
				h := float32(140 * m / top)
				bar := ui.Box(c).Width(18).MaxWidthPercent(100).Radius(5, 5, 2, 2)
				bar.Height(max(h*grow(bar, time.Duration(i)*22*time.Millisecond), 2))
				today := i == len(days)-1
				switch {
				case today:
					bar.LinearGradient(ui.LinearGradient{Angle: 180, From: p.accent.Mix(ui.RGB(255, 255, 255), 0.2), To: p.accent, Oklab: true}).
						Shadow(0, 4, 12, -4, p.accent.Alpha(0.5))
				case m == 0:
					bar.Background(p.track)
				case p.dark:
					bar.Background(p.accent.Alpha(0.55))
				default:
					bar.Background(p.accent.Alpha(0.4))
				}
				label, fg := day.Format("Mon")[:1], p.muted
				if today {
					label, fg = "T", p.accentInk
				}
				ui.Text(c, label).FontSize(11).FontWeight(600).TextColor(fg)
			})
		}
	})
}

// Settings.

func (a *App) settingsPage(c *ui.Context) {
	p := pal(c)
	s := &a.data.Settings
	changed := false
	ui.Scroll(c).Fill().Gap(12).Children(func() {
		card(c).Gap(14).Children(func() {
			heading(c, "FOCUS PRESETS")
			for i := range s.Presets {
				pr := &s.Presets[i]
				ui.Column(c.Key(i)).Gap(8).Children(func() {
					ui.TextInput(c.Key("name"), &pr.Name).Label("Preset name").Width(220).OnChange(func() { changed = true })
					ui.Row(c).Gap(10).Wrap().Children(func() {
						for _, f := range []struct {
							label  string
							v      *int
							lo, hi float64
						}{{"Focus (min)", &pr.Work, 1, 180}, {"Break (min)", &pr.Break, 1, 60}, {"Long break (min)", &pr.LongBreak, 1, 90}, {"Long break every", &pr.LongEvery, 1, 12}} {
							v := float64(*f.v)
							ui.Field(c.Key(f.label), f.label, func() {
								if ui.NumberInput(c, &v, f.lo, f.hi, 1).Label(f.label).Changed() {
									*f.v = int(v)
									changed = true
								}
							}).Width(180)
						}
					})
				})
			}
		})
		card(c).Gap(12).Children(func() {
			heading(c, "BEHAVIOUR")
			toggleRow(c, "Start breaks and the next session automatically", &s.AutoStart, &changed)
			toggleRow(c, "Do Not Disturb while focusing (meetings still come through)", &s.FocusDND, &changed)
			toggleRow(c, "Offer to end a session before a meeting", &s.MeetingAware, &changed)
			toggleRow(c, "Play the notification sound", &s.Sound, &changed)
			login := s.OpenAtLogin
			toggleRow(c, "Start with Windows", &s.OpenAtLogin, &changed)
			if login != s.OpenAtLogin {
				a.h.setLogin(s.OpenAtLogin)
			}
			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Default alert for new events").Grow(1)
				lead := optionFor(leadOptions, leadMinutes, s.DefaultLead)
				if ui.Select(c.Key("lead"), &lead, leadOptions).Label("Default alert").Changed() {
					s.DefaultLead = valueFor(leadOptions, leadMinutes, lead, 10)
					changed = true
				}
			})
		})
		card(c).Gap(12).Children(func() {
			heading(c, "SHORTCUT")
			ui.Text(c, "Opens the tray panel with quick-add, from any app.").FontSize(12.5).TextColor(p.muted)
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.TextInput(c.Key("hotkey"), &a.hotkeyDraft).Label("Global shortcut").Width(220)
				pillButton(c.Key("apply"), nil, "Apply", false).OnClick(func() {
					if err := a.h.setHotkey(a.hotkeyDraft); err != nil {
						a.hotkeyErr = err.Error()
						return
					}
					a.hotkeyErr = ""
					s.Hotkey = a.hotkeyDraft
					a.save()
				})
			})
			if a.hotkeyErr != "" {
				ui.Text(c, a.hotkeyErr).FontSize(12.5).TextColor(p.danger)
			}
			ui.Text(c, "In the app: Ctrl+N new event · Ctrl+Shift+N new reminder · Ctrl+1–5 pages · Space starts the timer on Focus").
				FontSize(12).TextColor(p.faint)
		})
		card(c).Gap(10).Children(func() {
			heading(c, "DATA")
			ui.Text(c, a.store.Path()).FontSize(12).TextColor(p.muted).Selectable()
			pillButton(c.Key("folder"), icFolder, "Open folder", false).OnClick(func() {
				_ = mygo.Shell.OpenPath(dataDir())
			})
		})
	})
	if changed {
		if a.timer.Fresh() && a.timer.Phase == PhaseWork {
			a.timer.Reset(a.preset())
		}
		a.save()
	}
}

func toggleRow(c *ui.Context, label string, v *bool, changed *bool) {
	ui.Row(c.Key(label)).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).Grow(1).MinWidth(0)
		if ui.Switch(c, v).Label(label).Changed() {
			*changed = true
		}
	})
}
