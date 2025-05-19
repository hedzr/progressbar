// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

// NewDownloadTasks is a wrapped NewTasks to simplify the http
// downloading task.
//
// Via this API you can easily download one and more http files.
//
//	type TitledUrl string
//
//	func (t TitledUrl) String() string { return string(t) }
//
//	func (t TitledUrl) Title() string {
//		if parse, err := url.Parse(string(t)); err != nil {
//			return string(t)
//		}
//		return path.Base(parse.Path)
//	}
//
//	func doEachGroup(group []string) {
//		tasks := progressbar.NewDownloadTasks(progressbar.New(),
//			// progressbar.WithTaskAddOnTaskCompleted(func...),
//		)
//		defer tasks.Close()
//
//		for _, ver := range group {
//			url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz")
//			tasks.Add(url1.String(), url1,
//				progressbar.WithBarStepper(whichStepper),
//			)
//		}
//
//		tasks.Wait() // start waiting for all tasks completed gracefully
//	}
//
//	func downloadGroups() {
//		for _, group := range [][]string{
//			{"1.14.2", "1.15.1"},
//			{"1.16.1", "1.17.1", "1.18.3"},
//		} {
//			doEachGroup(group)
//		}
//	}
func NewDownloadTasks(bar MultiPB, opts ...DownloadTasksOpt) *DownloadTasks {
	r := &DownloadTasks{bar: bar}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

type DownloadTasksOpt func(tsk *DownloadTasks)

func WithDownloadTaskOnStart(fn OnStartCB) DownloadTasksOpt {
	return func(tsk *DownloadTasks) {
		tsk.onStartCB = fn
	}
}

type DownloadTasks struct {
	bar       MultiPB
	tasks     []*DownloadTask
	wg        sync.WaitGroup
	onStartCB OnStartCB
}

func (s *DownloadTasks) Close() {
	s.bar.Close()
}

// Add a url as a downloading task, which will be started at
// background right now.
//
// The downloaded content is stored into local file, `filename`
// specified.
// The `filename` will be shown in progressbar as a title. You
// can customize its title with `interface{ Title() string`.
// A sample could be:
//
//	type TitledUrl string
//
//	func (t TitledUrl) String() string { return string(t) }
//
//	func (t TitledUrl) Title() string {
//		if parse, err := url.Parse(string(t)); err != nil {
//			return string(t)
//		}
//		return path.Base(parse.Path)
//	}
//
//	func doEachGroup(group []string) {
//		tasks := progressbar.NewDownloadTasks(progressbar.New(),
//			// progressbar.WithTaskAddOnTaskCompleted(func...),
//		)
//		defer tasks.Close()
//
//		for _, ver := range group {
//			url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz")
//			tasks.Add(url1.String(), url1,
//				progressbar.WithBarStepper(whichStepper),
//			)
//		}
//
//		tasks.Wait() // start waiting for all tasks completed gracefully
//	}
func (s *DownloadTasks) Add(url string, filename any, opts ...Opt) {
	task := new(DownloadTask)
	task.wg = &s.wg
	task.Url = url
	if s, ok := filename.(string); ok {
		task.Filename = s
	} else if sfn, ok := filename.(interface{ Title() string }); ok {
		task.Filename = sfn.Title()
	} else if sfn, ok := filename.(interface{ String() string }); ok {
		task.Filename = sfn.String()
	} else {
		task.Filename = fmt.Sprintf("%v", filename)
	}
	task.Title = task.Filename
	if sfn, ok := filename.(interface{ Title() string }); ok {
		task.Title = sfn.Title()
	}
	task.onStartCB = s.onStartCB

	var o []Opt
	o = append(o,
		WithBarWorker(task.doWorker),
		WithBarOnCompleted(task.onCompleted),
		WithBarOnStart(task.onStart),
	)
	o = append(o, opts...)

	s.bar.Add(
		100,
		task.Title, // fmt.Sprintf("downloading %v", s.fn),
		// // WithBarSpinner(14),
		// // WithBarStepper(3),
		// WithBarStepper(0),
		// WithBarWorker(s.doWorker),
		// WithBarOnCompleted(s.onCompleted),
		// WithBarOnStart(s.onStart),
		o...,
	)

	s.wg.Add(1)
}

func (s *DownloadTasks) Wait() {
	s.wg.Wait()
}

type DownloadTask struct {
	Url, Filename, Title string

	Req  *http.Request
	Resp *http.Response
	File *os.File

	Writer io.Writer
	Buffer []byte

	wg        *sync.WaitGroup
	doneCount int32
	onStartCB OnStartCB
}

type OnStartCB func(task *DownloadTask, bar PB) (err error)

func (s *DownloadTask) Close() {
	if s.Resp != nil {
		err := s.Resp.Body.Close()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	if s.File != nil {
		err := s.File.Close()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	if s.wg != nil {
		s.terminateTrigger()
	}
}

// func (s *DownloadTask) Run() {
// 	// go s.run()
// }
//
// func (s *DownloadTask) run() {
// 	// for{
// 	// 	select {
// 	//
// 	// 	}
// 	// }
// }

func (s *DownloadTask) Complete() {
	s.terminateTrigger()
}

func (s *DownloadTask) onCompleted(bar PB) {
	s.terminateTrigger()
}

func (s *DownloadTask) terminateTrigger() {
	if atomic.CompareAndSwapInt32(&s.doneCount, 0, 1) {
		wg := s.wg
		s.wg = nil
		wg.Done()
	}
}

func (s *DownloadTask) onStart(bar PB) {
	if s.Req == nil {
		var err error

		if s.onStartCB != nil {
			if err = s.onStartCB(s, bar); err != nil {
				log.Printf("Error: %v", err)
				return
			}
			return
		}

		s.Req, err = http.NewRequest("GET", s.Url, nil) //nolint:gocritic
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		s.File, err = os.OpenFile(s.Filename, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		s.Resp, err = http.DefaultClient.Do(s.Req)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}

		const BUFFERSIZE = 4096
		s.Buffer = make([]byte, BUFFERSIZE)

		bar.UpdateRange(0, s.Resp.ContentLength)

		s.Writer = io.MultiWriter(s.File, bar)
	}
}

func (s *DownloadTask) doWorker(bar PB, exitCh <-chan struct{}) (stop bool) {
	// _, _ = io.Copy(s.w, s.resp.Body)

	if s.Req == nil || s.File == nil {
		return
	}

	if s.Resp == nil {
		log.Printf("Warn: %v", "invalid http request or response (nil).")
		return
	}

	for {
		n, err := s.Resp.Body.Read(s.Buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("Error: %v", err)
			return
		}
		if n == 0 {
			break
		}

		if _, err = s.Writer.Write(s.Buffer[:n]); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		select {
		case <-exitCh:
			return
		default: // avoid block at <-exitCh
		}

		// time.Sleep(time.Millisecond * 100)
	}
	return
}
