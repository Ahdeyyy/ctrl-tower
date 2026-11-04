package main

import "strings"

func init() { logos[7] = round7() }

// bleed turns a tile drawn with the app icon's margin into one that fills
// its canvas, as small icons should: at 16 px the margin costs a pixel and
// a half a side. The mark scales up with it, about the center.
func bleed(svg string) string {
	svg = strings.Replace(svg, `<rect x="80" y="80" width="864" height="864" rx="200"`, `<rect x="0" y="0" width="1024" height="1024" rx="232"`, 1)
	i := strings.Index(svg, `</defs>`) + len(`</defs>`)
	j := strings.Index(svg[i:], `/>`) + i + 2 // past the tile
	return svg[:j] + `<g transform="translate(512 512) scale(1.12) translate(-512 -512)">` + strings.TrimSuffix(svg[j:], `</svg>`) + `</g></svg>`
}

func round7() []logo {
	r6 := map[string]logo{}
	for _, l := range round6() {
		r6[l.id] = l
	}
	var out []logo
	out = append(out, r6["23a"])
	for _, id := range []string{"23b", "23c", "23d"} {
		l := r6[id]
		out = append(out, logo{id: strings.Replace(id, "23", "24", 1), name: l.name + " · full bleed", svg: bleed(l.svg)})
	}
	return out
}
