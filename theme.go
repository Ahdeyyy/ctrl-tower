package main

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// palette is the app's look, worked out from the window's theme so it
// follows Windows' light and dark modes and its accent color. Panes are
// solid over a near-white or near-black gradient, so text holds its
// contrast; color comes from the accent and the few hues that mean
// something (reminders, breaks, now), not from tinted glass.
type palette struct {
	dark       bool
	bgTop      ui.Color // the window's gradient, from the top…
	bgBottom   ui.Color // …to the bottom
	surface    ui.Color // panes
	raised     ui.Color // what floats over panes: dialogs, popovers
	hover      ui.Color
	pressed    ui.Color
	rim        ui.Color // a pane's edge
	divider    ui.Color // lines within a pane
	shine      ui.Color // the light along a pane's top edge
	shadow     ui.Color
	lift       ui.Color // the shadow of what floats: dialogs, popovers
	text       ui.Color
	muted      ui.Color
	faint      ui.Color
	accent     ui.Color
	accentSoft ui.Color
	accentInk  ui.Color // the accent as text, legible on a pane
	accentText ui.Color // text on the accent
	danger     ui.Color
	warn       ui.Color
	ok         ui.Color
	breakTint  ui.Color // the ring's color on a break
	field      ui.Color // text inputs
	track      ui.Color // the empty part of rings, bars and dots
}

func pal(c *ui.Context) palette {
	t := c.Theme()
	accent := t.Accent
	if t.Dark {
		return palette{
			dark:       true,
			bgTop:      ui.Hex("#12141b").Mix(accent, 0.06),
			bgBottom:   ui.Hex("#08090c"),
			surface:    ui.Hex("#15171e"),
			raised:     ui.Hex("#1c1f28"),
			hover:      accent.Alpha(0.10),
			pressed:    accent.Alpha(0.16),
			rim:        ui.RGBA(255, 255, 255, 0.08),
			divider:    ui.RGBA(255, 255, 255, 0.07),
			shine:      ui.RGBA(255, 255, 255, 0.05),
			shadow:     ui.RGBA(0, 0, 0, 0.45),
			lift:       ui.RGBA(0, 0, 0, 0.6),
			text:       ui.Hex("#f3f4f7"),
			muted:      ui.Hex("#a9afbd"),
			faint:      ui.Hex("#7e8596"),
			accent:     accent,
			accentSoft: accent.Alpha(0.18),
			accentInk:  accent.Mix(ui.RGB(255, 255, 255), 0.28),
			accentText: t.AccentText,
			danger:     ui.Hex("#f97066"),
			warn:       ui.Hex("#fdb022"),
			ok:         ui.Hex("#47cd89"),
			breakTint:  ui.Hex("#34d399"),
			field:      ui.Hex("#0e1015"),
			track:      ui.RGBA(255, 255, 255, 0.08),
		}
	}
	return palette{
		bgTop:      ui.Hex("#f7f8fb").Mix(accent, 0.05),
		bgBottom:   ui.Hex("#eceef2"),
		surface:    ui.Hex("#ffffff"),
		raised:     ui.Hex("#ffffff"),
		hover:      accent.Alpha(0.07),
		pressed:    accent.Alpha(0.13),
		rim:        ui.RGBA(16, 24, 40, 0.09),
		divider:    ui.RGBA(16, 24, 40, 0.07),
		shine:      ui.RGBA(255, 255, 255, 0),
		shadow:     ui.RGBA(16, 24, 40, 0.07),
		lift:       ui.RGBA(16, 24, 40, 0.18),
		text:       ui.Hex("#0e1117"),
		muted:      ui.Hex("#4a5163"),
		faint:      ui.Hex("#6a7182"),
		accent:     accent,
		accentSoft: accent.Alpha(0.11),
		accentInk:  accent.Mix(ui.RGB(0, 0, 0), 0.18),
		accentText: t.AccentText,
		danger:     ui.Hex("#d92d20"),
		warn:       ui.Hex("#b54708"),
		ok:         ui.Hex("#067647"),
		breakTint:  ui.Hex("#059669"),
		field:      ui.Hex("#f5f6f9"),
		track:      ui.RGBA(16, 24, 40, 0.07),
	}
}

// Motion. The strong ease-out for things entering and leaving, springs
// for things that follow the pointer or a choice; nothing a click waits
// on runs longer than 300 ms, and MyGo drops movement while Windows asks
// for less animation.
var (
	easeOut   = ui.CubicBezier(0.23, 1, 0.32, 1)
	enterAnim = ui.Animation{Duration: 220 * time.Millisecond, Ease: easeOut}
	colorAnim = ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseInOut}
	pressAnim = ui.Animation{Duration: 120 * time.Millisecond, Ease: easeOut}
	// slide moves indicators to a new choice: quick, with a hint of
	// overshoot so it reads as one thing travelling.
	slide     = ui.Animation{Spring: ui.SpringOptions{Duration: 320 * time.Millisecond, Bounce: 0.12}}
	rowMotion = ui.ElementTransition{
		Duration: 200 * time.Millisecond,
		Ease:     easeOut,
		Enter:    &ui.Motion{Collapse: true},
		Exit:     &ui.Motion{Collapse: true},
	}
)

// appTheme adjusts the window's theme to the palette, so MyGo's own
// controls (inputs, selects, switches, toasts) match the panes.
func appTheme(c *ui.Context) {
	t := *c.Theme()
	p := pal(c)
	t.Background = p.bgBottom
	t.Surface = p.field
	if t.Dark {
		t.SurfaceHover = ui.Hex("#1a1d25")
		t.SurfacePressed = ui.Hex("#232733")
		t.Border = ui.RGBA(255, 255, 255, 0.12)
		t.Inverse, t.InverseText = ui.Hex("#f3f4f7"), ui.Hex("#0e1117")
		t.Scrollbar = ui.RGBA(255, 255, 255, 0.28)
	} else {
		t.SurfaceHover = ui.Hex("#eef0f4")
		t.SurfacePressed = ui.Hex("#e4e7ec")
		t.Border = ui.RGBA(16, 24, 40, 0.16)
		t.Inverse, t.InverseText = ui.Hex("#161922"), ui.Hex("#f3f4f7")
		t.Scrollbar = ui.RGBA(16, 24, 40, 0.30)
	}
	t.Radius = 8
	t.Text, t.TextMuted = p.text, p.faint
	t.Danger, t.Warning, t.Success = p.danger, p.warn, p.ok
	t.Focus = p.accent.Alpha(0.55)
	t.Selection = p.accent.Alpha(0.28)
	c.SetTheme(&t)
}

// backdrop paints the window: a gradient from just off white (or black),
// touched with the accent at the top, to a neutral bottom.
func backdrop(c *ui.Context) {
	p := pal(c)
	c.Root().Background(p.bgBottom).LinearGradient(ui.LinearGradient{
		Angle: 180, From: p.bgTop, To: p.bgBottom, Oklab: true,
	})
}

// pane styles e as a solid pane: the surface, a hairline edge and a soft
// shadow that lifts it off the background.
func pane(c *ui.Context, e ui.Element) ui.Element {
	p := pal(c)
	return e.Background(p.surface).Border(1, p.rim).Radius(16).
		InsetShadow(0, 1, 0, 0, p.shine).
		Shadow(0, 6, 20, -8, p.shadow)
}

// card is a pane of content.
func card(c *ui.Context) ui.Element {
	return pane(c, ui.Column(c)).Padding(16).Gap(10)
}

// pressScale gives a control a slight give while pressed, and returns it.
func pressScale(b ui.Element, pressed float32) ui.Element {
	scale := float32(1)
	if b.Pressed() {
		scale = pressed
	}
	return b.AnimatePose(ui.Pose{Scale: scale}, pressAnim)
}

// iconButton is a square button showing an icon, named for screen
// readers and tooltips.
func iconButton(c *ui.Context, ic *ui.SVG, label string) ui.Element {
	p := pal(c)
	b := ui.ButtonBase(c).Size(32, 32).Shrink(0).Radius(8).Center().
		TextColor(p.muted).Label(label).Tooltip(label)
	bg, fg := p.hover.Alpha(0), p.muted
	if b.Pressed() {
		bg, fg = p.pressed, p.accentInk
	} else if b.Hovered() {
		bg, fg = p.hover, p.accentInk
	}
	b.Background(b.AnimateColor("bg", bg, colorAnim)).TextColor(b.AnimateColor("fg", fg, colorAnim))
	pressScale(b, 0.9)
	b.Children(func() { ui.Icon(c, ic).Size(16, 16) })
	return b
}

// pillButton is a rounded button with an icon and a label; primary ones
// take the accent and lift on hover.
func pillButton(c *ui.Context, ic *ui.SVG, label string, primary bool) ui.Element {
	p := pal(c)
	b := ui.ButtonBase(c).Row().Gap(6).Padding(7, 14).Radius(999).Shrink(0).AlignItems(ui.Center).Label(label)
	bg, fg := p.surface, p.text
	if primary {
		bg, fg = p.accent, p.accentText
	}
	if b.Pressed() {
		bg = bg.Mix(ui.RGB(0, 0, 0), 0.12)
	} else if b.Hovered() {
		if primary {
			bg = bg.Mix(ui.RGB(255, 255, 255), 0.12)
		} else {
			bg = p.hover.Over(p.surface)
		}
	}
	b.Background(b.AnimateColor("bg", bg, colorAnim)).TextColor(fg)
	lift := float32(0)
	if b.Hovered() && !b.Pressed() {
		lift = 1
	}
	lift = b.AnimateTo("lift", lift, colorAnim)
	if primary {
		b.Shadow(0, 2+3*lift, 8+8*lift, -3, p.accent.Alpha(0.28+0.17*lift))
	} else {
		b.Border(1, p.rim.Mix(p.accent, 0.5*lift)).Shadow(0, 1+lift, 2+4*lift, 0, p.shadow)
	}
	pressScale(b, 0.96)
	b.Children(func() {
		if ic != nil {
			ui.Icon(c, ic).Size(15, 15)
		}
		ui.Text(c, label).SingleLine().FontWeight(600).FontSize(13)
	})
	return b
}

// segmented is a segmented control whose chosen segment is a pill that
// slides to the new choice. It reports Changed as ui.Segmented does.
func segmented(c *ui.Context, selected *int, labels ...string) ui.Element {
	p := pal(c)
	parts := ui.SegmentedBase(c, selected, len(labels))
	track := parts.Track.AlignItems(ui.Stretch).Padding(3).Gap(2).Radius(10).
		Background(p.field).Border(1, p.rim)
	var x, w float32
	track.Children(func() {
		for i, label := range labels {
			s := parts.Segment(i).Padding(5, 14).Radius(7).Label(label)
			fg := p.muted
			switch {
			case i == *selected:
				fg = p.text
				b, tb := s.LayoutBounds(), track.LayoutBounds()
				x, w = b.X-tb.X, b.W
			case s.Hovered():
				fg = p.text
			}
			s.TextColor(s.AnimateColor("fg", fg, colorAnim))
			s.Children(func() { ui.Text(c, label).SingleLine().FontSize(13).FontWeight(600) })
		}
	})
	if w > 0 {
		x, w = track.AnimateTo("x", x, slide), track.AnimateTo("w", w, slide)
		pill, sh := p.surface, p.shadow
		if p.dark {
			pill = ui.Hex("#272b36")
		}
		track.Draw(func(pt *ui.Painter, r ui.Rect) {
			seg := ui.Rect{X: r.X + x, Y: r.Y + 3, W: w, H: r.H - 6}
			pt.Shadow(seg, 7, 0, 1, 3, 0, sh)
			pt.Fill(seg, pill, 7)
		})
	}
	return track
}

// chip is a small label in a pill, as a tag or a repeat rule.
func chip(c *ui.Context, ic *ui.SVG, text string, tint ui.Color) ui.Element {
	return ui.Row(c).Gap(4).Padding(2, 8).Radius(999).Background(tint.Alpha(0.13)).
		TextColor(tint).Shrink(0).AlignItems(ui.Center).Children(func() {
		if ic != nil {
			ui.Icon(c, ic).Size(12, 12)
		}
		ui.Text(c, text).FontSize(11.5).FontWeight(600).SingleLine()
	})
}

func heading(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).FontSize(11.5).FontWeight(700).TextColor(pal(c).muted).LetterSpacing(0.7)
}

// field is a text input.
func field(c *ui.Context, value *string, placeholder string) ui.Element {
	return ui.TextInput(c, value).Placeholder(placeholder).Label(placeholder)
}

// Entrances.

// popIn plays e in as it appears: from 96% and transparent to its place,
// in 220 ms on the strong ease-out. Dialogs and the overlay open a few
// times a day, which leaves room for it; it plays once per opening.
func popIn(e ui.Element) {
	v := e.Keyframes("enter", []ui.Keyframe{{Value: 0}, {At: enterAnim.Duration, Value: 1, Ease: easeOut}})
	if v >= 1 {
		return
	}
	s := 0.96 + 0.04*v
	e.Opacity(v).Scale(s, s)
}

// riseIn fades e in as it appears while it rises dy DIPs into place,
// after delay: pages as they are chosen, rows one after another.
func riseIn(e ui.Element, dy float32, d, delay time.Duration) {
	v := e.Keyframes("rise", []ui.Keyframe{{Value: 0}, {At: d, Value: 1, Ease: easeOut}}, ui.KeyframeOptions{Delay: delay})
	if v >= 1 {
		return
	}
	e.Opacity(v).Translate(0, dy*(1-v))
}

// stagger is the delay of the ith item of a list entering: one after
// another, the whole list in at most a quarter of a second.
func stagger(i int) time.Duration {
	return time.Duration(min(i, 8)) * 28 * time.Millisecond
}

// grow is how far a bar has grown since it appeared, from 0 to 1, after
// delay.
func grow(e ui.Element, delay time.Duration) float32 {
	return e.Keyframes("grow", []ui.Keyframe{{Value: 0}, {At: 520 * time.Millisecond, Value: 1, Ease: easeOut}}, ui.KeyframeOptions{Delay: delay})
}
