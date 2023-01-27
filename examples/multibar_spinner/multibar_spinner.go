// Copyright © 2022 Atonal Authors
//

package main

import (
	"os"
	"strconv"

	"github.com/hedzr/progressbar"
	"github.com/hedzr/progressbar/cursor"
)

var whichSpinner = 22

func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url := "https://dl.google.com/go/go" + ver + ".src.tar.gz" // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
		tasks.Add(url, fn,
			progressbar.WithBarSpinner(whichSpinner),
			progressbar.WithBarWidth(16),
		)
	}

	// time.Sleep(5 * time.Millisecond)
	tasks.Wait() // start waiting for all tasks completed gracefully
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

	for _, group := range [][]string{
		{"1.14.2", "1.15.1"},
		{"1.16.1"},
	} {
		doEachGroup(group)
	}
}
