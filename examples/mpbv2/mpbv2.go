// Copyright © 2022 Atonal Authors
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

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

// func doEachGroup2(group []string) {
// 	tasks := progressbar.NewDownloadTasks(
// 		progressbar.New(),
// 		progressbar.WithDownloadTaskOnStart(func(task *progressbar.DownloadTask, bar progressbar.PB) (err error) {
// 			if task.Req == nil {
// 				task.Req, err = http.NewRequest("GET", task.Url, nil) //nolint:gocritic
// 				if err != nil {
// 					log.Printf("Error: %v", err)
// 					return
// 				}
// 				task.File, err = os.OpenFile(task.Filename, os.O_CREATE|os.O_WRONLY, 0o644)
// 				if err != nil {
// 					log.Printf("Error: %v", err)
// 					return
// 				}
// 				task.Resp, err = http.DefaultClient.Do(task.Req)
// 				if err != nil {
// 					log.Printf("Error: %v", err)
// 					return
// 				}

// 				const BUFFERSIZE = 4096
// 				task.Buffer = make([]byte, BUFFERSIZE)

// 				bar.UpdateRange(0, task.Resp.ContentLength)

// 				task.Writer = io.MultiWriter(task.File, bar)

// 			}
// 			return
// 		}),
// 	)
// 	defer tasks.Close()

// 	for _, ver := range group {
// 		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
// 		// fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
// 		// fmt.Printf("adding %v (title: %v)\n", url1.String(), url1.Title())
// 		tasks.Add(url1.String(), url1,
// 			progressbar.WithBarStepper(whichStepper),
// 		)
// 	}

// 	log.Printf("tasks.Wait() for group %v", group)
// 	tasks.Wait() // start waiting for all tasks completed gracefully
// 	log.Printf("tasks.Wait() ends for group %v", group)
// }

// func doEachGroup(group []string) {
// 	tasks := progressbar.NewDownloadTasks(progressbar.New())
// 	defer tasks.Close()

// 	for _, ver := range group {
// 		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
// 		// fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
// 		// fmt.Printf("adding %v (title: %v)\n", url1.String(), url1.Title())
// 		tasks.Add(url1.String(), url1,
// 			progressbar.WithBarStepper(whichStepper),
// 		)
// 	}

// 	log.Printf("tasks.Wait() for group %v", group)
// 	tasks.Wait() // start waiting for all tasks completed gracefully
// 	log.Printf("tasks.Wait() ends for group %v", group)
// }

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
	} else {
		data := make([]byte, delta)
		j.written += int64(delta)
		_, _ = j.writer.Write(data)
	}
	time.Sleep(time.Millisecond * 20)
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
func (j *Job) onStart(bar progressbar.MiniResizeableBar) {
	j.writer = bar
}

func (j *Job) doWorker(bar progressbar.MiniResizeableBar, exitCh <-chan struct{}) (stop bool) {
	// step by step, do yours

	defer j.mpb.Redraw() // and redraw the bar

	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()
	defer j.wg.Done()

stopped:
	for {
		select {
		case <-exitCh:
			break stopped
		case <-ticker.C: // or update pbar every 50ms
			if j.Update(100) >= j.totalTicks {
				break stopped
			}
			if got := j.mpb.PercentI(0); *percentPtr > 0 && got > *percentPtr {
				stop = true
				break stopped
			}
		}
	}
	return
}
func (j *Job) onCompleted(bar progressbar.MiniResizeableBar) {
	// trigger terminated
}

type groupedJobs struct {
	group []string
	title string
}

func (s groupedJobs) Title() string { return s.title }

func doEachGroupWithTasks(mpb progressbar.GroupedPB, wg *sync.WaitGroup, group groupedJobs) {
	if mpb == nil {
		mpb := progressbar.New().(progressbar.GroupedPB)
		defer mpb.Close() // cleanup
	}
	if wg == nil {
		wg = &sync.WaitGroup{}
		defer wg.Wait() // waiting for all tasks done.
	}

	var jobs []*Job

	for _, ver := range group.group {
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		job := &Job{Url: url1, mpb: mpb, wg: wg}

		job.totalTicks = int64(3500 + int(rand.Int31n(2000))) // 3500ms
		job.index = mpb.AddToGroup(
			group.Title(),
			job.totalTicks,
			url1.Title(),
			progressbar.WithBarResumeable(*resumePtr),
			progressbar.WithBarInitialValue(0), // no sense, just a placeholder, comment it safely
			progressbar.WithBarOnStart(job.onStart),
			progressbar.WithBarWorker(job.doWorker),
			progressbar.WithBarOnCompleted(job.onCompleted),
			progressbar.WithBarStepper(whichStepper),
			progressbar.WithBarStepperPostInit(func(bar progressbar.BarT) {
				bar.SetHighlightColor(color.FgDarkGray)

				// no sense but it could be a sapmle for demostrating how
				// to modify a bar at post-initial time.
				if *resumePtr {
					bar.SetInitialValue(0)
				}
			}),
		)

		wg.Add(1)
		jobs = append(jobs, job)
		job.Start()
	}
}

func downloadGroups1Worked() {
	// mpb := progressbar.NewGPB(progressbar.WithOnDone(func(mpb progressbar.MultiPB) {
	// 	println()
	// 	println("DONE.")
	// }))
	mpb := progressbar.NewGPB()
	defer mpb.Close() // cleanup

	var wg sync.WaitGroup
	defer wg.Wait() // waiting for all tasks done.

	for _, group := range []groupedJobs{
		{[]string{"1.14.1", "1.15.1"}, "AAA"},
		{[]string{"1.16.1", "1.17.1", "1.18.1"}, "BBB"},
	} {
		doEachGroupWithTasks(mpb, &wg, group)
	}

	for _, group := range []groupedJobs{
		{[]string{"1.20.1", "1.21.1"}, "CCC"},
		{[]string{"1.22.1"}, "DDD"},
	} {
		doEachGroupWithTasks(mpb, &wg, group)
	}

	// mpb.Close() // cleanup
}

func downloadGroups2Worked() {
	mpb := progressbar.New().(progressbar.GroupedPB)

	for _, group := range []groupedJobs{
		{[]string{"1.14.1", "1.15.1"}, "AAA"},
		{[]string{"1.16.1", "1.17.1", "1.18.1"}, "BBB"},
	} {
		doEachGroupWithTasks(mpb, nil, group)
	}

	mpb.Close() // cleanup
	mpb = nil

	mpb = progressbar.New().(progressbar.GroupedPB)

	for _, group := range []groupedJobs{
		{[]string{"1.20.1", "1.21.1"}, "CCC"},
		{[]string{"1.22.1"}, "DDD"},
	} {
		doEachGroupWithTasks(mpb, nil, group)
	}

	mpb.Close() // cleanup
}

func downloadGroups3Worked() {
	for _, group := range []groupedJobs{
		{[]string{"1.14.1", "1.15.1"}, "AAA"},
		{[]string{"1.16.1", "1.17.1", "1.18.1"}, "BBB"},
	} {
		doEachGroupWithTasks(nil, nil, group)
	}

	for _, group := range []groupedJobs{
		{[]string{"1.20.1", "1.21.1"}, "CCC"},
		{[]string{"1.22.1"}, "DDD"},
	} {
		doEachGroupWithTasks(nil, nil, group)
	}
}

func downloadGroupsV2Worked() {
	// const mySchema = `{{.Indent}}{{.Prepend}} <font color="green">{{.Title}}</font> {{.Percent}} {{.Bar}} {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}`
	// var versions = []string{"1.16.1", "1.17.1", "1.18.1", "1.19.1", "1.20.1", "1.21.1", "1.22.1", "1.23.1", "1.24.1"}
	var versions = []string{"1.24.1"}

	var mpb *progressbar.MPBV2
	if schema := os.Getenv("SCHEMA"); schema != "" {
		mpb = progressbar.NewV2(progressbar.WithSchema(schema))
	} else {
		mpb = progressbar.NewV2()
	}
	defer mpb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// define a counter job here
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	counterJob := func(bar *progressbar.MPBV2, grp *progressbar.GroupV2, tsk *progressbar.TaskBar, progress int64, args ...any) (delta int64, err error) {
		time.Sleep(time.Duration(rng.Intn(60)+30) * time.Millisecond)
		delta += int64(rng.Intn(5) + 1)
		return
	}

	// define the downloading job adder here
	verIdx := 0
	addDownloadJob := func(bar *progressbar.MPBV2, i, j int) {
		ver := versions[verIdx]
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		bar.AddDownloadingBar(
			"Group "+strconv.Itoa(i), "Task #"+strconv.Itoa(j),
			&progressbar.DownloadTask{
				Url:      url1.String(),
				Filename: url1.Title(),
				Title:    url1.Title(),
			},
			// more opts can be set here
			// progressbar.WithTaskBarStepper(whichStepper),
			// progressbar.WithTaskBarSpinner(whichSpinner),
		)
		verIdx++
	}

	total, num, numTasks := int64(100), 2, 3
	// we would add some progressing task groups,
	for i := range num {
		// in a single task group, we add some tasks,
		for j := range numTasks {
			// one of which is a downloading task.
			if (j == numTasks-1 || i == 0) && verIdx < len(versions) {
				addDownloadJob(mpb, i, j)
				continue
			}
			// and the rests are counter tasks.
			mpb.AddBar("Group "+strconv.Itoa(i), "Task #"+strconv.Itoa(j), 0, total, counterJob)
		}
	}

	// var wg sync.WaitGroup
	// wg.Add(num * numTasks)

	// so you will get a multi-group multi-task progress bar by Run it.
	mpb.Run(ctx)
}

var (
	percentPtr *int
	resumePtr  *bool
	whichPtr   *int
	algorPtr   *int

	defaultOnStartCBPtr *bool

	whichStepper = 0
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

	if s := os.Getenv("ALGOR"); s != "" {
		var err error
		if algor, err = strconv.Atoi(s); err != nil {
			log.Fatalf("Wrong environment variable ALGOR found. It MUST BE 0 or 1-3. Cause: %v", err)
		}
	}

	logger.Info("using algor", "algor", algor, "whichStepper", whichStepper, "resume", *resumePtr, "stopat", *percentPtr)
	switch algor {
	case 3:
		downloadGroups3Worked()
	case 2:
		downloadGroups2Worked()
	case 1:
		downloadGroups1Worked()
	default:
		downloadGroupsV2Worked()
	}
}
