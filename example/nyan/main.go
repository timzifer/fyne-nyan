// Command nyan shows every state of the nyan progress bars.
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/timzifer/fyne-nyan/internal/showcase"
)

func main() {
	a := app.NewWithID("io.github.timzifer.fyne-nyan")
	w := a.NewWindow("fyne-nyan")

	s := showcase.New(a)
	s.Start()

	w.SetContent(container.NewVScroll(container.NewPadded(s.Content)))
	w.Resize(fyne.NewSize(720, 760))
	w.ShowAndRun()
}
