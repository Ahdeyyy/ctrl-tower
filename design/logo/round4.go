package main

import "fmt"

func init() { logos[4] = round4() }

// towerB is the tower tightened for small sizes: the roof cap sits close
// on the cab, the window band is bolder, the shaft fuller.
func towerB(cx, top, s float64, fill, window string) string {
	f := func(v float64) float64 { return v * s }
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(96), top, f(192), f(34), f(17), fill) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(150), top+f(46), cx+f(150), cx+f(106), top+f(176), cx-f(106), fill, fill, f(24)) +
		fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"/>`, cx-f(108), top+f(84), f(216), f(50), f(25), window) +
		fmt.Sprintf(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f Z" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round"/>`,
			cx-f(52), top+f(200), cx+f(52), cx+f(78), top+f(450), cx-f(78), fill, fill, f(20))
}

func round4() []logo {
	var out []logo
	add := func(id, name, svg string) { out = append(out, logo{id: id, name: name, svg: svg}) }
	blue := lin("bg", "#4f8bff", "#1e40af", 0, 0, 1, 1) + gloss + amber

	// The tower ring: tower larger in the ring, the ring a touch slimmer
	// to give it room.
	{
		hx, hy := pt(c, c, 300, 300)
		add("21d", "Tower ring", doc(blue, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".22" stroke-width="56"/>`+
			`<path d="`+arc(c, c, 300, 0, 300)+`" fill="none" stroke="#fff" stroke-width="56" stroke-linecap="round"/>`+
			towerB(512, 318, 0.88, "#fff", "#2b5bd9")+rimmed(hx, hy, 32)))
	}
	// The same, the shaft standing on the ring: the tower grounded, the
	// ring its horizon.
	{
		hx, hy := pt(c, c, 300, 300)
		add("21e", "Tower ring · grounded", doc(blue, "url(#bg)", glossRect+
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-opacity=".22" stroke-width="56"/>`+
			`<path d="`+arc(c, c, 300, 0, 300)+`" fill="none" stroke="#fff" stroke-width="56" stroke-linecap="round"/>`+
			`<clipPath id="in"><circle cx="512" cy="512" r="272"/></clipPath><g clip-path="url(#in)">`+
			towerB(512, 300, 1.0, "#fff", "#2b5bd9")+`</g>`+rimmed(hx, hy, 32)))
	}
	// The tower alone, bolder.
	add("02d", "Tower", doc(blue, "url(#bg)", glossRect+
		`<path d="M512 220 V270" stroke="#fff" stroke-width="22" stroke-linecap="round"/>`+rimmed(512, 204, 32)+
		towerB(512, 278, 1.2, "#fff", "#2b5bd9")))
	return out
}
