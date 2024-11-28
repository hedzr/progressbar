// Copyright © 2022 Atonal Authors
//

package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

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

	log.Printf("tasks.Wait() for group %v", group)
	tasks.Wait() // start waiting for all tasks completed gracefully
	log.Printf("tasks.Wait() ends for group %v", group)
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

//

//

//

type Job struct {
	Url TitledUrl

	writer     io.Writer // writing to pbar to update scrollpos in the scrolling range.
	totalTicks int64     // up-bound of the scrolling range.
	written    int64

	mpb   progressbar.MultiPB
	index int
	wg    *sync.WaitGroup

	m sync.Mutex
}

func (j *Job) Start() {
	// some initial stuffs can be put here
}

func (j *Job) Update(delta int) int64 {
	j.m.Lock()
	defer j.m.Unlock()

	if j.written+int64(delta) >= j.totalTicks {
		delta = int(j.totalTicks - j.written)
		data := make([]byte, delta)
		j.written += int64(delta)
		_, _ = j.writer.Write(data)
		j.mpb.Redraw()
		return j.written
	}
	data := make([]byte, delta)
	j.written += int64(delta)
	_, _ = j.writer.Write(data)
	return j.written
}

// In a job, onStart is the initial point to get the progressbar.PB.
//
// To update the scrolling position, you shall write delta bytes
// into the bar (a io.Writer).
//
// The total bytes was initialized at progressbar.MultiPB.Add(...).
// You can update the total bytes (upper bound) with
// progressbar.PB.UpdateRange(...).
func (j *Job) onStart(bar progressbar.PB) {
	j.writer = bar
}
func (j *Job) doWorker(bar progressbar.PB, exitCh <-chan struct{}) {
	// step by step, do yours

	defer j.mpb.Redraw() // and redraw the bar

	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()
	defer j.wg.Done()

	for {
		select {
		case <-ticker.C: // or update pbar every 50ms
			if j.Update(100) >= j.totalTicks {
				return
			}
		}
	}
}
func (j *Job) onCompleted(bar progressbar.PB) {
	// trigger terminated
}

func doEachGroupWithTasks(group []string) {
	mpb := progressbar.New()

	var jobs []*Job
	var wg sync.WaitGroup

	for _, ver := range group {
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		job := &Job{Url: url1, mpb: mpb, wg: &wg}

		job.totalTicks = int64(3500 + int(rand.Int31n(2000))) // 3500ms
		job.index = mpb.Add(
			job.totalTicks,
			url1.Title(),
			progressbar.WithBarOnStart(job.onStart),
			progressbar.WithBarWorker(job.doWorker),
			progressbar.WithBarOnCompleted(job.onCompleted),
			progressbar.WithBarStepper(whichStepper),
		)

		wg.Add(1)
		jobs = append(jobs, job)
		job.Start()
	}

	wg.Wait() // waiting for all tasks done.

	mpb.Close() // cleanup
}

func downloadGroups() {
	for _, group := range [][]string{
		{"1.14.1", "1.15.1"},
		{"1.16.1", "1.17.1", "1.18.1"},
	} {
		doEachGroupWithTasks(group)
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
