package nyan

import (
	"errors"
	"image"
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// fakeClock replaces the animation clock for the duration of a test.
func fakeClock(t *testing.T) *time.Time {
	t.Helper()
	cur := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	old := now
	now = func() time.Time { return cur }
	t.Cleanup(func() { now = old })
	return &cur
}

func TestNewProgressBar(t *testing.T) {
	test.NewTempApp(t)

	bar := NewProgressBar()
	if bar.Min != 0 || bar.Max != 1 || bar.Value != 0 {
		t.Fatalf("unexpected defaults: %+v", bar)
	}

	zero := &ProgressBar{}
	test.WidgetRenderer(zero)
	if zero.Max != 1 {
		t.Errorf("zero value bar should default Max to 1, got %v", zero.Max)
	}
}

func TestProgressBar_Value(t *testing.T) {
	test.NewTempApp(t)
	bar := NewProgressBar()
	r := test.WidgetRenderer(bar).(*progressRenderer)

	for _, tt := range []struct {
		in, value float64
		text      string
	}{
		{0, 0, "0%"},
		{0.5, 0.5, "50%"},
		{0.999, 0.999, "99%"},
		{1, 1, "100%"},
		{2, 1, "100%"},
		{-1, 0, "0%"},
	} {
		bar.SetValue(tt.in)
		if bar.Value != tt.value {
			t.Errorf("SetValue(%v): value = %v, want %v", tt.in, bar.Value, tt.value)
		}
		if r.label.Text != tt.text {
			t.Errorf("SetValue(%v): text = %q, want %q", tt.in, r.label.Text, tt.text)
		}
	}
}

func TestProgressBar_Range(t *testing.T) {
	test.NewTempApp(t)
	bar := &ProgressBar{Min: 10, Max: 20, Value: 15}
	r := test.WidgetRenderer(bar).(*progressRenderer)
	if r.scene.ratio != 0.5 || r.label.Text != "50%" {
		t.Errorf("ratio = %v, text = %q", r.scene.ratio, r.label.Text)
	}

	bar.Max = bar.Min
	bar.Refresh()
	if r.scene.ratio != 0 {
		t.Errorf("empty range should give ratio 0, got %v", r.scene.ratio)
	}
}

func TestProgressBar_TextFormatter(t *testing.T) {
	test.NewTempApp(t)
	bar := NewProgressBar()
	bar.TextFormatter = func() string { return "nyan " + time.Duration(bar.Value*float64(time.Second)).String() }
	r := test.WidgetRenderer(bar).(*progressRenderer)

	bar.SetValue(0.25)
	if r.label.Text != "nyan 250ms" {
		t.Errorf("text = %q", r.label.Text)
	}
}

func TestProgressBar_Binding(t *testing.T) {
	test.NewTempApp(t)
	data := binding.NewFloat()
	bar := NewProgressBarWithData(data)
	r := test.WidgetRenderer(bar).(*progressRenderer)

	if err := data.Set(0.5); err != nil {
		t.Fatal(err)
	}
	if bar.Value != 0.5 || r.label.Text != "50%" {
		t.Errorf("bound value = %v, text = %q", bar.Value, r.label.Text)
	}

	other := binding.NewFloat()
	bar.Bind(other)
	_ = data.Set(0.1)
	if bar.Value != 0 {
		t.Errorf("rebinding should drop the old source and show the new one, got %v", bar.Value)
	}

	bar.Unbind()
	_ = other.Set(0.8)
	if bar.Value != 0 {
		t.Errorf("unbound bar should keep its value, got %v", bar.Value)
	}
	bar.Unbind() // second call must be harmless
}

type brokenFloat struct{ binding.Float }

func (brokenFloat) Get() (float64, error) { return 0, errBroken }

var errBroken = errors.New("broken")

func TestProgressBar_BindingError(t *testing.T) {
	test.NewTempApp(t)
	bar := NewProgressBar()
	bar.SetValue(0.3)
	bar.Bind(brokenFloat{binding.NewFloat()})
	if bar.Value != 0.3 {
		t.Errorf("failing data source must not change the value, got %v", bar.Value)
	}
}

func TestMinSize(t *testing.T) {
	test.NewTempApp(t)
	th := test.Theme()
	pad := th.Size(theme.SizeNameInnerPadding) * 2
	text := fyne.MeasureText("100%", th.Size(theme.SizeNameText), fyne.TextStyle{Bold: true})

	for name, w := range map[string]fyne.Widget{"determinate": NewProgressBar(), "infinite": NewProgressBarInfinite()} {
		size := w.MinSize()
		if size.Height != text.Height+pad {
			t.Errorf("%s: height = %v, want %v", name, size.Height, text.Height+pad)
		}
		if unit := size.Height / gridRows; size.Width < unit*catWidth {
			t.Errorf("%s: width %v is too small for the cat", name, size.Width)
		}
	}
}

func TestAnimation(t *testing.T) {
	test.NewTempApp(t)
	clock := fakeClock(t)
	bar := NewProgressBar()
	r := test.WidgetRenderer(bar).(*progressRenderer)

	*clock = clock.Add(frameDuration * 3)
	r.step()
	if r.scene.tick != 3 {
		t.Errorf("tick = %d, want 3", r.scene.tick)
	}

	bar.Hide()
	*clock = clock.Add(frameDuration)
	r.step()
	if r.scene.tick != 3 {
		t.Errorf("hidden bar should not animate, tick = %d", r.scene.tick)
	}
	bar.Show()
	r.step()
	if r.scene.tick != 4 {
		t.Errorf("tick = %d, want 4", r.scene.tick)
	}

	r.Destroy()
	if r.running {
		t.Error("destroy should stop the animation")
	}
}

func TestProgressBarInfinite(t *testing.T) {
	test.NewTempApp(t)
	clock := fakeClock(t)
	bar := NewProgressBarInfinite()
	r := test.WidgetRenderer(bar).(*infiniteRenderer)

	if !bar.Running() || !r.running {
		t.Fatal("infinite bar should start running")
	}
	bar.Start() // no-op while running

	*clock = clock.Add(time.Second)
	r.step()
	if r.scene.travel != flightSpeed {
		t.Errorf("travel = %v, want %v", r.scene.travel, flightSpeed)
	}

	bar.Stop()
	if bar.Running() || r.running {
		t.Error("Stop should stop the animation")
	}
	bar.Stop() // no-op while stopped

	*clock = clock.Add(time.Second)
	bar.Start()
	r.step()
	if r.scene.travel != flightSpeed {
		t.Errorf("restart should resume from the old position, travel = %v", r.scene.travel)
	}

	bar.Hide()
	if bar.Running() {
		t.Error("Hide should stop the bar")
	}
	bar.Show()
	if !bar.Running() {
		t.Error("Show should start the bar")
	}

	r.Destroy()
	if r.running {
		t.Error("Destroy should stop the animation")
	}
}

func TestProgressBarInfinite_StoppedBeforeShown(t *testing.T) {
	test.NewTempApp(t)
	bar := NewProgressBarInfinite()
	bar.Stop()
	r := test.WidgetRenderer(bar).(*infiniteRenderer)
	if bar.Running() || r.running {
		t.Error("a bar stopped before rendering must stay stopped")
	}

	zero := &ProgressBarInfinite{}
	if !zero.Running() || !test.WidgetRenderer(zero).(*infiniteRenderer).running {
		t.Error("zero value bar should run")
	}
}

func TestPaint_Empty(t *testing.T) {
	if img := paint(0, 10, scene{}); !img.Rect.Empty() {
		t.Errorf("expected empty image, got %v", img.Rect)
	}
}

// rainbowLength measures how far the red band reaches on a mid-height row.
func rainbowLength(img *image.NRGBA) int {
	n := 0
	for x := 0; x < img.Rect.Dx(); x++ {
		for y := 0; y < img.Rect.Dy(); y++ {
			if img.NRGBAAt(x, y) == rainbow[0] {
				n = x + 1
				break
			}
		}
	}
	return n
}

func TestPaint_Progress(t *testing.T) {
	prev := -1
	for _, r := range []float64{0, 0.25, 0.5, 0.75, 1} {
		l := rainbowLength(paint(400, 36, scene{ratio: r}))
		if l <= prev {
			t.Errorf("rainbow should grow with progress: ratio %v gives %d after %d", r, l, prev)
		}
		prev = l
	}
	if a, b := paint(400, 36, scene{ratio: 7}), paint(400, 36, scene{ratio: 1}); rainbowLength(a) != rainbowLength(b) {
		t.Error("ratio should be clamped to 1")
	}
}

func TestPaint_Infinite(t *testing.T) {
	// 400px at 2px per unit is 200 columns
	cols := 200
	cycle := cols + catWidth
	for _, tt := range []struct {
		travel float64
		catX   int
	}{
		{0, 0},
		{float64(cols - 1), cols - 1},
		{float64(cols), -catWidth},
		{float64(cycle), 0},
	} {
		if x := (scene{infinite: true, travel: tt.travel}).catX(cols); x != tt.catX {
			t.Errorf("travel %v: cat at %d, want %d", tt.travel, x, tt.catX)
		}
	}

	// the flag is finite: in the middle of the flight it does not reach the left edge
	img := paint(400, 36, scene{infinite: true, travel: float64(cols / 2)})
	if c := img.NRGBAAt(4, 36/2); c != colorSpace {
		t.Errorf("expected space at the left edge, got %v", c)
	}
	if rainbowLength(img) == 0 {
		t.Error("expected a rainbow behind the cat")
	}

	// after wrapping, the old flag is still leaving on the right
	img = paint(400, 36, scene{infinite: true, travel: float64(cols + 2)})
	if rainbowLength(img) < 400-20 {
		t.Errorf("old flag should still be visible on the right, rainbow ends at %d", rainbowLength(img))
	}
}

func TestPaint_Fade(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	fill(img, 0, 0, 4, 4, colorSpace)
	blend(img, 0, 0, 2, 2, colorStar, 0.5)
	if got, want := img.NRGBAAt(0, 0), mix(colorSpace, colorStar, 0.5); got != want {
		t.Errorf("blend = %v, want %v", got, want)
	}
	if img.NRGBAAt(3, 3) != colorSpace {
		t.Error("blend should not touch pixels outside the region")
	}
}

func TestPaint_RoundCorners(t *testing.T) {
	img := paint(100, 36, scene{radius: 8})
	if a := img.NRGBAAt(0, 0).A; a != 0 {
		t.Errorf("corner should be transparent, alpha %d", a)
	}
	if c := img.NRGBAAt(50, 1); c.A != 0xff {
		t.Errorf("edge centre should be opaque, got %v", c)
	}
	sharp := paint(100, 36, scene{})
	if c := sharp.NRGBAAt(0, 0); c != colorSpace {
		t.Errorf("no radius should keep the corner, got %v", c)
	}
}

func TestCatSprites(t *testing.T) {
	for f, s := range catSprites {
		var used int
		for _, c := range s.px {
			if c == 0 {
				continue
			}
			if _, ok := palette[c]; !ok {
				t.Fatalf("frame %d uses unknown colour %q", f, c)
			}
			used++
		}
		if used == 0 {
			t.Errorf("frame %d is empty", f)
		}
	}
	if catSprites[0].at(0, 0) != 0 {
		t.Error("top left of the sprite should be transparent")
	}
	if catSprites[0] == catSprites[2] {
		t.Error("frames should be distinct")
	}
	if c := catSprites[0].at(catWidth-1, 0); c != 0 {
		t.Errorf("unexpected pixel %q", c)
	}
	var s sprite
	s.set(-1, -1, 'k') // out of range must not panic
}

func TestRendering(t *testing.T) {
	test.NewTempApp(t)
	clock := fakeClock(t)

	for _, v := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		name := "light"
		if v == theme.VariantDark {
			name = "dark"
		}
		fyne.CurrentApp().Settings().SetTheme(fixedVariant{theme.DefaultTheme(), v})

		bar := NewProgressBar()
		bar.SetValue(0.42)
		inf := NewProgressBarInfinite()
		w := test.NewWindow(nil)
		w.SetContent(container.NewVBox(bar, inf))
		w.Resize(fyne.NewSize(320, 100))

		*clock = clock.Add(frameDuration * 5)
		test.WidgetRenderer(bar).(*progressRenderer).step()
		test.WidgetRenderer(inf).(*infiniteRenderer).step()
		test.AssertRendersToImage(t, "bars_"+name+".png", w.Canvas())
		w.Close()
	}
}

// fixedVariant forces a theme variant regardless of the system setting.
type fixedVariant struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (f fixedVariant) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(n, f.variant)
}
