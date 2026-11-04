package main

import "fmt"

func init() { logos[3] = round3() }

// rimmed is the beacon for light or saturated tiles: a crisp white rim in
// place of a halo, which goes muddy over blue.
func rimmed(x, y, r float64) string {
	return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="#fff"/><circle cx="%.1f" cy="%.1f" r="%.1f" fill="url(#am)"/>`, x, y, r+12, x, y, r)
}

// tower draws the tower at scale s about (cx, top): a roof cap, the cab
// with its band of windows, a slim shaft. fill is its color, window the
// band's.
func tower(cx, top, s float64, fill, window string) string {
	f := func(v float64) float64 { return v * s }
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(92), top, f(184), f(30), f(15), fill) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(150), top+f(54), cx+f(150), cx+f(108), top+f(176), cx-f(108), fill, fill, f(24)) +
		fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(106), top+f(92), f(212), f(38), f(19), window) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" fill-opacity=".9" stroke="%s" stroke-opacity=".9" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(48), top+f(204), cx+f(48), cx+f(72), top+f(470), cx-f(72), fill, fill, f(18))
}

func round3() []logo {
	var out []logo
	add := func(id, name, svg string) { out = append(out, logo{id: id, name: name, svg: svg}) }
	blue := lin("bg", "#4f8bff", "#1e40af", 0, 0, 1, 1) + gloss + amber
	ink := lin("bg", "#1c2130", "#0a0c11", 0, 0, 0, 1) + gloss + amber
	navy := lin("bg", "#22336b", "#0b1124", 0, 0, 0, 1) + gloss + amber
	ring := lin("ar", "#22d3ee", "#3b82f6", 1, 0, 0, 1)

	// Tower: a roof cap and a slimmer, taller shaft, so it reads as a tower
	// and not a cup; the beacon rimmed.
	add("02c", "Tower", doc(blue, "url(#bg)", glossRect+
		`<path d="M512 214 V262" stroke="#fff" stroke-width="20" stroke-linecap="round"/>`+rimmed(512, 200, 30)+
		tower(512, 272, 1.15, "#fff", "#2a5ad8")))

	// Tower ring on ink: the tower larger, filling the ring.
	{
		hx, hy := pt(c, c, 300, 300)
		add("21b", "Tower ring · ink", doc(ink+ring, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".09" stroke-width="60"/>`+
			`<path d="`+arc(c, c, 300, 0, 300)+`" fill="none" stroke="url(#ar)" stroke-width="60" stroke-linecap="round"/>`+
			tower(512, 330, 0.78, "#fff", "#141824")+beacon(hx, hy, 32)))
	}
	// Tower ring on the accent, in white.
	{
		hx, hy := pt(c, c, 300, 300)
		add("21c", "Tower ring · blue", doc(blue, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".22" stroke-width="60"/>`+
			`<path d="`+arc(c, c, 300, 0, 300)+`" fill="none" stroke="#fff" stroke-width="60" stroke-linecap="round"/>`+
			tower(512, 330, 0.78, "#fff", "#2a5ad8")+rimmed(hx, hy, 30)))
	}

	// Windsock: banded curves for a body, a beacon on the mast for the
	// family, and the sock dipping as it does in a breeze.
	{
		// Bands between the mouth (x 360) and the tail (x 770), each edge
		// a gentle curve.
		edge := func(x float64) (float64, float64) {
			t := (x - 360) / 410
			return 318 + 70*t + 20*t*t, 498 - 30*t + 20*t*t
		}
		xs := []float64{360, 497, 633, 770}
		segs := ""
		for i := 0; i < 3; i++ {
			fill := "#fb923c"
			if i == 1 {
				fill = "#fff"
			}
			a, b := xs[i], xs[i+1]
			at, ab := edge(a)
			bt, bb := edge(b)
			mt, _ := edge((a + b) / 2)
			_, mb := edge((a + b) / 2)
			segs += fmt.Sprintf(`<path d="M%g %.1f Q%.1f %.1f %g %.1f L%g %.1f Q%.1f %.1f %g %.1f Z" fill="%s"/>`,
				a, at, (a+b)/2, mt-6, b, bt, b, bb, (a+b)/2, mb+6, a, ab, fill)
		}
		add("18c", "Windsock", doc(navy, "url(#bg)", glossRect+
			`<path d="M336 300 V800" stroke="#fff" stroke-width="44" stroke-linecap="round"/>`+
			segs+`<ellipse cx="360" cy="408" rx="26" ry="90" fill="#ea580c"/>`+beacon(336, 236, 34)))
	}

	// Caret: under it the beacon in its own small timer ring.
	add("09c", "Caret", doc(blue, "url(#bg)", glossRect+
		`<path d="M318 476 L512 282 L706 476" fill="none" stroke="#fff" stroke-width="92" stroke-linecap="round" stroke-linejoin="round"/>`+
		`<path d="`+arc(512, 660, 112, 0, 270)+`" fill="none" stroke="#fff" stroke-width="36" stroke-linecap="round"/>`+
		`<circle cx="512" cy="660" r="54" fill="url(#am)"/>`))
	return out
}
