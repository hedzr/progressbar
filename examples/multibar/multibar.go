// Copyright © 2022 Atonal Authors
//

package main

import (
	"log"
	"os"
	"strconv"

	"github.com/hedzr/progressbar"
	"github.com/hedzr/progressbar/cursor"
)

var whichStepper = 1

func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url := "https://dl.google.com/go/go" + ver + ".src.tar.gz" // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
		tasks.Add(url, fn,
			progressbar.WithBarStepper(whichStepper),
		)
	}

	tasks.Wait() // start waiting for all tasks completed gracefully
}

func downloadGroups() {
	for _, group := range [][]string{
		{"1.14.2", "1.15.1"},
		{"1.16.1", "1.17.1", "1.18.3"},
	} {
		doEachGroup(group)
	}
}

func main() {
	cursor.Hide()
	defer cursor.Show()

	if s := os.Getenv("WHICH"); s != "" {
		var err error
		if whichStepper, err = strconv.Atoi(s); err != nil {
			log.Fatalf("Wrong environment variable WHICH found. It MUST BE a valid number. Cause: %v", err)
		}
	}
	if len(os.Args) > 1 {
		var err error
		if whichStepper, err = strconv.Atoi(os.Args[1]); err != nil {
			log.Fatalf("Wrong argument ('%v') found. It MUST BE a valid number. Cause: %v", os.Args[1], err)
		} else if whichStepper > progressbar.MaxSteppers() {
			log.Fatalf("Wrong argument ('%v') found. The Maxinium value is %v. Cause: %v", os.Args[1], progressbar.MaxSteppers(), err)
		}
	}

	downloadGroups()
}
