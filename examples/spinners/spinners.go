// Copyright © 2022 Atonal Authors
//

package main

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hedzr/progressbar"
	"github.com/hedzr/progressbar/cursor"
)

const schema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <b><font color="green">{{.Title}}</font></b> {{.Append}}`

var whichSpinner = 0
var count = 0

func forAllSpinners() {
	tasks := progressbar.NewTasks(progressbar.New())
	defer tasks.Close()

	max := count
	_, h, _ := terminal.GetSize(int(os.Stdout.Fd()))
	if max >= h {
		max = h
	}

	for i := whichSpinner; i < whichSpinner+max; i++ {
		tasks.Add(
			progressbar.WithTaskAddBarOptions(
				progressbar.WithBarSpinner(i),
				progressbar.WithBarUpperBound(100),
				progressbar.WithBarWidth(16),
				progressbar.WithBarTextSchema(schema),
			),
			progressbar.WithTaskAddBarTitle("Task "+strconv.Itoa(i)), // fmt.Sprintf("Task %v", i)),
			progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
				for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
					ms := time.Duration(20 + rand.Intn(1800)) //nolint:gosec //just a demo
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

	count = progressbar.MaxSpinners()

	if len(os.Args) > 1 {
		i, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil && i >= 0 && int(i) < progressbar.MaxSpinners() {
			whichSpinner, count = int(i), 1
		}
	}

	forAllSpinners()
}
