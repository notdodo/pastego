package gui

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/edoz90/pastego/filesupport"
	"github.com/jroimartin/gocui"
)

var BaseDir string
var MainGui *gocui.Gui

// Move the cursor of 'list' view and show the content of the highlighted bin
func scrollView(g *gocui.Gui, v *gocui.View, dy int) error {
	if v != nil {
		_, cy := v.Cursor()
		vc, _ := g.View("content")
		vl, _ := g.View("list")
		l, _ := v.Line(cy + dy)
		if len(l) > 0 {
			g.Update(func(g *gocui.Gui) error {
				moveTo(dy, v)

				// Update the view
				vc.Clear()
				_, cy := vl.Cursor()
				l, _ := vl.Line(cy)
				l = filepath.Clean(BaseDir + string(filepath.Separator) + l)
				if _, err := os.Stat(l); err == nil {
					if b, err := ioutil.ReadFile(l); err == nil {
						PrintTo("content", string(b))
					}
				}
				return nil
			})
		}
	}
	return nil
}

// Jump forward/backward of a defined offset
func moveTo(step int, v *gocui.View) {
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	_, sy := v.Size()
	offset := (sy - 1) / 2
	// Start list
	if cy <= offset || (oy == 0 && step < 0) {
		v.SetCursor(cx, cy+step)
	} else {
		var l string
		var e error
		if step > 0 {
			// End list
			l, e = v.Line(sy)
		} else {
			// Middle list
			l, e = v.Line(cy + step + offset)
		}
		if e == nil && len(l) > 0 {
			v.SetOrigin(ox, oy+step)
		} else {
			v.SetCursor(cx, cy+step)
		}
	}
}

func ListDir() {
	listDir(MainGui, BaseDir)
}

//Get the list the of the saved files
func listDir(g *gocui.Gui, dir string) {
	g.Update(func(g *gocui.Gui) error {
		v, _ := g.View("list")
		v.Clear()
		dir, _ = filepath.Abs(filepath.Clean(dir))
		files, _ := ioutil.ReadDir(dir)
		for _, f := range files {
			if !f.IsDir() {
				PrintTo("list", f.Name())
			}
		}
		v.Title = "Files: " + strconv.Itoa(len(files)+1)
		scrollView(g, v, 0)
		return nil
	})
}

// Set up the layout
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	// List view: list of all bins founds and saved
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
	// Content view: display the selected file content
	if v, err := g.SetView("content", maxX/4-4, 0, maxX-1, maxY-12); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "Content"
	}
	// Log view: shows the actual progress of the bot
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

// Jump to next block of findings: jump from a letter to another
func jumpToNext(g *gocui.Gui, direction int) error {
	var dy int
	if direction > 0 {
		dy = 1
	} else {
		dy = -1
	}
	v, _ := g.View("list")
	for {
		_, cy := v.Cursor()
		l0, _ := v.Line(cy)
		l1, _ := v.Line(cy + dy)
		if len(l0) > 0 && len(l1) > 0 {
			startCharBefore := string([]rune(l0)[0])
			startCharAfter := string([]rune(l1)[0])
			if startCharBefore != startCharAfter {
				scrollView(g, v, dy)
				break
			}
			moveTo(dy, v)
		} else {
			scrollView(g, v, 0)
			break
		}
	}
	return nil
}

func jumpToNextDown(g *gocui.Gui, v *gocui.View) error {
	return jumpToNext(g, 1)
}

func jumpToNextUp(g *gocui.Gui, v *gocui.View) error {
	return jumpToNext(g, -1)
}

func initKeybindings(g *gocui.Gui) error {
	// quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	// quit
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}
	// move up
	if err := g.SetKeybinding("list", gocui.KeyArrowUp, gocui.ModNone, scrollUp); err != nil {
		return err
	}
	// move up
	if err := g.SetKeybinding("list", 'k', gocui.ModNone, scrollUp); err != nil {
		return err
	}
	// move down
	if err := g.SetKeybinding("list", gocui.KeyArrowDown, gocui.ModNone, scrollDown); err != nil {
		return err
	}
	// move down
	if err := g.SetKeybinding("list", 'j', gocui.ModNone, scrollDown); err != nil {
		return err
	}
	// jump forward
	if err := g.SetKeybinding("list", 'n', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return scrollView(g, v, 15)
	}); err != nil {
		return err
	}
	// jump to next letter
	if err := g.SetKeybinding("list", 'N', gocui.ModNone, jumpToNextDown); err != nil {
		return err
	}
	// jump backward
	if err := g.SetKeybinding("list", 'p', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return scrollView(g, v, -15)
	}); err != nil {
		return err
	}
	// jump to previous letter
	if err := g.SetKeybinding("list", 'P', gocui.ModNone, jumpToNextUp); err != nil {
		return err
	}
	// move to first element
	if err := g.SetKeybinding("list", gocui.KeyHome, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.Update(func(g *gocui.Gui) error {
			v.SetOrigin(0, 0)
			v.SetCursor(0, 0)
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	// delete an entry/file
	if err := g.SetKeybinding("list", 'd', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		vl, _ := g.View("list")
		_, cy := vl.Cursor()
		l, _ := vl.Line(cy)

		if err := filesupport.DeleteFile(l, BaseDir); err == nil {
			// Update cursor state
			g.Update(func(g *gocui.Gui) error {
				listDir(g, BaseDir)
				_, sy := v.Size()
				// Realign the view
				if l, e := vl.Line(sy); len(l) <= 0 || e != nil {
					ox, oy := v.Origin()
					v.SetOrigin(ox, oy-1)
				}
				return nil
			})
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
