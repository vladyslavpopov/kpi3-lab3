package ui

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w         screen.Window
	tx        chan screen.Texture
	done      chan struct{}
	sz        size.Event
	figurePos image.Point
}

func (pw *Visualizer) Main() {
	pw.tx = make(chan screen.Texture)
	pw.done = make(chan struct{})
	pw.figurePos = image.Point{X: 400, Y: 400}
	driver.Main(pw.run)
}

func (pw *Visualizer) Update(t screen.Texture) {
	pw.tx <- t
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  800,
		Height: 800,
	})
	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	pw.w = w

	events := make(chan any)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("new event: %v", e)
			}
			if detectTerminate(e) {
				close(events)
				break
			}
			events <- e
		}
	}()

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case t = <-pw.tx:
			w.Send(paint.Event{})
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true
		}
	case key.Event:
		if e.Code == key.CodeEscape {
			return true
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {
	case size.Event:
		pw.sz = e

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		if e.Button == mouse.ButtonRight && e.Direction == mouse.DirPress {
			pw.figurePos = image.Point{X: int(e.X), Y: int(e.Y)}
			pw.w.Send(paint.Event{})
		}

	case paint.Event:
		if t == nil {
			pw.drawDefaultUI()
		} else {
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

func (pw *Visualizer) drawDefaultUI() {
	pw.w.Fill(pw.sz.Bounds(), color.RGBA{G: 0xff, A: 0xff}, draw.Src)

	size := 200
	vertical := image.Rect(
		pw.figurePos.X-50, pw.figurePos.Y-size,
		pw.figurePos.X+50, pw.figurePos.Y+size,
	)
	horizontal := image.Rect(
		pw.figurePos.X-size, pw.figurePos.Y-50,
		pw.figurePos.X+size, pw.figurePos.Y+50,
	)
	pw.w.Fill(vertical, color.RGBA{R: 0xff, G: 0xff, A: 0xff}, draw.Src)
	pw.w.Fill(horizontal, color.RGBA{R: 0xff, G: 0xff, A: 0xff}, draw.Src)
}
