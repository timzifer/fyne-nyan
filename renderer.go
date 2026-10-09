package nyan

import (
	"image"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

const (
	frameDuration = 70 * time.Millisecond
	flightSpeed   = 60.0 // units per second in indeterminate mode
)

// now is replaced in tests to get deterministic frames.
var now = time.Now

var colorPill = color.NRGBA{A: 0xa0}

// renderer is shared by both progress bar flavours.
type renderer struct {
	owner    fyne.Widget
	raster   *canvas.Raster
	pill     *canvas.Rectangle
	label    *canvas.Text
	showText bool

	scene     scene
	animation *fyne.Animation
	epoch     time.Time
	running   bool
}

func newRenderer(owner fyne.Widget, showText bool) *renderer {
	r := &renderer{owner: owner, showText: showText}
	r.raster = canvas.NewRaster(r.generate)
	r.raster.ScaleMode = canvas.ImageScalePixels
	r.pill = canvas.NewRectangle(colorPill)
	r.label = canvas.NewText("", color.White)
	r.label.Alignment = fyne.TextAlignCenter
	r.label.TextStyle.Bold = true
	r.pill.Hidden = !showText
	r.label.Hidden = !showText
	r.animation = &fyne.Animation{
		Duration:    time.Second,
		RepeatCount: fyne.AnimationRepeatForever,
		Tick:        func(float32) { r.step() },
	}
	r.epoch = now()
	return r
}

func (r *renderer) generate(w, h int) image.Image {
	sc := r.scene
	if size := r.owner.Size(); size.Height > 0 {
		sc.radius = r.theme().Size(theme.SizeNameInputRadius) * float32(h) / size.Height
	}
	return paint(w, h, sc)
}

func (r *renderer) theme() fyne.Theme {
	if t, ok := r.owner.(interface{ Theme() fyne.Theme }); ok {
		return t.Theme()
	}
	return theme.Current()
}

// step advances the animation and repaints when the picture changed.
func (r *renderer) step() {
	if !r.running || !r.owner.Visible() {
		return
	}
	elapsed := now().Sub(r.epoch)
	tick := int(elapsed / frameDuration)
	travel := float64(int(elapsed.Seconds() * flightSpeed))
	if tick == r.scene.tick && (!r.scene.infinite || travel == r.scene.travel) {
		return
	}
	r.scene.tick = tick
	if r.scene.infinite {
		r.scene.travel = travel
	}
	r.raster.Refresh()
}

func (r *renderer) start() {
	if r.running {
		return
	}
	r.running = true
	// continue where we stopped instead of jumping ahead
	r.epoch = now().Add(-time.Duration(r.scene.tick) * frameDuration)
	if r.scene.infinite {
		r.epoch = now().Add(-time.Duration(r.scene.travel / flightSpeed * float64(time.Second)))
	}
	r.animation.Start()
}

func (r *renderer) stop() {
	if !r.running {
		return
	}
	r.running = false
	r.animation.Stop()
}

func (r *renderer) setText(text string) {
	r.label.Text = text
	th := r.theme()
	r.label.TextSize = th.Size(theme.SizeNameText)
	r.Layout(r.owner.Size())
	r.label.Refresh()
	r.pill.Refresh()
}

func (r *renderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
	if !r.showText {
		return
	}
	pad := r.theme().Size(theme.SizeNameInnerPadding)
	text := fyne.MeasureText(r.label.Text, r.label.TextSize, r.label.TextStyle)
	pill := fyne.NewSize(text.Width+pad*1.5, text.Height)
	r.pill.CornerRadius = pill.Height / 2
	r.pill.Resize(pill)

	// the readout flies just ahead of the cat, near the end it moves behind
	scale := r.scale()
	// rasters are generated at ceil(size*scale) pixels, mirror that here
	unit, cols := grid(int(math.Ceil(float64(size.Width*scale))), int(math.Ceil(float64(size.Height*scale))))
	left := float32(r.scene.catX(cols)*unit) / scale
	x := left + float32(catWidth*unit)/scale + pad/2
	if x+pill.Width > size.Width-pad/2 {
		x = left + float32(bodyLeft*unit)/scale - pill.Width - pad/2
	}
	x = max(x, pad/2)
	pos := fyne.NewPos(x, (size.Height-pill.Height)/2)
	r.pill.Move(pos)
	r.label.Resize(pill)
	r.label.Move(pos)
}

func (r *renderer) scale() float32 {
	if c := fyne.CurrentApp().Driver().CanvasForObject(r.owner); c != nil {
		return c.Scale()
	}
	return 1
}

func (r *renderer) MinSize() fyne.Size {
	th := r.theme()
	pad := th.Size(theme.SizeNameInnerPadding) * 2
	text := fyne.MeasureText("100%", th.Size(theme.SizeNameText), fyne.TextStyle{Bold: true})
	size := text.AddWidthHeight(pad, pad)
	// leave room for the cat and a bit of rainbow
	unit := size.Height / gridRows
	return fyne.NewSize(max(size.Width, unit*(catWidth+tailTuck*2)), size.Height)
}

func (r *renderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster, r.pill, r.label}
}

func (r *renderer) Refresh() {
	r.raster.Refresh()
}

func (r *renderer) Destroy() {
	r.stop()
}

var _ fyne.WidgetRenderer = (*renderer)(nil)
