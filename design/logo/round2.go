package main

import "fmt"

func init() { logos[2] = round2() }

// gloss is a soft light over the top of a tile, which makes it read as a
// thing to touch rather than a flat swatch.
const gloss = `<linearGradient id="gl" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#fff" stop-opacity=".22"/><stop offset=".55" stop-color="#fff" stop-opacity="0"/></linearGradient>`
const glossRect = `<rect x="80" y="80" width="864" height="864" rx="200" fill="url(#gl)"/>`

const amber = `<radialGradient id="am" cx=".4" cy=".35" r=".7"><stop offset="0" stop-color="#fde68a"/><stop offset="1" stop-color="#f59e0b"/></radialGradient>`

// beacon is the amber dot of now, with a soft halo.
func beacon(x, y, r float64) string {
	return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="#fbbf24" fill-opacity=".22"/><circle cx="%.1f" cy="%.1f" r="%.1f" fill="url(#am)"/>`, x, y, r*1.7, x, y, r)
}

func round2() []logo {
	var out []logo
	add := func(id, name, svg string) { out = append(out, logo{id: id, name: name, svg: svg}) }
	blue := lin("bg", "#4f8bff", "#1e40af", 0, 0, 1, 1) + gloss + amber
	ink := lin("bg", "#1c2130", "#0a0c11", 0, 0, 0, 1) + gloss + amber
	navy := lin("bg", "#22336b", "#0b1124", 0, 0, 0, 1) + gloss + amber
	violet := lin("bg", "#6d6af5", "#5b21b6", 0, 0, 1, 1) + gloss + amber
	ring := lin("ar", "#22d3ee", "#3b82f6", 1, 0, 0, 1)

	// 09 → caret bolder, the dot a beacon.
	add("09b", "Caret", doc(blue, "url(#bg)", glossRect+
		`<path d="M318 500 L512 306 L706 500" fill="none" stroke="#fff" stroke-width="92" stroke-linecap="round" stroke-linejoin="round"/>`+
		beacon(512, 680, 66)))

	// 03 → the ring with the Ctrl caret at its heart and the beacon as its head.
	{
		hx, hy := pt(c, c, 280, 300)
		add("03b", "Ring caret", doc(ink+ring, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="280" fill="none" stroke="#fff" stroke-opacity=".08" stroke-width="64"/>`+
			`<path d="`+arc(c, c, 280, 0, 300)+`" fill="none" stroke="url(#ar)" stroke-width="64" stroke-linecap="round"/>`+
			`<path d="M402 570 L512 460 L622 570" fill="none" stroke="#fff" stroke-width="64" stroke-linecap="round" stroke-linejoin="round"/>`+
			beacon(hx, hy, 34)))
	}

	// 18 → three bold bands, on navy so the orange sings.
	{
		top := func(x float64) float64 { return 300 + (x-352)*56/408 }
		bot := func(x float64) float64 { return 492 - (x-352)*56/408 }
		xs := []float64{352, 488, 624, 760}
		segs := ""
		for i := 0; i < 3; i++ {
			fill := "#fb923c"
			if i == 1 {
				fill = "#ffffff"
			}
			a, b := xs[i], xs[i+1]
			segs += fmt.Sprintf(`<path d="M%g %.1f L%g %.1f L%g %.1f L%g %.1f Z" fill="%s"/>`, a, top(a), b, top(b), b, bot(b), a, bot(a), fill)
		}
		add("18b", "Windsock", doc(navy, "url(#bg)", glossRect+
			`<path d="M330 260 V790" stroke="#fff" stroke-width="44" stroke-linecap="round"/>`+segs))
	}

	// 16 → a white key on the accent, with the indicator light lit.
	add("16b", "Keycap", doc(blue, "url(#bg)", glossRect+
		`<rect x="248" y="268" width="528" height="512" rx="128" fill="#172554" fill-opacity=".45"/>`+
		`<rect x="248" y="236" width="528" height="512" rx="128" fill="#fff"/>`+
		`<path d="M402 560 L512 450 L622 560" fill="none" stroke="#2563eb" stroke-width="64" stroke-linecap="round" stroke-linejoin="round"/>`+
		`<rect x="320" y="310" width="72" height="24" rx="12" fill="#f59e0b"/>`))

	// 02 → three shapes: cab, shaft, beacon.
	add("02b", "Tower", doc(blue, "url(#bg)", glossRect+
		beacon(512, 238, 34)+
		`<path d="M352 330 H672 L626 470 H398 Z" fill="#fff" stroke="#fff" stroke-width="28" stroke-linejoin="round"/>`+
		`<rect x="404" y="372" width="216" height="40" rx="20" fill="#2a5ad8"/>`+
		`<path d="M448 500 H576 L604 792 H420 Z" fill="#fff" fill-opacity=".9" stroke="#fff" stroke-opacity=".9" stroke-width="20" stroke-linejoin="round"/>`))

	// 01 → one ring, a bold sweep, the blip a beacon.
	{
		x, y := pt(c, c, 300, 70)
		bx, by := pt(c, c, 205, 40)
		add("01b", "Radar", doc(navy+`<linearGradient id="sw" x1="1" y1=".3" x2="0" y2=".8"><stop offset="0" stop-color="#60a5fa" stop-opacity=".9"/><stop offset="1" stop-color="#60a5fa" stop-opacity="0"/></linearGradient>`, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".22" stroke-width="32"/>`+
			fmt.Sprintf(`<path d="M512 512 L512 212 A300 300 0 0 1 %.1f %.1f Z" fill="url(#sw)"/>`, x, y)+
			fmt.Sprintf(`<path d="M512 512 L%.1f %.1f" stroke="#bfdbfe" stroke-width="28" stroke-linecap="round"/>`, x, y)+
			beacon(bx, by, 34)+`<circle cx="512" cy="512" r="32" fill="#fff"/>`))
	}

	// 05 → room between the letters, the t's dot a beacon.
	add("05b", "ct", doc(violet, "url(#bg)", glossRect+
		`<path d="`+arc(390, 570, 140, 135, 405)+`" fill="none" stroke="#fff" stroke-width="76" stroke-linecap="round"/>`+
		`<path d="M650 380 V600 Q650 700 750 700" fill="none" stroke="#fff" stroke-width="76" stroke-linecap="round"/>`+
		`<path d="M580 470 H730" stroke="#fff" stroke-width="76" stroke-linecap="round"/>`+
		beacon(650, 290, 30)))

	// New: the tower inside the timer's ring, its beacon the ring's head.
	{
		hx, hy := pt(c, c, 300, 300)
		add("21", "Tower ring", doc(ink+ring, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".08" stroke-width="52"/>`+
			`<path d="`+arc(c, c, 300, 0, 300)+`" fill="none" stroke="url(#ar)" stroke-width="52" stroke-linecap="round"/>`+
			`<path d="M402 392 H622 L590 486 H434 Z" fill="#fff" stroke="#fff" stroke-width="22" stroke-linejoin="round"/>`+
			`<path d="M458 506 H566 L588 676 H436 Z" fill="#fff" fill-opacity=".85" stroke="#fff" stroke-opacity=".85" stroke-width="16" stroke-linejoin="round"/>`+
			beacon(hx, hy, 30)))
	}
	return out
}
