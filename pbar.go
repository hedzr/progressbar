// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"io"
	"sync"
	"time"
)

func defaultBytes(mpbar MultiPB, maxBytes int64, title string, opts ...Opt) PB {
	pb := &pbar{
		mpbar:     mpbar,
		max:       maxBytes,
		title:     title,
		stepper:   steppers[0].init(),
		startTime: time.Now(),
		// sigRedraw: make(chan struct{}),
		// sigExit:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(pb)
	}

	go pb.run()

	return pb
}

type PB interface {
	io.Writer
	Close()
	String() string

	UpdateRange(min, max int64)
	LowerBound() int64
	UpperBound() int64

	Step(delta int64)
}

type (
	Worker         func(bar PB, exitCh <-chan struct{})
	OnCompleted    func(bar PB)
	OnStart        func(bar PB)
	OnDataPrepared func(bar PB, data *SchemaData)
)

type pbar struct {
	stopTime  time.Time
	startTime time.Time

	stepper         BarT           // stepper or spinner here
	mpbar           MultiPB        //
	stepperPostInit func(bar BarT) //

	worker         Worker
	onComp         OnCompleted
	onStart        OnStart
	onDataPrepared OnDataPrepared

	title string

	read int64
	min  int64
	max  int64
	row  int

	muPainting sync.Mutex

	completed bool
}

func (pb *pbar) Close() {
	pb.muPainting.Lock()
	defer pb.muPainting.Unlock()

	// if atomic.CompareAndSwapInt32(&pb.closed, 0, 1) {
	// 	close(pb.sigExit)
	// 	close(pb.sigRedraw)
	// }
}

func (pb *pbar) LowerBound() int64 { return pb.min }
func (pb *pbar) UpperBound() int64 { return pb.max }

func (pb *pbar) UpdateRange(min, max int64) {
	pb.muPainting.Lock()
	defer pb.muPainting.Unlock()

	pb.min, pb.max = min, max
}

func (pb *pbar) Step(delta int64) {
	pb.muPainting.Lock()
	defer pb.muPainting.Unlock()

	pb.read += delta
	pb.invalidate()
}

func (pb *pbar) Write(data []byte) (n int, err error) {
	pb.muPainting.Lock()
	defer pb.muPainting.Unlock()

	n = len(data)
	pb.read += int64(n)
	pb.invalidate()
	return
}

func (pb *pbar) invalidate() {
	if pb.read >= pb.max {
		pb.completed = true

		if pb.onComp != nil {
			cb := pb.onComp
			pb.onComp = nil
			cb(pb)
		}
	}
	pb.redraw()
}

func (pb *pbar) redraw() {
	pb.mpbar.Redraw()
}

func (pb *pbar) run() {
	if pb.onStart != nil {
		pb.onStart(pb)
	}

	if pb.stepperPostInit != nil {
		pb.stepperPostInit(pb.stepper)
	}

	if pb.worker != nil {
		go func() {
			pb.worker(pb, pb.mpbar.SignalExit())
		}()
	}
}

// func (pb *pbar) run() {
// 	if pb.worker != nil {
// 		go func() {
// 			pb.worker(pb.sigExit)
// 		}()
// 	}
//
// 	for {
// 		select {
// 		case <-pb.sigRedraw:
// 			if pb.spinner != nil {
// 				pb.spinner.draw(pb)
// 			} else {
// 				pb.stepper.draw(pb)
// 			}
// 		case <-pb.sigExit:
// 			return
// 		}
// 	}
// }

func (pb *pbar) Bytes() []byte {
	return pb.stepper.Bytes(pb)
}

func (pb *pbar) String() string {
	return pb.stepper.String(pb)
}

func (pb *pbar) locker() func() {
	pb.muPainting.Lock()
	return pb.muPainting.Unlock
}
