package progressbar

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hedzr/is/term/color"
)

func NewV2() *MPBV2 {
	s := &MPBV2{
		chPaint: make(chan struct{}, 8192),
		logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})),
	}
	return s
}

type MPBV2 struct {
	startIdx   int // for repaint groups
	closed     int32
	chPaint    chan struct{}
	muPainting sync.RWMutex

	logger *slog.Logger
	groups []*GroupV2

	schema string
}

type GroupV2 struct {
	Name         string
	tasks        []*TaskBar
	done         int32
	titlePainted int32
	muTasks      sync.RWMutex
	block        color.RowsBlock
	wg           sync.WaitGroup
	dad          Repaintable // pointed to *MPBV2
}

type TaskBar struct {
	Name string

	stopTime  time.Time
	startTime time.Time

	min, max   int64
	progress   int64
	job        Job
	downloader *DownloadTask
	running    int32

	dad            Repaintable // pointed to *MPBV2
	stepper        BarT        // stepper or spinner here
	onDataPrepared OnDataPrepared
}

type Job func(bar *MPBV2, grp *GroupV2, tsk *TaskBar, progress int64, args ...any) (delta int64, err error)

type Writer interface {
	io.Writer
	io.StringWriter
}

type Repaintable interface {
	Repaint()
}

type Logger interface {
	Logger() *slog.Logger
}

type TaskBarOpt func(*TaskBar)

var _ Logger = (*MPBV2)(nil)
var _ Repaintable = (*MPBV2)(nil)
var _ MiniResizeableBar = (*TaskBar)(nil)

//
// ------------------------------ TASK OPTS
//

func WithTaskBarStepper(stepperIndex int, opts ...StepperOpt) TaskBarOpt {
	return func(tb *TaskBar) {
		if s, ok := steppers[stepperIndex]; ok {
			tb.stepper = s.init(opts...)
		}
	}
}

func WithTaskBarSpinner(spinnersIndex int, opts ...StepperOpt) TaskBarOpt {
	return func(tb *TaskBar) {
		if s, ok := spinners[spinnersIndex]; ok {
			tb.stepper = s.init(opts...)
		}
	}
}

//
// ------------------------------ MPBV2
//

func (s *MPBV2) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.chPaint)
		s.chPaint = nil
	}
}

func (s *MPBV2) AddDownloadingBar(group, task string, d *DownloadTask, opts ...TaskBarOpt) (err error) {
	s.muPainting.Lock()
	defer s.muPainting.Unlock()

	var grp *GroupV2
	if grp, err = s.findGroup(group); err != nil {
		grp = &GroupV2{Name: group, dad: s}
		grp.block = color.NewRowsBlock()
		s.groups = append(s.groups, grp)
		err = nil
	}
	err = grp.AddDownloader(s, task, d, opts...)
	return
}

func (s *MPBV2) AddBar(group, task string, min, max int64, job Job, opts ...TaskBarOpt) (err error) {
	s.muPainting.Lock()
	defer s.muPainting.Unlock()

	var grp *GroupV2
	if grp, err = s.findGroup(group); err != nil {
		grp = &GroupV2{Name: group, dad: s}
		grp.block = color.NewRowsBlock()
		s.groups = append(s.groups, grp)
		err = nil
	}
	err = grp.AddTask(s, task, min, max, job, opts...)
	return
}

func (s *MPBV2) findGroup(group string) (grp *GroupV2, err error) {
	for _, grp = range s.groups {
		if grp.Name == group {
			return
		}
	}
	return nil, errNotFound
}

func (s *MPBV2) Logger() *slog.Logger { return s.logger }

func (s *MPBV2) GroupByIndex(index int) *GroupV2 {
	s.muPainting.RLock()
	defer s.muPainting.RUnlock()
	if index >= 0 && index < len(s.groups) {
		return s.groups[index]
	}
	return nil
}

func (s *MPBV2) GroupByName(group string) *GroupV2 {
	s.muPainting.RLock()
	defer s.muPainting.RUnlock()
	for _, grp := range s.groups {
		if grp.Name == group {
			return grp
		}
	}
	return nil
}

// func (s *MPBV2) Wait() {
// 	//
// }

func (s *MPBV2) Run(ctx context.Context) {
	exitCh := make(chan struct{}, 8)
	pc := newPaintCtx(s)

	color.Hide()

	defer func() {
		pc.full = true
		s.repaint(pc)
		s.stop(ctx, pc)
		close(exitCh)
		color.Show()
	}()

	// collect downloading tasks and initialize them
	downloaders := s.start(ctx, pc)

	var gi int
	var grp *GroupV2
	for {
		select {
		case <-ctx.Done():
			exitCh <- struct{}{}
			return
		case <-s.chPaint:
			s.repaint(pc)
		default:
			if grp, gi = s.chooseGroup(gi); grp != nil {
				if downloaders != nil {
					// pick up the tasks in current group
					// and run these tasks.
					// each tash is a one-time go routine,
					// so running it once is enough.
					// to shutdown it gracefully, send sth
					// to exitCh.
					if tsks, ok := downloaders[grp]; ok {
						for _, tsk := range tsks {
							if _, _, done := tsk.Done(); !done {
								if atomic.CompareAndSwapInt32(&tsk.running, 0, 1) {
									go func() {
										stop := tsk.downloader.doWorker(tsk, exitCh)
										if _, _, done = tsk.Done(); done || stop {
											atomic.AddInt32(&grp.done, 1)
										}
									}()
								}
							}
						}
						_ = grp
					}
				}
				if allDone := grp.runJobs(ctx, s); allDone {
					pc.full = true
					// try cleanup stacked signals in chPaint
					var ignored = true
					for ignored {
						emptyIt(s.chPaint)
						ignored = grp.repaint(pc)
					}
					grp.block.Bottom()
				}
			} else {
				return
			}
		}
	}
}

func (s *MPBV2) stop(ctx context.Context, pc *paintCtx) {
	s.muPainting.RLock()
	defer s.muPainting.RUnlock()

	for _, grp := range s.groups {
		grp.muTasks.Lock()
		for _, tsk := range grp.tasks {
			if tsk.downloader != nil {
				tsk.downloader.onCompleted(tsk)
			}
		}
		grp.muTasks.Unlock()
	}
	_, _ = ctx, pc
}

func (s *MPBV2) start(ctx context.Context, pc *paintCtx) (downloaders map[*GroupV2][]*TaskBar) {
	var m = make(map[*GroupV2][]*TaskBar)
	for _, grp := range s.groups {
		grp.muTasks.Lock()
		for _, tsk := range grp.tasks {
			if tsk.downloader != nil {
				m[grp] = append(m[grp], tsk)
				tsk.downloader.onStart(tsk)
				if _, _, done := tsk.Done(); done {
					atomic.AddInt32(&grp.done, 1)
				}
			}

			// make taskbar.stepper safety
			if tsk.stepper == nil {
				WithTaskBarStepper(0)(tsk)
				tsk.startNow()
			}
		}
		grp.muTasks.Unlock()
	}
	if len(m) > 0 {
		downloaders = m
	}
	_, _ = ctx, pc
	return
}

func (s *MPBV2) chooseGroup(gi int) (grp *GroupV2, idx int) {
	s.muPainting.RLock()
	defer s.muPainting.RUnlock()

	if gi >= len(s.groups) {
		gi -= len(s.groups)
	}
	grp, idx = s.groups[gi], gi
	if grp.AllDone() {
		gi++
		if gi < len(s.groups) {
			return s.chooseGroup(gi)
		}
		grp = nil
	}
	return
}

func (s *MPBV2) Repaint() {
	s.chPaint <- struct{}{}
}

// func (s *MPBV2) RepaintNow() {
// 	s.repaint()
// }

func (s *MPBV2) repaint(pc *paintCtx) {
	if s.muPainting.TryRLock() {
		defer s.muPainting.RUnlock()
		s.repaintImpl(pc)
	} else if pc.full {
		// try cleanup stacked signals in chPaint
		var ignored = true
		for ignored {
			emptyIt(s.chPaint)
			time.Sleep(time.Millisecond)
			if s.muPainting.TryRLock() {
				ignored = false
				for _, grp := range s.groups {
					grp.repaint(pc)
				}
				s.muPainting.RUnlock()
			}
		}
	}
}

func (s *MPBV2) repaintImpl(pc *paintCtx) {
	var grp *GroupV2
	if grp, s.startIdx = s.chooseGroup(s.startIdx); grp != nil {
		if s.startIdx > 0 {
			g := s.groups[s.startIdx-1]
			pc.lastDoneCount = atomic.LoadInt32(&g.done)
			pc.lastDone = g.allDone()
		}
		grp.repaint(pc)
	}
}

// func (s *MPBV2) onDraw() {
// 	//
// }

var (
	errNotFound    = errors.New("not-found")
	errTaskExisted = errors.New("task-existed")
)
