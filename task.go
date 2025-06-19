// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"io"
	"sync"
	"sync/atomic"
)

// NewTasks creates a Tasks container which you can add the tasks
// in it.
//
//	tasks := progressbar.NewTasks(progressbar.New())
//	defer tasks.Close()
//
//	max := count
//	_, h, _ := terminal.GetSize(int(os.Stdout.Fd()))
//	if max >= h {
//		max = h
//	}
//
//	for i := whichStepper; i < whichStepper+max; i++ {
//		tasks.Add(
//			progressbar.WithTaskAddBarOptions(
//				progressbar.WithBarStepper(i),
//				progressbar.WithBarUpperBound(100),
//				progressbar.WithBarWidth(32),
//			),
//			progressbar.WithTaskAddBarTitle("Task "+strconv.Itoa(i)), // fmt.Sprintf("Task %v", i)),
//			progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
//				for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
//					ms := time.Duration(10 + rand.Intn(300)) //nolint:gosec //just a demo
//					time.Sleep(time.Millisecond * ms)
//					bar.Step(1)
//				}
//			}),
//		)
//	}
//
//	tasks.Wait()
//
// Above.
func NewTasks(bar MultiPB) *Tasks {
	return &Tasks{bar: bar}
}

type Tasks struct {
	bar   MultiPB
	tasks []*DownloadTask
	wg    sync.WaitGroup
}

type taskOptions struct {
	onStart    OnStart
	onStop     OnCompleted
	onWork     Worker
	title      string
	barOptions []Opt
}

type TaskOpt func(s *taskOptions)

func WithTaskAddBarTitle(title string) TaskOpt {
	return func(s *taskOptions) {
		s.title = title
	}
}

func WithTaskAddBarOptions(opts ...Opt) TaskOpt {
	return func(s *taskOptions) {
		s.barOptions = opts
	}
}

func WithTaskAddOnTaskInitializing(cb OnStart) TaskOpt {
	return func(s *taskOptions) {
		s.onStart = cb
	}
}

func WithTaskAddOnTaskCompleted(cb OnCompleted) TaskOpt {
	return func(s *taskOptions) {
		s.onStop = cb
	}
}

func WithTaskAddOnTaskProgressing(cb Worker) TaskOpt {
	return func(s *taskOptions) {
		s.onWork = cb
	}
}

func (s *Tasks) Close() {
	s.bar.Close()
}

func (s *Tasks) Add(opts ...TaskOpt) *Tasks {
	to := new(taskOptions)
	for _, opt := range opts {
		opt(to)
	}

	task := &sTask{
		wg:              &s.wg,
		w:               s.bar,
		buf:             nil,
		doneCount:       0,
		onStartProc:     to.onStart,
		onCompletedProc: to.onStop,
		onStepProc:      to.onWork,
	}

	var o []Opt
	o = append(o,
		WithBarWorker(task.onStep),
		WithBarOnCompleted(task.onCompleted),
		WithBarOnStart(task.onStart),
	)
	o = append(o, to.barOptions...)

	s.bar.Add(
		100,
		to.title,
		o...,
	)

	s.wg.Add(1)
	return s
}

func (s *Tasks) Wait() {
	s.wg.Wait()
}

type sTask struct {
	w               io.Writer
	wg              *sync.WaitGroup
	onStartProc     OnStart
	onCompletedProc OnCompleted
	onStepProc      Worker
	buf             []byte
	doneCount       int32
}

func (s *sTask) Close() {
	if wg := s.wg; wg != nil {
		wg.Done()
	}
}

// type Stepper interface {
// 	Step(delta int64)
// }

func (s *sTask) onStep(bar MiniResizeableBar, exitCh <-chan struct{}) (stop bool) {
	if s.onStepProc != nil {
		s.onStepProc(bar, exitCh)
	}
	return
}

func (s *sTask) onStart(bar MiniResizeableBar) {
	if s.onStartProc != nil {
		s.onStartProc(bar)
	}
}

func (s *sTask) onCompleted(bar MiniResizeableBar) {
	if s.onCompletedProc != nil {
		s.onCompletedProc(bar)
	}
	wg := s.wg
	s.wg = nil
	wg.Done()
	atomic.AddInt32(&s.doneCount, 1)
}
