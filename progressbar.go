// Package nyan provides Fyne progress bars starring a pop-tart cat that
// flies through space and leaves a rainbow behind. Somebody had to do it.
//
// The artwork is a homage to Nyan Cat by Chris Torres, drawn from scratch.
//
// ProgressBar and ProgressBarInfinite are drop-in replacements for
// widget.ProgressBar and widget.ProgressBarInfinite.
package nyan

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// ProgressBar is a determinate progress bar. The cat flies from left to
// right as the value grows from Min to Max.
type ProgressBar struct {
	widget.BaseWidget

	Min, Max, Value float64

	// TextFormatter can be used to have a custom format of progress text.
	// If set, it overrides the percentage readout and runs each time the value updates.
	TextFormatter func() string `json:"-"`

	data     binding.Float
	listener binding.DataListener
}

// NewProgressBar creates a new progress bar widget.
// The default Min is 0 and Max is 1, Values set should be between those numbers.
// The display will convert this to a percentage.
func NewProgressBar() *ProgressBar {
	p := &ProgressBar{Min: 0, Max: 1}
	p.ExtendBaseWidget(p)
	return p
}

// NewProgressBarWithData returns a progress bar connected with the specified data source.
func NewProgressBarWithData(data binding.Float) *ProgressBar {
	p := NewProgressBar()
	p.Bind(data)
	return p
}

// Bind connects the specified data source to this ProgressBar.
// The current value will be displayed and any changes in the data will cause the widget to update.
func (p *ProgressBar) Bind(data binding.Float) {
	p.Unbind()
	listener := binding.NewDataListener(func() {
		v, err := data.Get()
		if err != nil {
			fyne.LogError("Error getting current data value", err)
			return
		}
		p.SetValue(v)
	})
	p.data, p.listener = data, listener
	data.AddListener(listener)
}

// Unbind disconnects any configured data source from this ProgressBar.
// The current value will remain at the last value of the data source.
func (p *ProgressBar) Unbind() {
	if p.data != nil {
		p.data.RemoveListener(p.listener)
	}
	p.data, p.listener = nil, nil
}

// SetValue changes the current value of this progress bar (from p.Min to p.Max).
// The widget will be refreshed to indicate the change.
func (p *ProgressBar) SetValue(v float64) {
	p.Value = v
	p.Refresh()
}

// MinSize returns the size that this widget should not shrink below.
func (p *ProgressBar) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (p *ProgressBar) CreateRenderer() fyne.WidgetRenderer {
	p.ExtendBaseWidget(p)
	if p.Min == 0 && p.Max == 0 {
		p.Max = 1
	}
	r := &progressRenderer{renderer: newRenderer(p, true), bar: p}
	r.update()
	r.start()
	return r
}

// ratio clamps the value into range and returns the progress from 0 to 1.
func (p *ProgressBar) ratio() float64 {
	p.Value = min(max(p.Value, p.Min), p.Max)
	delta := p.Max - p.Min
	if delta <= 0 {
		return 0
	}
	return (p.Value - p.Min) / delta
}

type progressRenderer struct {
	*renderer
	bar *ProgressBar
}

func (r *progressRenderer) update() {
	r.scene.ratio = r.bar.ratio()
	text := strconv.Itoa(int(r.scene.ratio*100)) + "%"
	if r.bar.TextFormatter != nil {
		text = r.bar.TextFormatter()
	}
	r.setText(text)
}

func (r *progressRenderer) Refresh() {
	r.update()
	r.renderer.Refresh()
}
