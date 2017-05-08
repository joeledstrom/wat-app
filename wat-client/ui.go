package main

import (
	"log"
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)


const (
	inputView = "inputView"
	messageView = "messageView"
)

func RunUi(sendChannel chan<- string, recvChannel <-chan string) {
	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	g.Cursor = true
	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}


	g.SetKeybinding(inputView, gocui.KeyEnter, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {

			sendChannel <- v.Buffer()

			v.Clear()
			v.SetCursor(0,0)
			v.SetOrigin(0,0)

			return nil
		})


	go func() {
		for {
			line := <-recvChannel
			g.Execute(func (g *gocui.Gui) error {
				v, _ := g.View(messageView)
				fmt.Fprintf(v, "%s\n", strings.TrimSpace(line))
				return nil
			})
		}

	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func layout(g *gocui.Gui) error {

	maxX, maxY := g.Size()

	if v, err := g.SetView(inputView, -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true
		v.Wrap = false
		v.Autoscroll = true
		g.SetCurrentView(inputView)

	}

	if v, err := g.SetView(messageView, 0, 0, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Wrap = true
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}




