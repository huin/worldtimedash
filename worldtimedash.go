package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui"
	termbox "github.com/nsf/termbox-go"
)

var (
	tzFlag      = flag.String("tzlist", "", "Timezones to display")
	timeFmtFlag = flag.String("timeFmt", "2006/01/02T03:04-0700 (MST)", "Time format")
	tmuxFlag    = flag.Bool("tmux", false, "Embedded in tmux, resize dashboard to height")
)

func main() {
	flag.Parse()
	var tzs []*time.Location
	if len(*tzFlag) == 0 {
		tzs = []*time.Location{time.Local}
	} else {
		tzNames := strings.Split(*tzFlag, ",")
		for _, tzn := range tzNames {
			tz, err := time.LoadLocation(tzn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't load timezone %q termui: %v\n", tzn, err)
				os.Exit(1)
			}
			tzs = append(tzs, tz)
		}
	}

	if err := ui.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't initialize termui:", err)
		os.Exit(1)
	}
	defer ui.Close()

	timesWidget := ui.NewList()
	timesWidget.Items = make([]string, len(tzs))
	timesWidget.HasBorder = false
	timesWidget.Width = 30
	timesWidget.Height = len(tzs)
	timesWidget.X = 0
	timesWidget.Y = 0

	resize := func() {
		if *tmuxFlag {
			tmuxCmd := exec.Command("tmux", "resize", "-y", strconv.Itoa(len(tzs)))
			tmuxCmd.Stdin = nil
			tmuxCmd.Stdout = nil
			tmuxCmd.Stderr = nil
			if err := tmuxCmd.Start(); err != nil {
				// Not much to do in error case. Maybe we should have an error display area?
				return
			}
			_ = tmuxCmd.Wait()
			// Termbox can lose track of what it needs to redraw, force refresh.
			termbox.Sync()
		}
	}
	resize()

	draw := func() {
		now := time.Now()
		for i, tz := range tzs {
			timesWidget.Items[i] = now.In(tz).Format(*timeFmtFlag)
		}

		ui.Render(timesWidget)
	}
	draw()

	evt := ui.EventCh()
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case e := <-evt:
			switch e.Type {
			case ui.EventKey:
				switch e.Ch {
				case 'q':
					return
				}
			case ui.EventResize:
				resize()
			}
		case _ = <-ticker:
			draw()
		}
	}
}
