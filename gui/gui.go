package gui

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jroimartin/gocui"
)

var BaseDir string
var MainGui *gocui.Gui

func scrollView(g *gocui.Gui, v *gocui.View, dy int) error {
	if v != nil {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		_, sy := v.Size()
		offset := (sy - 1) / 2
		l, _ := v.Line(cy + dy)
		if len(l) > 0 {
			if cy <= offset || (oy == 0 && dy < 0) {
				v.SetCursor(cx, cy+dy)
			} else {
				l, _ := v.Line(cy + dy + offset)
				if len(l) > 0 {
					v.SetOrigin(ox, oy+dy)
				} else {
					v.SetCursor(cx, cy+dy)
				}
			}
		}
	}

	// Print the content of the file
	go g.Execute(func(g *gocui.Gui) error {
		vc, _ := g.View("content")
		vc.Clear()
		vl, _ := g.View("list")
		_, cy := vl.Cursor()
		l, _ := vl.Line(cy)
		l = filepath.Clean(BaseDir + string(filepath.Separator) + l)
		if _, err := os.Stat(l); err == nil {
			b, err := ioutil.ReadFile(l)
			if err == nil {
				fmt.Fprintln(vc, string(b))
			}
		}
		return nil
	})
	return nil
}

func ListDir() {
	listDir(MainGui, BaseDir)
}

func listDir(g *gocui.Gui, dir string) {
	v, _ := g.View("list")
	v.Clear()
	dir, _ = filepath.Abs(filepath.Clean(dir))
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if !f.IsDir() {
			fmt.Fprintln(v, f.Name())
		}
	}
	scrollView(g, v, 0)
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("list", 0, 0, maxX/4-5, maxY-1); err != nil {
		v.Title = "Files"
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err := g.SetCurrentView("list"); err != nil {
			return err
		}
		v.Highlight = true
		v.FgColor = gocui.ColorCyan | gocui.AttrBold
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorYellow | gocui.AttrBold
		listDir(g, BaseDir)
		scrollView(g, v, 0)
	}
	if v, err := g.SetView("content", maxX/4-4, 0, maxX-1, maxY-12); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "Content"
	}
	if v, err := g.SetView("log", maxX/4-4, maxY-11, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "Log"
		v.Autoscroll = true
	}
	return nil
}

func GetGui(g *gocui.Gui, gui string) *gocui.View {
	v, e := g.View(gui)
	if e != nil {
		log.Panicln(e)
	}
	return v
}

func PrintTo(gui string, s string) error {
	v, e := MainGui.View(gui)
	if e != nil {
		return e
	} else {
		fmt.Fprintln(v, s)
		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func scrollUp(g *gocui.Gui, v *gocui.View) error {
	return scrollView(g, v, -1)
}

func scrollDown(g *gocui.Gui, v *gocui.View) error {
	return scrollView(g, v, 1)
}

func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", gocui.KeyArrowUp, gocui.ModNone, scrollUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", 'k', gocui.ModNone, scrollUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", gocui.KeyArrowDown, gocui.ModNone, scrollDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", 'j', gocui.ModNone, scrollDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", gocui.KeyHome, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", 'd', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		vl, _ := g.View("list")
		_, cy := vl.Cursor()
		l, _ := vl.Line(cy)
		l, _ = filepath.Abs(filepath.Clean(BaseDir + string(filepath.Separator) + l))
		if _, err := os.Stat(l); !os.IsNotExist(err) {
			if err := os.Remove(l); err != nil {
				log.Panicln(err)
			}
		}
		listDir(g, BaseDir)
		_, cy = vl.Cursor()
		l, _ = vl.Line(cy)
		if l, _ = vl.Line(cy); len(l) <= 0 {
			scrollView(g, v, -1)
		} else {
			scrollView(g, v, 0)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func SetGui(outputTo string) {
	MainGui, _ = gocui.NewGui(gocui.Output256)
	BaseDir, _ = filepath.Abs(filepath.Clean(outputTo))
	defer MainGui.Close()
	MainGui.SetManagerFunc(layout)
	MainGui.Mouse = false
	if err := initKeybindings(MainGui); err != nil {
		log.Fatalln(err)
	}
	if err := MainGui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}