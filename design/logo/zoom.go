package main

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/egoist/mygo/ui"
)

// zoomSheet renders each logo at its true 16 and 32 px on light and dark
// backgrounds, then blows the pixels up eight times, nearest neighbour, so
// what the tray and taskbar will really show can be inspected.
func zoomSheet(path string, svgs []*ui.SVG) error {
	const k = 8
	bgs := []color.RGBA{{243, 244, 247, 255}, {21, 23, 30, 255}}
	sizes := []int{16, 32}
	cellW := 0
	for _, s := range sizes {
		cellW += s*k + 16
	}
	rowH := 32*k + 16
	out := image.NewRGBA(image.Rect(0, 0, cellW*len(bgs), rowH*len(svgs)))
	for i, s := range svgs {
		for b, bg := range bgs {
			x := b * cellW
			for _, n := range sizes {
				tt := ui.NewTester(func(c *ui.Context) {
					c.Root().Background(ui.RGB(bg.R, bg.G, bg.B))
					ui.Image(c, s).Size(float32(n), float32(n))
				}, n, n)
				tt.Frame()
				img := tt.Image()
				for py := 0; py < rowH; py++ {
					for px := 0; px < n*k+16; px++ {
						out.Set(x+px, i*rowH+py, bg)
					}
				}
				oy := i*rowH + 8 + (32-n)*k/2
				for py := 0; py < n*k; py++ {
					for px := 0; px < n*k; px++ {
						out.Set(x+8+px, oy+py, img.At(px/k, py/k))
					}
				}
				x += n*k + 16
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, out)
}
