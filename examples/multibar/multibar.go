// Copyright © 2022 Atonal Authors
//

package main

import (
	"fmt"
	"time"

	"github.com/hedzr/progressbar"
)

func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		fn := fmt.Sprintf("go%v.src.tar.gz", ver)

		tasks.Add(url, fn,
			progressbar.WithBarStepper(1),
		)
	}

	time.Sleep(5 * time.Millisecond)
	tasks.Wait()
}

func main() {
	// cursor.Hide()
	// defer cursor.Show()

	for _, group := range [][]string{
		{"1.14.2", "1.15.1"},
		{"1.16.1", "1.17.1", "1.18.3"},
	} {
		doEachGroup(group)
	}
}
