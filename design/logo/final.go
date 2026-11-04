package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/ui"
)

// exportFinal writes the chosen marks to design/logo/final, each as SVG
// and as PNGs at the sizes it is for (and twice them, for high DPI).
func exportFinal() {
	r7 := map[string]logo{}
	for _, l := range round7() {
		r7[l.id] = l
	}
	dir := filepath.Join("design", "logo", "final")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	for _, f := range []struct {
		id, name string
		sizes    []int
	}{
		{"23a", "app-icon", []int{1024, 256}},
		{"24b", "taskbar", []int{32, 64}},
		{"24c", "tray", []int{16, 32}},
		{"24d", "tray-alt-ring", []int{16, 32}},
	} {
		svg := r7[f.id].svg
		if err := os.WriteFile(filepath.Join(dir, f.name+".svg"), []byte(svg), 0o644); err != nil {
			log.Fatal(err)
		}
		s := ui.MustParseSVG([]byte(svg))
		for _, n := range f.sizes {
			save(filepath.Join(dir, f.name+"-"+itoa(n)+".png"), n, n, func(c *ui.Context) {
				ui.Image(c, s).Size(float32(n), float32(n))
			})
		}
	}
}

func itoa(n int) string {
	b := []byte{}
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}
