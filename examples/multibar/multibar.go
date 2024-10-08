// Copyright © 2022 Atonal Authors
//

package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/hedzr/progressbar"
	"github.com/hedzr/progressbar/cursor"
)

var whichStepper = 1

type TitledUrl string

func (t TitledUrl) String() string {
	return string(t)
}

func (t TitledUrl) Title() string {
	parse, err := url.Parse(string(t))
	if err != nil {
		return string(t)
	}
	return path.Base(parse.Path)
}

func doEachGroup2(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New(),
		progressbar.WithDownloadTaskOnStart(func(task *progressbar.DownloadTask, bar progressbar.PB) (err error) {
			if task.Req == nil {
				task.Req, err = http.NewRequest("GET", task.Url, nil) //nolint:gocritic
				if err != nil {
					log.Printf("Error: %v", err)
					return
				}
				task.File, err = os.OpenFile(task.Filename, os.O_CREATE|os.O_WRONLY, 0o644)
				if err != nil {
					log.Printf("Error: %v", err)
					return
				}
				task.Resp, err = http.DefaultClient.Do(task.Req)
				if err != nil {
					log.Printf("Error: %v", err)
					return
				}

				const BUFFERSIZE = 4096
				task.Buffer = make([]byte, BUFFERSIZE)

				bar.UpdateRange(0, task.Resp.ContentLength)

				task.Writer = io.MultiWriter(task.File, bar)

			}
			return
		}),
	)
	defer tasks.Close()

	for _, ver := range group {
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		// fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
		// fmt.Printf("adding %v (title: %v)\n", url1.String(), url1.Title())
		tasks.Add(url1.String(), url1,
			progressbar.WithBarStepper(whichStepper),
		)
	}

	// log.Printf("tasks.Wait() for group %v", group)
	tasks.Wait() // start waiting for all tasks completed gracefully
	// log.Printf("tasks.Wait() ends for group %v", group)
}

func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		// fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
		// fmt.Printf("adding %v (title: %v)\n", url1.String(), url1.Title())
		tasks.Add(url1.String(), url1,
			progressbar.WithBarStepper(whichStepper),
		)
	}

	log.Printf("tasks.Wait() for group %v", group)
	tasks.Wait() // start waiting for all tasks completed gracefully
	log.Printf("tasks.Wait() ends for group %v", group)
}

func downloadGroups() {
	for _, group := range [][]string{
		{"1.14.1", "1.15.1"},
		{"1.16.1", "1.17.1", "1.18.1"},
	} {
		doEachGroup2(group)
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
