// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

func NewDownloadTasks(bar MultiPB) *DownloadTasks {
	return &DownloadTasks{bar: bar}
}

type DownloadTasks struct {
	wg    sync.WaitGroup
	bar   MultiPB
	tasks []*aTask
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
		s.wg.Done()
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
	wg := s.wg
	s.wg = nil
	wg.Done()
	atomic.AddInt32(&s.doneCount, 1)
}

func (s *aTask) onStart(bar PB) {
	if s.req == nil {
		var err error
		s.req, err = http.NewRequest("GET", s.url, nil)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		s.f, err = os.OpenFile(s.fn, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		s.resp, err = http.DefaultClient.Do(s.req)
		if err != nil {
			log.Printf("Error: %v", err)
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
		if err != nil && err != io.EOF {
			log.Printf("Error: %v", err)
			return
		}
		if n == 0 {
			break
		}

		select {
		case <-exitCh:
			return
		default:
		}

		if _, err = s.w.Write(s.buf[:n]); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		// time.Sleep(time.Millisecond * 100)
	}
}
