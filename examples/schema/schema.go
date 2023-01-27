// Copyright © 2022 Atonal Authors
//

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

// const schema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <b><font color="green">{{.Title}}</font></b> {{.Append}}`
const schema = `{{.Indent}}{{.Prepend}} <b><font color="light-yellow">{{.Title}}</font></b> <font color="blue">{{.Percent}}</font> [{{.Bar}}]{{.Append}}`

var whichStepper = 0
var count = 0

func forAllSteppers() {
	tasks := progressbar.NewTasks(progressbar.New())
	defer tasks.Close()

	max := count
	_, h, _ := terminal.GetSize(int(os.Stdout.Fd()))
	if max >= h {
		max = h
	}

	for i := whichStepper; i < whichStepper+max; i++ {
		tasks.Add(
			progressbar.WithTaskAddBarOptions(
				progressbar.WithBarStepper(i),
				progressbar.WithBarUpperBound(100),
				progressbar.WithBarWidth(32),
				progressbar.WithBarTextSchema(schema),
			),
			progressbar.WithTaskAddBarTitle("Task "+strconv.Itoa(i)), // fmt.Sprintf("Task %v", i)),
			progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
				for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
					ms := time.Duration(10 + rand.Intn(500)) //nolint:gosec //just a demo
					time.Sleep(time.Millisecond * ms)
					bar.Step(1)
				}
			}),
		)
	}

	tasks.Wait()
}

func main() {
	cursor.Hide()
	defer cursor.Show()

	count = progressbar.MaxSteppers()

	if len(os.Args) > 1 {
		i, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil && i >= 0 && int(i) < progressbar.MaxSteppers() {
			whichStepper, count = int(i), 1
		}
	}

	forAllSteppers()
}