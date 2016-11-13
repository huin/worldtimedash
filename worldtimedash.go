package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	ui "github.com/gizak/termui"
	termbox "github.com/nsf/termbox-go"
)

var (
	tzFlag      = flag.String("tzlist", "", "Timezones to display, delimited by commas")
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
	timesWidget.Border = false
	// RuneCountInString is technically incorrect because it doesn't account for
	// combining characters, but it should be fine most of the time.
	timesWidget.Width = utf8.RuneCountInString(*timeFmtFlag) + 1
	timesWidget.Height = len(tzs)
	timesWidget.X = 0
	timesWidget.Y = 0

	resize := func() {
		if *tmuxFlag {
			tmuxCmd := exec.Command(
				"tmux", "resize",
				"-t", os.Getenv("TMUX_PANE"),
				"-y", strconv.Itoa(len(tzs)),
			)
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

	ui.Merge("timer", ui.NewTimerCh(1*time.Second))
	ui.Handle("/timer", func(ui.Event) { draw() })
	ui.Handle("/sys/wnd/resize", func(ui.Event) {
		resize()
	})
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Loop()
}
