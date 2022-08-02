// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"io"
	"sync"
	"sync/atomic"
)

func NewTasks(bar MultiPB) *Tasks {
	return &Tasks{bar: bar}
}

type Tasks struct {
	wg    sync.WaitGroup
	bar   MultiPB
	tasks []*aTask
}

type taskOptions struct {
	title      string
	barOptions []Opt
	onStart    OnStart
	onStop     OnCompleted
	onWork     Worker
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
	task.wg = &s.wg

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
	wg *sync.WaitGroup
	w  io.Writer

	buf []byte

	doneCount int32

	onStartProc     OnStart
	onCompletedProc OnCompleted
	onStepProc      Worker
}

func (s *sTask) Close() {
}

// type Stepper interface {
// 	Step(delta int64)
// }

func (s *sTask) onStep(bar PB, exitCh <-chan struct{}) {
	if s.onStepProc != nil {
		s.onStepProc(bar, exitCh)
	}
	return
}

func (s *sTask) onStart(bar PB) {
	if s.onStartProc != nil {
		s.onStartProc(bar)
	}
}

func (s *sTask) onCompleted(bar PB) {
	if s.onCompletedProc != nil {
		s.onCompletedProc(bar)
	}
	wg := s.wg
	s.wg = nil
	wg.Done()
	atomic.AddInt32(&s.doneCount, 1)
}
