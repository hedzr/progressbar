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
		stepper:   steppers[0],
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
	Worker      func(bar PB, exitCh <-chan struct{})
	OnCompleted func(bar PB)
	OnStart     func(bar PB)
)

type pbar struct {
	min, max  int64
	title     string
	startTime time.Time
	stopTime  time.Time

	stepper barT

	read       int64
	row        int
	muPainting sync.Mutex
	completed  bool

	mpbar MultiPB

	worker  Worker
	onComp  OnCompleted
	onStart OnStart
}

func (pb *pbar) Close() {
	// if atomic.CompareAndSwapInt32(&pb.closed, 0, 1) {
	// 	close(pb.sigExit)
	// 	close(pb.sigRedraw)
	// }
}

func (pb *pbar) LowerBound() int64 { return pb.min }
func (pb *pbar) UpperBound() int64 { return pb.max }

func (pb *pbar) UpdateRange(min, max int64) {
	pb.min, pb.max = min, max
}

func (pb *pbar) Step(delta int64) {
	pb.read += delta
	pb.invalidate()
}

func (pb *pbar) Write(data []byte) (n int, err error) {
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
