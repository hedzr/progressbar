// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
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

func WithDownloadTaskLogger(logger *slog.Logger) DownloadTasksOpt {
	return func(tsk *DownloadTasks) {
		tsk.logger = logger
	}
}

type DownloadTasks struct {
	bar       MultiPB
	tasks     []*DownloadTask
	wg        sync.WaitGroup
	onStartCB OnStartCB
	logger    *slog.Logger
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
	if s.logger == nil {
		s.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	}

	task := new(DownloadTask)
	task.logger = s.logger
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

	logger *slog.Logger
}

type OnStartCB func(task *DownloadTask, bar PB) (err error)

func (s *DownloadTask) Close() {
	if s.Resp != nil {
		err := s.Resp.Body.Close()
		if err != nil {
			s.logger.Error("Close http response failure", "err", err)
		}
	}
	if s.File != nil {
		err := s.File.Close()
		if err != nil {
			s.logger.Error("Close file failure", "err", err)
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

func getFileSize(filepath string) (int64, error) {
	var fileSize int64
	fi, err := os.Stat(filepath)
	if err != nil {
		return fileSize, err
	}
	if fi.IsDir() {
		return fileSize, nil
	}
	fileSize = fi.Size()
	return fileSize, nil
}

func (s *DownloadTask) onStart(bar PB) {
	if s.Req == nil {
		var err error

		if s.onStartCB != nil {
			if err = s.onStartCB(s, bar); err != nil {
				s.logger.Error("user customized onStartCB returns failure state", "err", err)
			}
			return
		}

		var existingFileSize int64
		existingFileSize, _ = getFileSize(s.Filename)

		resumeable := bar.Resumeable() && existingFileSize > 0
		s.logger.Debug("resumeable state", "resumeable", bar.Resumeable(), "resume-point", existingFileSize)

		s.Req, err = http.NewRequest("GET", s.Url, nil) //nolint:gocritic
		if err != nil {
			s.logger.Error("creating a new http request failed", "err", err)
			return
		}
		if resumeable {
			s.Req.Header.Set("Range", fmt.Sprintf("bytes=%v-", existingFileSize))
			s.File, err = os.OpenFile(s.Filename, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				s.logger.Error("sending header for resumeable trunks failed", "err", err, "resume-point", existingFileSize)
				return
			}
			whence := io.SeekEnd
			_, err = s.File.Seek(0, whence)
			// fmt.Printf("size of %q: %d - resumeable enabled - seeked to end of file.\n", task.Filename, existingFileSize)
		} else {
			s.File, err = os.OpenFile(s.Filename, os.O_CREATE|os.O_WRONLY, 0o644)
		}
		if err != nil {
			s.logger.Error("opening/seeking on local file failed", "err", err)
			return
		}
		s.Resp, err = http.DefaultClient.Do(s.Req)
		// println(s.Resp.StatusCode)
		if s.Resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			const BUFFERSIZE = 4096
			s.Buffer = make([]byte, BUFFERSIZE)

			bar.UpdateRange(0, existingFileSize)
			bar.SetInitialValue(existingFileSize)
			s.Writer = bar

			// setup bar for resumeable downloader
			s.File.Close()
			s.File = nil // make redraw() safety
			s.Req = nil  // make redraw() safety
			s.Complete() // this task has been completed because all pieces were downloaded

			s.logger.Debug(fmt.Sprintf("size of %q: %d/%d - resumeable enabled - seeked to end of file.\n", s.Filename, existingFileSize, s.Resp.ContentLength))
			return
		}
		if err != nil {
			s.logger.Error("getting http response object failed", "err", err)
			return
		}

		const BUFFERSIZE = 4096
		s.Buffer = make([]byte, BUFFERSIZE)

		if s.Resp.StatusCode == http.StatusPartialContent {
			if resumeable && existingFileSize > 0 {
				bar.SetInitialValue(existingFileSize)
			}
			bar.UpdateRange(0, s.Resp.ContentLength+existingFileSize)
			s.logger.Debug(fmt.Sprintf("size of %q: %d/%d - resumeable enabled - seeked to end of file. PARTIAL\n", s.Filename, existingFileSize, s.Resp.ContentLength))
		} else {
			bar.UpdateRange(0, s.Resp.ContentLength)
		}

		s.Writer = io.MultiWriter(s.File, bar)
	}
}

func (s *DownloadTask) doWorker(bar PB, exitCh <-chan struct{}) (stop bool) {
	// _, _ = io.Copy(s.w, s.resp.Body)

	if s.Req == nil || s.File == nil {
		return
	}

	if s.Resp == nil {
		s.logger.Warn("invalid http request or response (nil).")
		return
	}

	for {
		n, err := s.Resp.Body.Read(s.Buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			s.logger.Error("reading from http response failed", "err", err)
			return
		}
		if n == 0 {
			break
		}

		if _, err = s.Writer.Write(s.Buffer[:n]); err != nil {
			s.logger.Error("writing trunk to local file failed", "err", err)
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
