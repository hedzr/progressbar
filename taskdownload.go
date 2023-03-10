// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"errors"
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
//	func doEachGroup(group []string) {
//		tasks := progressbar.NewDownloadTasks(progressbar.New())
//		defer tasks.Close()
//
//		for _, ver := range group {
//			url := "https://dl.google.com/go/go" + ver + ".src.tar.gz"
//			fn := "go" + ver + ".src.tar.gz"
//			// url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
//			// fn := fmt.Sprintf("go%v.src.tar.gz", ver)
//
//			tasks.Add(url, fn,
//				progressbar.WithBarStepper(whichStepper),
//			)
//		}
//		tasks.Wait()
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
func NewDownloadTasks(bar MultiPB) *DownloadTasks {
	return &DownloadTasks{bar: bar}
}

type DownloadTasks struct {
	bar   MultiPB
	tasks []*aTask
	wg    sync.WaitGroup
}

func (s *DownloadTasks) Close() {
	s.bar.Close()
}

func (s *DownloadTasks) Add(url, filename string, opts ...Opt) {
	task := new(aTask)
	task.wg = &s.wg
	task.url = url
	task.fn = filename

	var o []Opt
	o = append(o,
		WithBarWorker(task.doWorker),
		WithBarOnCompleted(task.onCompleted),
		WithBarOnStart(task.onStart),
	)
	o = append(o, opts...)

	s.bar.Add(
		100,
		task.fn, // fmt.Sprintf("downloading %v", s.fn),
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

type aTask struct {
	url, fn string

	req  *http.Request
	resp *http.Response
	f    *os.File

	wg *sync.WaitGroup
	w  io.Writer

	buf []byte

	doneCount int32
}

func (s *aTask) Close() {
	if s.resp != nil {
		err := s.resp.Body.Close()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	if s.f != nil {
		err := s.f.Close()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	if s.wg != nil {
		s.terminateTrigger()
	}
}

func (s *aTask) Run() {
	// go s.run()
}

func (s *aTask) run() {
	// for{
	// 	select {
	//
	// 	}
	// }
}

func (s *aTask) onCompleted(bar PB) {
	s.terminateTrigger()
}

func (s *aTask) terminateTrigger() {
	if atomic.CompareAndSwapInt32(&s.doneCount, 0, 1) {
		wg := s.wg
		s.wg = nil
		wg.Done()
	}
}

func (s *aTask) onStart(bar PB) {
	if s.req == nil {
		var err error
		s.req, err = http.NewRequest("GET", s.url, nil) //nolint:gocritic
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		s.f, err = os.OpenFile(s.fn, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		s.resp, err = http.DefaultClient.Do(s.req)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		bar.UpdateRange(0, s.resp.ContentLength)

		s.w = io.MultiWriter(s.f, bar)

		const BUFFERSIZE = 4096
		s.buf = make([]byte, BUFFERSIZE)
	}
}

func (s *aTask) doWorker(bar PB, exitCh <-chan struct{}) {
	// _, _ = io.Copy(s.w, s.resp.Body)

	for {
		n, err := s.resp.Body.Read(s.buf)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("Error: %v", err)
			return
		}
		if n == 0 {
			break
		}

		if _, err = s.w.Write(s.buf[:n]); err != nil {
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
}
