package nyan

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ProgressBarInfinite is an indeterminate progress bar. The cat keeps
// flying from left to right, leaves the bar and comes back in on the left,
// until Stop() is called.
type ProgressBarInfinite struct {
	widget.BaseWidget

	// stopped is inverted so the zero value is a running bar.
	stopped bool
}

// NewProgressBarInfinite creates a new progress bar widget that flies indefinitely.
func NewProgressBarInfinite() *ProgressBarInfinite {
	p := &ProgressBarInfinite{}
	p.ExtendBaseWidget(p)
	return p
}

// Show this widget, if it was previously hidden.
func (p *ProgressBarInfinite) Show() {
	p.stopped = false
	p.BaseWidget.Show()
}

// Hide this widget, if it was previously visible.
func (p *ProgressBarInfinite) Hide() {
	p.stopped = true
	p.BaseWidget.Hide()
}

// Start the animation.
func (p *ProgressBarInfinite) Start() {
	if !p.stopped {
		return
	}
	p.stopped = false
	p.Refresh()
}

// Stop the animation, the cat stays where it is.
// It can be called before the bar is shown to start in the stopped state.
func (p *ProgressBarInfinite) Stop() {
	if p.stopped {
		return
	}
	p.stopped = true
	p.Refresh()
}

// Running returns the current state of the animation.
func (p *ProgressBarInfinite) Running() bool {
	return !p.stopped
}

// MinSize returns the size that this widget should not shrink below.
func (p *ProgressBarInfinite) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (p *ProgressBarInfinite) CreateRenderer() fyne.WidgetRenderer {
	p.ExtendBaseWidget(p)
	r := &infiniteRenderer{renderer: newRenderer(p, false), bar: p}
	r.scene.infinite = true
	r.Refresh()
	return r
}

type infiniteRenderer struct {
	*renderer
	bar *ProgressBarInfinite
}

func (r *infiniteRenderer) Refresh() {
	if r.bar.Running() {
		r.start()
	} else {
		r.stop()
	}
	r.renderer.Refresh()
}
