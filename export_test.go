package nyan

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// SetClock replaces the animation clock, the returned func restores it.
func SetClock(f func() time.Time) (restore func()) {
	old := now
	now = f
	return func() { now = old }
}

// Animate advances the animation of a nyan widget to the current clock.
func Animate(w fyne.Widget) {
	switch r := test.WidgetRenderer(w).(type) {
	case *progressRenderer:
		r.step()
	case *infiniteRenderer:
		r.step()
	}
}
