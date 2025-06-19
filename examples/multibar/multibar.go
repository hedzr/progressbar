// Copyright © 2022 Atonal Authors
//

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/hedzr/is/term/color"
	"github.com/hedzr/progressbar/v2"
)

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

func onStartCB() progressbar.OnStartCB {
	resumeable := *resumePtr
	return func(task *progressbar.DownloadTask, bar progressbar.MiniResizeableBar) (err error) {
		if task.Req == nil {
			var existingFileSize int64
			existingFileSize, _ = getFileSize(task.Filename)
			// if existingFileSize > 0 && resumeable {
			// 	fmt.Printf("size of %q: %d - resumeable enabled.\n", task.Filename, existingFileSize)
			// }
			logger.Debug("resumeable state", "resumeable", bar.Resumeable(), "resume-point", existingFileSize)

			task.Req, err = http.NewRequest("GET", task.Url, nil) //nolint:gocritic
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			if resumeable && existingFileSize > 0 {
				task.Req.Header.Set("Range", fmt.Sprintf("bytes=%v-", existingFileSize))
				task.File, err = os.OpenFile(task.Filename, os.O_APPEND|os.O_WRONLY, 0o644)
				if err != nil {
					log.Printf("Error: %v", err)
					return
				}
				whence := io.SeekEnd
				_, err = task.File.Seek(0, whence)
				// fmt.Printf("size of %q: %d - resumeable enabled - seeked to end of file.\n", task.Filename, existingFileSize)
			} else {
				task.File, err = os.OpenFile(task.Filename, os.O_CREATE|os.O_WRONLY, 0o644)
			}
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			task.Resp, err = http.DefaultClient.Do(task.Req)
			// println(task.Resp.StatusCode)
			if task.Resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
				const BUFFERSIZE = 4096
				task.Buffer = make([]byte, BUFFERSIZE)

				bar.UpdateRange(0, existingFileSize)
				bar.SetInitialValue(existingFileSize)
				task.File.Close()
				task.File = nil
				task.Req = nil
				task.Writer = bar
				task.Complete()

				// fmt.Printf("size of %q: %d/%d - resumeable enabled - seeked to end of file.\n", task.Filename, existingFileSize, task.Resp.ContentLength)

				return nil
			}
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}

			const BUFFERSIZE = 4096
			task.Buffer = make([]byte, BUFFERSIZE)

			if task.Resp.StatusCode == http.StatusPartialContent {
				if resumeable && existingFileSize > 0 {
					bar.SetInitialValue(existingFileSize)
				}
				bar.UpdateRange(0, task.Resp.ContentLength+existingFileSize)
				slog.Debug(fmt.Sprintf("size of %q: %d/%d - resumeable enabled - seeked to end of file. PARTIAL\n", task.Filename, existingFileSize, task.Resp.ContentLength))
			} else {
				bar.UpdateRange(0, task.Resp.ContentLength)
			}

			task.Writer = io.MultiWriter(task.File, bar)
		}
		return
	}
}

func doEachGroup2(group []string) {
	cb := onStartCB()
	if *defaultOnStartCBPtr {
		cb = nil
		logger.Debug("disable the client-side onStartCB callback func")
	}
	tasks := progressbar.NewDownloadTasks(progressbar.New(),
		progressbar.WithDownloadTaskOnStart(cb),
		progressbar.WithDownloadTaskLogger(logger),
	)
	defer tasks.Close()

	for _, ver := range group {
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		// fmt.Printf("adding %v (title: %v)\n", url1.String(), url1.Title())
		tasks.Add(url1.String(), url1,
			progressbar.WithBarStepper(whichStepper),
			progressbar.WithBarResumeable(*resumePtr),
			// progressbar.WithBarLogger(logger),
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

var (
	percentPtr *int
	resumePtr  *bool
	whichPtr   *int
	algorPtr   *int

	defaultOnStartCBPtr *bool

	whichStepper = 1
	algor        int

	logger *slog.Logger
)

func init() {
	percentPtr = flag.Int("stopat", 0, "the percent which task should puase it at")
	resumePtr = flag.Bool("resume", false, "continue the uncompleted task")
	whichPtr = flag.Int("which", whichStepper, fmt.Sprintf("choose a stepper (0..%d)", progressbar.MaxSteppers()))
	algorPtr = flag.Int("algor", algor, "select a algor (0..2)")
	defaultOnStartCBPtr = flag.Bool("default", false, "use internal resumeable side instead of your code at client-side")
}

func main() {
	color.Hide()
	defer color.Show()

	flag.Parse()

	if *defaultOnStartCBPtr {
		lvl := new(slog.LevelVar)
		lvl.Set(slog.LevelInfo)
		if *resumePtr {
			lvl.Set(slog.LevelDebug)
		}
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: lvl,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	}

	if s := os.Getenv("WHICH"); s != "" {
		var err error
		if whichStepper, err = strconv.Atoi(s); err != nil {
			log.Fatalf("Wrong environment variable WHICH found. It MUST BE a valid number. Cause: %v", err)
		}
	}
	args := flag.Args()
	if len(args) > 1 {
		var err error
		if whichStepper, err = strconv.Atoi(args[1]); err != nil {
			log.Fatalf("Wrong argument ('%v') found. It MUST BE a valid number. Cause: %v", args[1], err)
		} else if whichStepper > progressbar.MaxSteppers() {
			log.Fatalf("Wrong argument ('%v') found. The Maxinium value is %v. Cause: %v", args[1], progressbar.MaxSteppers(), err)
		}
	}

	downloadGroups()
}
