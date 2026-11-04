package main

import "fmt"

func init() { logos[5] = round5() }

// towerC is the tower after the judges: a narrower cab that clears the
// beacon, a thicker roof cap, a tall dark window band that survives 16 px,
// a fuller shaft, and a footing so it stands rather than floats.
func towerC(cx, top, s float64, fill, window string, footing bool) string {
	f := func(v float64) float64 { return v * s }
	out := fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(86), top, f(172), f(44), f(22), fill) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(136), top+f(64), cx+f(136), cx+f(98), top+f(204), cx-f(98), fill, fill, f(24)) +
		fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(100), top+f(100), f(200), f(66), f(33), window) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(54), top+f(228), cx+f(54), cx+f(82), top+f(430), cx-f(82), fill, fill, f(20))
	if footing {
		out += fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(124), top+f(436), f(248), f(36), f(18), fill)
	}
	return out
}

// glowBeacon is the beacon with no rim: amber, with a soft light of its
// own that reads on blue as well as on ink.
func glowBeacon(x, y, r float64) string {
	return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="url(#gw)"/><circle cx="%.1f" cy="%.1f" r="%.1f" fill="url(#am)"/>`, x, y, r*2, x, y, r)
}

const glowDef = `<radialGradient id="gw"><stop offset=".45" stop-color="#fde68a" stop-opacity=".55"/><stop offset="1" stop-color="#fde68a" stop-opacity="0"/></radialGradient>`

// towerRing is the mark: the ring 290° round, its gap at the top left,
// the beacon at its head and the tower within.
func towerRing(defs, track, arcPaint, window string, stroke, scale float64) string {
	hx, hy := pt(c, c, 300, 290)
	h := 450 * scale // the tower's height with its footing
	return doc(defs+glowDef, "url(#bg)", glossRect+
		fmt.Sprintf(`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity="%s" stroke-width="%g"/>`, track, stroke)+
		`<path d="`+arc(c, c, 300, 0, 290)+fmt.Sprintf(`" fill="none" stroke="%s" stroke-width="%g" stroke-linecap="round"/>`, arcPaint, stroke)+
		towerC(512, 512-h/2+8, scale, "#fff", window, true)+glowBeacon(hx, hy, 40))
}

func round5() []logo {
	var out []logo
	add := func(id, name, svg string) { out = append(out, logo{id: id, name: name, svg: svg}) }
	blue := lin("bg", "#4f8bff", "#1e40af", 0, 0, 1, 1) + gloss + amber
	ink := lin("bg", "#1c2130", "#0a0c11", 0, 0, 0, 1) + gloss + amber + lin("ar", "#22d3ee", "#3b82f6", 1, 0, 0, 1)

	add("22a", "Tower ring · blue", towerRing(blue, ".14", "#fff", "#1e3a8a", 68, 0.94))
	add("22b", "Tower ring · ink", towerRing(ink, ".08", "url(#ar)", "#141824", 68, 0.94))
	// The tray cut: heavier ring, larger tower, no gloss to muddy it.
	{
		hx, hy := pt(c, c, 290, 290)
		add("22c", "Tray cut (16 px)", doc(blue+glowDef, "url(#bg)",
			`<circle cx="512" cy="512" r="290" fill="none" stroke="#fff" stroke-opacity=".14" stroke-width="100"/>`+
				`<path d="`+arc(c, c, 290, 0, 290)+`" fill="none" stroke="#fff" stroke-width="100" stroke-linecap="round"/>`+
				towerC(512, 512-450*0.86/2+6, 0.86, "#fff", "#1e3a8a", false)+
				fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="58" fill="url(#am)"/>`, hx, hy)))
	}
	// The rematch Gemini asked for: the Ctrl caret in the same ring.
	{
		hx, hy := pt(c, c, 300, 290)
		add("03c", "Ring caret", doc(blue+glowDef, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".14" stroke-width="68"/>`+
			`<path d="`+arc(c, c, 300, 0, 290)+`" fill="none" stroke="#fff" stroke-width="68" stroke-linecap="round"/>`+
			`<path d="M392 580 L512 460 L632 580" fill="none" stroke="#fff" stroke-width="80" stroke-linecap="round" stroke-linejoin="round"/>`+
			glowBeacon(hx, hy, 40)))
	}
	return out
}
