// Package showcase builds the demo UI that shows every state of the nyan
// progress bars. It is used by the example app and to render the README
// screenshots.
package showcase

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	nyan "github.com/timzifer/fyne-nyan"
)

// Showcase holds the demo content and the widgets that tests drive.
type Showcase struct {
	Content fyne.CanvasObject

	Static   []*nyan.ProgressBar
	Custom   *nyan.ProgressBar
	Bound    *nyan.ProgressBar
	Data     binding.Float
	Slider   *widget.Slider
	Live     *nyan.ProgressBar
	Running  *nyan.ProgressBarInfinite
	Stopped  *nyan.ProgressBarInfinite
	Toggled  *nyan.ProgressBarInfinite
	Variants *widget.RadioGroup

	live    bool
	liveBtn *widget.Button
}

// New builds the showcase. Live updates only run after Start is called.
func New(app fyne.App) *Showcase {
	s := &Showcase{}

	// determinate
	determinate := container.New(layout.NewFormLayout())
	for _, v := range []float64{0, 0.25, 0.5, 0.75, 1} {
		bar := nyan.NewProgressBar()
		bar.SetValue(v)
		s.Static = append(s.Static, bar)
		determinate.Add(label(fmt.Sprintf("Value %.2f", v)))
		determinate.Add(bar)
	}

	s.Custom = &nyan.ProgressBar{Min: 0, Max: 10, Value: 7}
	s.Custom.TextFormatter = func() string {
		return fmt.Sprintf("%.0f / %.0f pop-tarts", s.Custom.Value, s.Custom.Max)
	}
	s.Custom.ExtendBaseWidget(s.Custom)
	determinate.Add(label("TextFormatter"))
	determinate.Add(s.Custom)

	s.Data = binding.NewFloat()
	_ = s.Data.Set(0.6)
	s.Bound = nyan.NewProgressBarWithData(s.Data)
	// set the step before binding, else the initial value snaps to 0 or 1
	s.Slider = widget.NewSlider(0, 1)
	s.Slider.Step = 0.01
	s.Slider.Bind(s.Data)
	determinate.Add(label("Data binding"))
	determinate.Add(container.NewVBox(s.Bound, s.Slider))

	s.Live = nyan.NewProgressBar()
	s.liveBtn = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), s.toggleLive)
	reset := widget.NewButtonWithIcon("Reset", theme.MediaReplayIcon(), func() { s.Live.SetValue(0) })
	determinate.Add(label("Live"))
	determinate.Add(container.NewBorder(nil, nil, nil, container.NewHBox(s.liveBtn, reset), s.Live))

	// indeterminate
	indeterminate := container.New(layout.NewFormLayout())
	s.Running = nyan.NewProgressBarInfinite()
	indeterminate.Add(label("Running"))
	indeterminate.Add(s.Running)

	s.Stopped = nyan.NewProgressBarInfinite()
	s.Stopped.Stop()
	var startStop *widget.Button
	startStop = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
		if s.Stopped.Running() {
			s.Stopped.Stop()
			startStop.SetText("Start")
			startStop.SetIcon(theme.MediaPlayIcon())
		} else {
			s.Stopped.Start()
			startStop.SetText("Stop")
			startStop.SetIcon(theme.MediaStopIcon())
		}
	})
	indeterminate.Add(label("Start / Stop"))
	indeterminate.Add(container.NewBorder(nil, nil, nil, startStop, s.Stopped))

	s.Toggled = nyan.NewProgressBarInfinite()
	visible := widget.NewCheck("Visible", func(on bool) {
		if on {
			s.Toggled.Show()
		} else {
			s.Toggled.Hide()
		}
	})
	visible.SetChecked(true)
	indeterminate.Add(label("Show / Hide"))
	indeterminate.Add(container.NewBorder(nil, nil, nil, visible, s.Toggled))

	// stock widgets for comparison
	stockBar := widget.NewProgressBar()
	stockBar.SetValue(0.5)
	stock := container.New(layout.NewFormLayout(),
		label("ProgressBar"), stockBar,
		label("ProgressBarInfinite"), widget.NewProgressBarInfinite(),
	)

	// theme switch
	s.Variants = widget.NewRadioGroup([]string{"System", "Light", "Dark"}, func(v string) {
		switch v {
		case "Light":
			app.Settings().SetTheme(Variant{theme.DefaultTheme(), theme.VariantLight})
		case "Dark":
			app.Settings().SetTheme(Variant{theme.DefaultTheme(), theme.VariantDark})
		default:
			app.Settings().SetTheme(theme.DefaultTheme())
		}
	})
	s.Variants.Horizontal = true
	s.Variants.SetSelected("System")

	title := widget.NewLabelWithStyle("fyne-nyan", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.SizeName = theme.SizeNameHeadingText

	s.Content = container.NewVBox(
		container.NewBorder(nil, nil, title, s.Variants),
		widget.NewCard("", "nyan.ProgressBar", determinate),
		widget.NewCard("", "nyan.ProgressBarInfinite", indeterminate),
		widget.NewCard("", "Stock Fyne widgets for comparison", stock),
	)
	return s
}

// Start runs the live demo: the "Live" bar fills up over and over again.
func (s *Showcase) Start() {
	s.toggleLive()
	go func() {
		for range time.Tick(50 * time.Millisecond) {
			fyne.Do(func() {
				if !s.live {
					return
				}
				v := s.Live.Value + 0.005
				if v > 1 {
					v = 0
				}
				s.Live.SetValue(v)
			})
		}
	}()
}

func (s *Showcase) toggleLive() {
	s.live = !s.live
	if s.live {
		s.liveBtn.SetText("Pause")
		s.liveBtn.SetIcon(theme.MediaPauseIcon())
	} else {
		s.liveBtn.SetText("Start")
		s.liveBtn.SetIcon(theme.MediaPlayIcon())
	}
}

func label(text string) *widget.Label {
	return widget.NewLabel(text)
}

// Variant forces a theme variant regardless of the system setting.
type Variant struct {
	fyne.Theme
	Variant fyne.ThemeVariant
}

// Color returns the colour of the forced variant.
func (v Variant) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return v.Theme.Color(n, v.Variant)
}
