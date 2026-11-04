package main

import "fmt"

func init() { logos[6] = round6() }

// towerD is the tower for the app tile: as towerC, with the window's
// corners squared off a little so the cab reads as glazing, not a slot.
func towerD(cx, top, s float64, fill, window string) string {
	f := func(v float64) float64 { return v * s }
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(86), top, f(172), f(44), f(22), fill) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(136), top+f(64), cx+f(136), cx+f(98), top+f(204), cx-f(98), fill, fill, f(24)) +
		fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(100), top+f(100), f(200), f(66), f(16), window) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(54), top+f(228), cx+f(54), cx+f(82), top+f(430), cx-f(82), fill, fill, f(20)) +
		fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(124), top+f(436), f(248), f(36), f(18), fill)
}

// towerSmall is the tower for 32 and 16 px: no roof cap, no footing, the
// cab one bold shape with one tall window, the shaft broad.
func towerSmall(cx, top, s float64, fill, window string) string {
	f := func(v float64) float64 { return v * s }
	return fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
		cx-f(150), top, cx+f(150), cx+f(104), top+f(170), cx-f(104), fill, fill, f(36)) +
		fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(100), top+f(44), f(200), f(84), f(20), window) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(62), top+f(196), cx+f(62), cx+f(92), top+f(400), cx-f(92), fill, fill, f(28))
}

func round6() []logo {
	var out []logo
	add := func(id, name, svg string) { out = append(out, logo{id: id, name: name, svg: svg}) }
	blue := lin("bg", "#4f8bff", "#1e40af", 0, 0, 1, 1) + gloss + amber + glowDef
	flat := lin("bg", "#4f8bff", "#1e40af", 0, 0, 1, 1) + amber

	// App icon: 22a without the track; glow kept, it is large here.
	{
		hx, hy := pt(c, c, 300, 290)
		add("23a", "App icon", doc(blue, "url(#bg)", glossRect+
			`<path d="`+arc(c, c, 300, 0, 290)+`" fill="none" stroke="#fff" stroke-width="68" stroke-linecap="round"/>`+
			towerD(512, 512-450*0.94/2+8, 0.94, "#fff", "#1e3a8a")+glowBeacon(hx, hy, 40)))
	}
	// Taskbar, 32 px: the small tower, a heavier ring, a flat beacon.
	{
		hx, hy := pt(c, c, 296, 290)
		add("23b", "Taskbar 32 px", doc(flat, "url(#bg)",
			`<path d="`+arc(c, c, 296, 0, 290)+`" fill="none" stroke="#fff" stroke-width="88" stroke-linecap="round"/>`+
				towerSmall(512, 512-400*0.86/2, 0.86, "#fff", "#1e3a8a")+
				fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="56" fill="url(#am)"/>`, hx, hy)))
	}
	// Tray, 16 px: heavier still.
	{
		hx, hy := pt(c, c, 290, 290)
		add("23c", "Tray 16 px", doc(flat, "url(#bg)",
			`<path d="`+arc(c, c, 290, 0, 290)+`" fill="none" stroke="#fff" stroke-width="112" stroke-linecap="round"/>`+
				towerSmall(512, 512-400*0.8/2, 0.8, "#fff", "#1e3a8a")+
				fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="68" fill="url(#am)"/>`, hx, hy)))
	}
	// Haiku's alternative for the tray: the ring and beacon alone.
	{
		hx, hy := pt(c, c, 260, 290)
		add("23d", "Tray · ring only", doc(flat, "url(#bg)",
			`<path d="`+arc(c, c, 260, 0, 290)+`" fill="none" stroke="#fff" stroke-width="120" stroke-linecap="round"/>`+
				fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="78" fill="url(#am)"/>`, hx, hy)))
	}
	return out
}
