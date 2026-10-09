package nyan_test

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	nyan "github.com/timzifer/fyne-nyan"
	"github.com/timzifer/fyne-nyan/internal/showcase"
)

var shots = flag.Bool("shots", false, "render the README screenshots into docs/")

const (
	shotScale = 2
	frameTime = 70 * time.Millisecond
)

// TestShots renders the README images: go test -run Shots -shots
func TestShots(t *testing.T) {
	if !*shots {
		t.Skip("use -shots to render the README images")
	}
	app := test.NewTempApp(t)
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	defer nyan.SetClock(func() time.Time { return clock })()

	if err := os.MkdirAll("docs", 0o755); err != nil {
		t.Fatal(err)
	}

	// full showcase in both variants
	for _, v := range []string{"Light", "Dark"} {
		s := showcase.New(app)
		s.Variants.SetSelected(v)
		s.Live.SetValue(0.35)
		c := newCanvas(container.NewPadded(s.Content), 720)
		clock = clock.Add(frameTime * 9)
		animate(s.Content)
		savePNG(t, "docs/showcase-"+strings.ToLower(v)+".png", c.Capture())
	}

	// animated hero
	app.Settings().SetTheme(showcase.Variant{Theme: theme.DefaultTheme(), Variant: theme.VariantDark})
	bar := nyan.NewProgressBar()
	inf := nyan.NewProgressBarInfinite()
	content := container.NewPadded(container.NewVBox(
		widget.NewLabel("nyan.ProgressBar"), bar,
		widget.NewLabel("nyan.ProgressBarInfinite"), inf,
	))
	c := newCanvas(content, 560)

	const frames, hold = 84, 14
	var images []image.Image
	for i := 0; i < frames; i++ {
		bar.SetValue(min(1, float64(i)/float64(frames-hold)))
		animate(content)
		images = append(images, c.Capture())
		clock = clock.Add(frameTime)
	}
	saveGIF(t, "docs/hero.gif", images)
}

func newCanvas(content fyne.CanvasObject, width float32) fyne.Canvas {
	c := software.NewCanvas()
	c.SetScale(shotScale)
	c.SetPadded(false)
	c.SetContent(content)
	c.Resize(fyne.NewSize(width, content.MinSize().Height))
	return c
}

// animate steps every nyan widget below o.
func animate(o fyne.CanvasObject) {
	switch w := o.(type) {
	case *nyan.ProgressBar, *nyan.ProgressBarInfinite:
		nyan.Animate(w.(fyne.Widget))
	case *fyne.Container:
		for _, child := range w.Objects {
			animate(child)
		}
	case fyne.Widget:
		for _, child := range test.WidgetRenderer(w).Objects() {
			animate(child)
		}
	}
}

func savePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(filepath.FromSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// saveGIF quantises all frames to one shared palette of the most used colours.
func saveGIF(t *testing.T, path string, frames []image.Image) {
	t.Helper()
	counts := map[color.NRGBA]int{}
	for _, img := range frames {
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				counts[color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)]++
			}
		}
	}
	colors := make([]color.NRGBA, 0, len(counts))
	for c := range counts {
		colors = append(colors, c)
	}
	sort.Slice(colors, func(i, j int) bool { return counts[colors[i]] > counts[colors[j]] })
	pal := make(color.Palette, 0, 256)
	for _, c := range colors[:min(256, len(colors))] {
		pal = append(pal, c)
	}

	anim := &gif.GIF{}
	for _, img := range frames {
		p := image.NewPaletted(img.Bounds(), pal)
		draw.Draw(p, p.Rect, img, img.Bounds().Min, draw.Src)
		anim.Image = append(anim.Image, p)
		anim.Delay = append(anim.Delay, int(frameTime/(10*time.Millisecond)))
	}
	f, err := os.Create(filepath.FromSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := gif.EncodeAll(f, anim); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
