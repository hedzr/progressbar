// Copyright © 2022 Atonal Authors
//

// Copyright © 2022 Atonal Authors
//

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hedzr/progressbar"
)

const schema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <b><font color="green">{{.Title}}</font></b> {{.Append}}`

var whichStepper = 0

func forAllSteppers() {
	tasks := progressbar.NewTasks(progressbar.New())
	defer tasks.Close()

	max := progressbar.MaxSteppers()
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
				// progressbar.WithBarTextSchema(schema),
			),
			progressbar.WithTaskAddBarTitle(fmt.Sprintf("Task %v", i)),
			progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
				for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
					ms := time.Duration(200 + rand.Intn(1800))
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
	// cursor.Hide()
	// defer cursor.Show()

	if len(os.Args) > 1 {
		i, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil && i >= 0 && int(i) < progressbar.MaxSteppers() {
			whichStepper = int(i)
		}
	}

	forAllSteppers()
}
