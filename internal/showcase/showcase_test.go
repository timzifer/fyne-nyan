package showcase

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestBinding(t *testing.T) {
	s := New(test.NewTempApp(t))
	w := test.NewWindow(s.Content)
	defer w.Close()

	if s.Slider.Value != 0.6 || s.Bound.Value != 0.6 {
		t.Fatalf("initial: slider %v, bar %v, want 0.6", s.Slider.Value, s.Bound.Value)
	}
	s.Slider.SetValue(0.3)
	if v, _ := s.Data.Get(); v != 0.3 || s.Bound.Value != 0.3 {
		t.Errorf("after slider change: data %v, bar %v, want 0.3", v, s.Bound.Value)
	}
}
