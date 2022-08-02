// Copyright © 2022 Atonal Authors
//

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hedzr/progressbar"
)

var whichSpinner = 22

func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		fn := fmt.Sprintf("go%v.src.tar.gz", ver)

		tasks.Add(url, fn,
			progressbar.WithBarSpinner(whichSpinner),
			progressbar.WithBarWidth(16),
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
