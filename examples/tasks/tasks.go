// Copyright © 2022 Atonal Authors
//

package main

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hedzr/progressbar"
	"github.com/hedzr/progressbar/cursor"
)

const schema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent|printf "%6s"}} | <b><font color="green">{{.Title}}</font></b> {{.Append}}`

var whichSpinner = 0

func forAllSpinners() {
	tasks := progressbar.NewTasks(progressbar.New())
	defer tasks.Close()

	for i := whichSpinner; i < whichSpinner+5; i++ {
		tasks.Add(
			progressbar.WithTaskAddBarOptions(
				progressbar.WithBarUpperBound(100),
				// progressbar.WithBarSpinner(i),
				// progressbar.WithBarWidth(8),
				progressbar.WithBarStepper(0),
				progressbar.WithBarTextSchema(schema), // change the bar layout here
			),
			progressbar.WithTaskAddBarTitle("Task "+strconv.Itoa(i)), // fmt.Sprintf("Task %v", i)),
			progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
				for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
					ms := time.Duration(20 + rand.Intn(500)) //nolint:gosec //just a demo
					time.Sleep(time.Millisecond * ms)
					bar.Step(1)
				}
			}),
		)
	}

	time.Sleep(5 * time.Millisecond)
	tasks.Wait()
}

func main() {
	cursor.Hide()
	defer cursor.Show()

	if len(os.Args) > 1 {
		i, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil && i >= 0 && int(i) < progressbar.MaxSpinners() {
			whichSpinner = int(i)
		}
	}

	forAllSpinners()
}
