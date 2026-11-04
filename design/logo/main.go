// Command logo generates the ctrl-tower logo candidates: an SVG and a PNG
// of each, and a labelled contact sheet, rendered with MyGo itself so they
// look as the app will draw them.
//
//	go run ./design/logo -round 1
//	go run ./design/logo -final
package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"
)

// logo is one candidate: a name and its SVG on a 1024 canvas.
type logo struct {
	id, name string
	svg      string
}

const c = 512.0 // the canvas's center

// pt is the point r from (cx, cy) at deg degrees clockwise from twelve
// o'clock.
func pt(cx, cy, r, deg float64) (float64, float64) {
	a := deg * math.Pi / 180
	return cx + r*math.Sin(a), cy - r*math.Cos(a)
}

// arc is an SVG path of a circle's arc from deg a to b, clockwise.
func arc(cx, cy, r, a, b float64) string {
	x1, y1 := pt(cx, cy, r, a)
	x2, y2 := pt(cx, cy, r, b)
	large := 0
	if b-a > 180 {
		large = 1
	}
	return fmt.Sprintf("M%.1f %.1f A%.1f %.1f 0 %d 1 %.1f %.1f", x1, y1, r, r, large, x2, y2)
}

// doc wraps a mark in the app tile: a rounded square of the given fill
// with its own gradient defs.
func doc(defs, tile, mark string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024"><defs>` + defs + `</defs>` +
		`<rect x="80" y="80" width="864" height="864" rx="200" fill="` + tile + `"/>` + mark + `</svg>`
}

func lin(id, from, to string, x1, y1, x2, y2 float64) string {
	return fmt.Sprintf(`<linearGradient id="%s" x1="%g" y1="%g" x2="%g" y2="%g"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, id, x1, y1, x2, y2, from, to)
}

var logos = map[int][]logo{1: round1()}

func round1() []logo {
	var out []logo
	add := func(name, svg string) {
		out = append(out, logo{id: fmt.Sprintf("%02d", len(out)+1), name: name, svg: svg})
	}
	navy := lin("bg", "#1d2b5c", "#0b1124", 0, 0, 0, 1)
	blue := lin("bg", "#4f8bff", "#1d4ed8", 0, 0, 1, 1)
	ink := lin("bg", "#1a1f2b", "#0a0c11", 0, 0, 0, 1)
	paper := lin("bg", "#ffffff", "#e9edf5", 0, 0, 0, 1)
	violet := lin("bg", "#6366f1", "#7c3aed", 0, 0, 1, 1)

	// 1 Radar: a sweep and a blip, the tower's scope.
	{
		x, y := pt(c, c, 300, 60)
		bx, by := pt(c, c, 210, 32)
		add("Radar", doc(navy+`<linearGradient id="sw" x1="1" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#60a5fa"/><stop offset="1" stop-color="#60a5fa" stop-opacity="0"/></linearGradient>`, "url(#bg)",
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#ffffff" stroke-opacity=".16" stroke-width="16"/>`+
				`<circle cx="512" cy="512" r="170" fill="none" stroke="#ffffff" stroke-opacity=".16" stroke-width="16"/>`+
				fmt.Sprintf(`<path d="M512 512 L512 212 A300 300 0 0 1 %.1f %.1f Z" fill="url(#sw)" opacity=".7"/>`, x, y)+
				fmt.Sprintf(`<path d="M512 512 L%.1f %.1f" stroke="#93c5fd" stroke-width="16" stroke-linecap="round"/>`, x, y)+
				fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="28" fill="#fdb022"/>`, bx, by)+
				`<circle cx="512" cy="512" r="24" fill="#fff"/>`))
	}
	// 2 Tower: a control tower's cab on its shaft.
	add("Tower", doc(blue, "url(#bg)",
		`<path d="M512 196 V280" stroke="#fff" stroke-width="20" stroke-linecap="round"/><circle cx="512" cy="196" r="20" fill="#fdb022"/>`+
			`<rect x="404" y="280" width="216" height="40" rx="12" fill="#fff"/>`+
			`<path d="M344 336 H680 L630 470 H394 Z" fill="#fff"/>`+
			`<rect x="384" y="368" width="256" height="44" rx="10" fill="#1d4ed8"/>`+
			`<path d="M430 470 H594 L626 760 H398 Z" fill="#fff" fill-opacity=".85"/>`+
			`<rect x="300" y="760" width="424" height="36" rx="18" fill="#fff"/>`))
	// 3 Arc: the focus timer's ring, three quarters gone.
	{
		hx, hy := pt(c, c, 270, 270)
		add("Arc", doc(ink+lin("ar", "#22d3ee", "#3b82f6", 1, 0, 0, 1), "url(#bg)",
			`<circle cx="512" cy="512" r="270" fill="none" stroke="#fff" stroke-opacity=".08" stroke-width="72"/>`+
				`<path d="`+arc(c, c, 270, 0, 270)+`" fill="none" stroke="url(#ar)" stroke-width="72" stroke-linecap="round"/>`+
				fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="18" fill="#fff"/>`, hx, hy)))
	}
	// 4 Beacon: a light sending out its signal.
	add("Beacon", doc(paper, "url(#bg)",
		`<circle cx="512" cy="620" r="64" fill="#2563eb"/>`+
			`<path d="`+arc(512, 620, 170, -45, 45)+`" fill="none" stroke="#2563eb" stroke-width="48" stroke-linecap="round"/>`+
			`<path d="`+arc(512, 620, 280, -45, 45)+`" fill="none" stroke="#2563eb" stroke-opacity=".5" stroke-width="48" stroke-linecap="round"/>`+
			`<path d="`+arc(512, 620, 390, -45, 45)+`" fill="none" stroke="#2563eb" stroke-opacity=".22" stroke-width="48" stroke-linecap="round"/>`))
	// 5 Monogram: a lowercase c and t.
	add("ct monogram", doc(violet, "url(#bg)",
		`<path d="`+arc(420, 560, 150, 135, 405)+`" fill="none" stroke="#fff" stroke-width="64" stroke-linecap="round"/>`+
			`<path d="M650 360 V600 Q650 690 740 690" fill="none" stroke="#fff" stroke-width="64" stroke-linecap="round"/>`+
			`<path d="M580 470 H740" stroke="#fff" stroke-width="64" stroke-linecap="round"/>`))
	// 6 Calendar: a page with today marked.
	add("Calendar", doc(blue, "url(#bg)",
		`<rect x="272" y="300" width="480" height="452" rx="72" fill="none" stroke="#fff" stroke-width="44"/>`+
			`<path d="M272 440 H752" stroke="#fff" stroke-width="44"/>`+
			`<path d="M400 240 V340 M624 240 V340" stroke="#fff" stroke-width="44" stroke-linecap="round"/>`+
			`<circle cx="610" cy="600" r="48" fill="#fdb022"/>`))
	// 7 Runway: lines converging on the horizon, a sun rising over it.
	add("Runway", doc(lin("bg", "#0b1020", "#1e3a8a", 0, 0, 0, 1), "url(#bg)",
		`<circle cx="512" cy="330" r="90" fill="#fdb022" fill-opacity=".18"/><circle cx="512" cy="330" r="48" fill="#fdb022"/>`+
			`<path d="M300 860 L484 440 M724 860 L540 440" stroke="#fff" stroke-width="28" stroke-linecap="round"/>`+
			`<path d="M512 830 V460" stroke="#fff" stroke-width="20" stroke-linecap="round" stroke-dasharray="60 50"/>`))
	// 8 Orbit: a planet and its moon, the days going round.
	{
		mx, my := 812.0, 512.0
		add("Orbit", doc(lin("bg", "#1e293b", "#0f172a", 0, 0, 0, 1)+lin("pl", "#60a5fa", "#2563eb", 0, 0, 1, 1), "url(#bg)",
			`<g transform="rotate(-20 512 512)">`+
				`<path d="M212 512 A300 110 0 0 1 812 512" fill="none" stroke="#fff" stroke-opacity=".35" stroke-width="20"/>`+
				`<circle cx="512" cy="512" r="130" fill="url(#pl)"/>`+
				`<path d="M812 512 A300 110 0 0 1 212 512" fill="none" stroke="#fff" stroke-opacity=".85" stroke-width="20"/>`+
				fmt.Sprintf(`<circle cx="%g" cy="%g" r="36" fill="#fdb022"/>`, mx, my)+`</g>`))
	}
	// 9 Caret: the Ctrl key's ⌃ over a point of focus.
	add("Ctrl caret", doc(blue, "url(#bg)",
		`<path d="M322 480 L512 290 L702 480" fill="none" stroke="#fff" stroke-width="80" stroke-linecap="round" stroke-linejoin="round"/>`+
			`<circle cx="512" cy="660" r="64" fill="#fdb022"/>`))
	// 10 Agenda: the day as blocks of time.
	add("Agenda", doc(paper, "url(#bg)",
		`<rect x="270" y="310" width="484" height="100" rx="50" fill="#2563eb"/>`+
			`<rect x="270" y="462" width="320" height="100" rx="50" fill="#f59e0b"/>`+
			`<rect x="270" y="614" width="420" height="100" rx="50" fill="#10b981"/>`))
	// 11 Horizon: the day rising.
	add("Horizon", doc(paper+lin("sun", "#fbbf24", "#f97316", 0, 0, 0, 1), "url(#bg)",
		`<path d="M322 600 A190 190 0 0 1 702 600 Z" fill="url(#sun)"/>`+
			`<path d="M250 600 H774" stroke="#1e3a8a" stroke-width="40" stroke-linecap="round"/>`+
			`<path d="M370 700 H654" stroke="#1e3a8a" stroke-opacity=".35" stroke-width="40" stroke-linecap="round"/>`))
	// 12 Pomodoro: the tomato the timer is named for.
	add("Pomodoro", doc(paper, "url(#bg)",
		`<circle cx="512" cy="570" r="240" fill="#ef4444"/>`+
			`<path d="M512 330 L560 280 L566 350 L630 340 L580 390 L512 370 L444 390 L394 340 L458 350 L464 280 Z" fill="#16a34a"/>`+
			`<path d="M512 330 V250" stroke="#16a34a" stroke-width="28" stroke-linecap="round"/>`+
			`<ellipse cx="420" cy="500" rx="44" ry="70" fill="#fff" fill-opacity=".3" transform="rotate(30 420 500)"/>`))
	// 13 Hourglass: time running.
	add("Hourglass", doc(ink, "url(#bg)",
		`<path d="M340 260 H684 M340 764 H684" stroke="#fff" stroke-width="40" stroke-linecap="round"/>`+
			`<path d="M376 300 H648 L512 512 Z" fill="#fff" fill-opacity=".9"/>`+
			`<path d="M512 512 L648 724 H376 Z" fill="#3b82f6"/>`))
	// 14 Dial: the clock face, hands at work.
	{
		var ticks strings.Builder
		for _, a := range []float64{0, 90, 180, 270} {
			x1, y1 := pt(c, c, 250, a)
			x2, y2 := pt(c, c, 205, a)
			fmt.Fprintf(&ticks, `<path d="M%.1f %.1f L%.1f %.1f" stroke="#fff" stroke-width="36" stroke-linecap="round"/>`, x1, y1, x2, y2)
		}
		add("Dial", doc(blue, "url(#bg)",
			`<circle cx="512" cy="512" r="300" fill="none" stroke="#fff" stroke-width="40"/>`+ticks.String()+
				`<path d="M512 380 V512 L610 572" fill="none" stroke="#fff" stroke-width="44" stroke-linecap="round" stroke-linejoin="round"/>`+
				`<circle cx="512" cy="512" r="30" fill="#fdb022"/>`))
	}
	// 15 Mast: a radio mast sending both ways.
	add("Mast", doc(navy, "url(#bg)",
		`<path d="M512 420 L420 770 M512 420 L604 770 M452 650 H572" stroke="#fff" stroke-width="32" stroke-linecap="round" stroke-linejoin="round" fill="none"/>`+
			`<circle cx="512" cy="380" r="44" fill="#fdb022"/>`+
			`<path d="`+arc(512, 380, 120, 225, 315)+`" fill="none" stroke="#fff" stroke-width="32" stroke-linecap="round"/>`+
			`<path d="`+arc(512, 380, 120, 45, 135)+`" fill="none" stroke="#fff" stroke-width="32" stroke-linecap="round"/>`+
			`<path d="`+arc(512, 380, 200, 225, 315)+`" fill="none" stroke="#fff" stroke-opacity=".5" stroke-width="32" stroke-linecap="round"/>`+
			`<path d="`+arc(512, 380, 200, 45, 135)+`" fill="none" stroke="#fff" stroke-opacity=".5" stroke-width="32" stroke-linecap="round"/>`))
	// 16 Keycap: the Ctrl key itself.
	add("Keycap", doc(lin("bg", "#eef2f8", "#d9e0ec", 0, 0, 0, 1), "url(#bg)",
		`<rect x="240" y="264" width="544" height="544" rx="120" fill="#b9c3d6"/>`+
			`<rect x="240" y="240" width="544" height="530" rx="120" fill="#fff"/>`+
			`<path d="M412 520 L512 420 L612 520" fill="none" stroke="#2563eb" stroke-width="60" stroke-linecap="round" stroke-linejoin="round"/>`))
	// 17 Venn: meetings and focus, overlapping.
	add("Venn", doc(ink, "url(#bg)",
		`<circle cx="424" cy="512" r="200" fill="none" stroke="#3b82f6" stroke-width="56"/>`+
			`<circle cx="600" cy="512" r="200" fill="none" stroke="#34d399" stroke-width="56" stroke-opacity=".9"/>`))
	// 18 Windsock: the airfield's, telling which way things go.
	{
		xs := []float64{340, 440, 540, 640, 740}
		top := func(x float64) float64 { return 290 + (x-340)*50/400 }
		bot := func(x float64) float64 { return 450 - (x-340)*50/400 }
		var segs strings.Builder
		for i := 0; i < 4; i++ {
			fill := "#f97316"
			if i%2 == 1 {
				fill = "#ffffff"
			}
			a, b := xs[i], xs[i+1]
			fmt.Fprintf(&segs, `<path d="M%g %g L%g %g L%g %g L%g %g Z" fill="%s"/>`, a, top(a), b, top(b), b, bot(b), a, bot(a), fill)
		}
		add("Windsock", doc(lin("bg", "#60a5fa", "#2563eb", 0, 0, 0, 1), "url(#bg)",
			`<path d="M330 250 V790" stroke="#fff" stroke-width="32" stroke-linecap="round"/>`+segs.String()))
	}
	// 19 Knob: a control turned up.
	{
		var dots strings.Builder
		for a := -135.0; a <= 135; a += 27 {
			x, y := pt(c, c, 300, a)
			fill, op := "#2563eb", "1"
			if a > 45 {
				fill, op = "#94a3b8", ".5"
			}
			fmt.Fprintf(&dots, `<circle cx="%.1f" cy="%.1f" r="18" fill="%s" fill-opacity="%s"/>`, x, y, fill, op)
		}
		x1, y1 := pt(c, c, 80, 45)
		x2, y2 := pt(c, c, 165, 45)
		add("Knob", doc(paper, "url(#bg)",
			dots.String()+`<circle cx="512" cy="530" r="210" fill="#cbd5e1"/><circle cx="512" cy="512" r="210" fill="#fff"/>`+
				fmt.Sprintf(`<path d="M%.1f %.1f L%.1f %.1f" stroke="#2563eb" stroke-width="40" stroke-linecap="round"/>`, x1, y1, x2, y2)))
	}
	// 20 Spark: a point of focus in its ring.
	add("Spark", doc(violet, "url(#bg)",
		`<circle cx="512" cy="512" r="280" fill="none" stroke="#fff" stroke-opacity=".3" stroke-width="24"/>`+
			`<path d="M512 292 Q536 488 732 512 Q536 536 512 732 Q488 536 292 512 Q488 488 512 292 Z" fill="#fff"/>`))
	return out
}

func main() {
	round := flag.Int("round", 1, "the round to render")
	final := flag.Bool("final", false, "export the chosen marks to design/logo/final")
	flag.Parse()
	if *final {
		exportFinal()
		return
	}
	set := logos[*round]
	if set == nil {
		log.Fatalf("no round %d", *round)
	}
	dir := filepath.Join("design", "logo", fmt.Sprintf("round%d", *round))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	sort.Slice(set, func(i, j int) bool { return set[i].id < set[j].id })
	svgs := make([]*ui.SVG, len(set))
	for i, l := range set {
		base := filepath.Join(dir, l.id+"-"+slug(l.name))
		if err := os.WriteFile(base+".svg", []byte(l.svg), 0o644); err != nil {
			log.Fatal(err)
		}
		svgs[i] = ui.MustParseSVG([]byte(l.svg))
		save(base+".png", 512, 512, func(c *ui.Context) {
			ui.Image(c, svgs[i]).Size(512, 512)
		})
	}
	// The contact sheet: each at app-tile size, and at 32 and 16 px, as
	// the taskbar and tray show it, on light and dark.
	const cols, cell = 5, 260
	rows := (len(set) + cols - 1) / cols
	save(filepath.Join(dir, "sheet.png"), cols*cell, rows*(cell+40), func(c *ui.Context) {
		c.Root().Background(ui.Hex("#f3f4f7"))
		ui.Row(c).Wrap().Fill().Children(func() {
			for i, l := range set {
				ui.Column(c.Key(i)).Size(cell, cell+40).Padding(10).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Image(c, svgs[i]).Size(180, 180)
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Padding(4).Radius(6).Background(ui.Hex("#ffffff")).Children(func() { ui.Image(c, svgs[i]).Size(32, 32) })
						ui.Box(c).Padding(4).Radius(6).Background(ui.Hex("#15171e")).Children(func() { ui.Image(c, svgs[i]).Size(32, 32) })
						ui.Box(c).Padding(4).Radius(6).Background(ui.Hex("#15171e")).Children(func() { ui.Image(c, svgs[i]).Size(16, 16) })
					})
					ui.Text(c, l.id+" · "+l.name).FontSize(13).FontWeight(600).TextColor(ui.Hex("#0e1117"))
				})
			}
		})
	})
	if err := zoomSheet(filepath.Join(dir, "zoom.png"), svgs); err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", dir)
}

func save(path string, w, h int, view func(c *ui.Context)) {
	// A transparent root, so the tile's rounded corners show what is under
	// the icon rather than the window's white.
	tt := ui.NewTester(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		view(c)
	}, w, h)
	tt.Frame()
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		log.Fatal(err)
	}
}

func slug(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}
