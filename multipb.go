// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/hedzr/progressbar/cursor"
)

type MultiPB interface {
	io.Writer
	Close()

	Add(maxBytes int64, title string, opts ...Opt) (index int)
	Remove(index int)

	Redraw()
	SignalExit() <-chan struct{}
}

func multiBar(opts ...MOpt) *mpbar {
	bar := &mpbar{
		out:       os.Stdout,
		sigRedraw: make(chan struct{}, 16),
		sigExit:   make(chan struct{}, 16),
	}

	for _, opt := range opts {
		opt(bar)
	}

	go bar.run()
	return bar
}

var defaultMPB = multiBar()

type mpbar struct {
	out       io.Writer
	sigRedraw chan struct{}
	sigExit   chan struct{}
	onDone    OnDone

	bars []*pbar

	rw sync.RWMutex

	dirtyFlag int32
	closed    int32
}

func (mpb *mpbar) Close() {
	if atomic.CompareAndSwapInt32(&mpb.closed, 0, 1) {
		close(mpb.sigExit)

		if mpb.bars != nil {
			mpb.rw.Lock()
			defer mpb.rw.Unlock()

			close(mpb.sigRedraw)
			mpb.sigRedraw = nil

			for _, pb := range mpb.bars {
				pb.Close()
			}
			mpb.bars = nil
		} else {
			close(mpb.sigRedraw)
			mpb.sigRedraw = nil
		}
	}
}

func (mpb *mpbar) Redraw() {
	if atomic.LoadInt32(&mpb.closed) == 0 {
		if mpb.sigRedraw != nil {
			mpb.sigRedraw <- struct{}{}
		}
	}
}

func (mpb *mpbar) SignalExit() <-chan struct{} { return mpb.sigExit }

func (mpb *mpbar) Add(maxBytes int64, title string, opts ...Opt) (index int) {
	pb := defaultBytes(mpb, maxBytes, title, opts...).(*pbar) //nolint:errcheck //the call is always ok
	pb.stepper.SetIndentChars(indentChars)

	mpb.rw.Lock()
	defer mpb.rw.Unlock()
	mpb.bars = append(mpb.bars, pb)
	return len(mpb.bars) - 1
}

func (mpb *mpbar) Remove(index int) {
	mpb.rw.Lock()
	defer mpb.rw.Unlock()
	if index > 0 && index < len(mpb.bars) {
		mpb.bars = append(mpb.bars[0:index], mpb.bars[index+1:]...)
	}
}

func (mpb *mpbar) run() {
	// // ticker := time.NewTicker(time.Millisecond * 50)
	// ticker := time.NewTicker(time.Second * 1)
	// defer ticker.Stop()

	for {
		select {
		// case <-ticker.C:
		// 	mpb.redrawNow()
		case <-mpb.sigRedraw:
			mpb.redrawNow()
		case <-mpb.sigExit:
			return
		}
	}
}

func (mpb *mpbar) redrawNow() {
	if !mpb.rw.TryRLock() {
		return
	}

	defer mpb.rw.RUnlock()

	if atomic.LoadInt32(&mpb.closed) == 1 {
		if ss, ok := mpb.out.(interface{ Sync() error }); ok {
			_ = ss.Sync()
		}
		if ss, ok := mpb.out.(interface{ Flush() error }); ok {
			_ = ss.Flush()
		}
		return
	}

	var done = true
	var cnt int

	var first = atomic.CompareAndSwapInt32(&mpb.dirtyFlag, 0, 1)
	if !first {
		cursor.Left(1000)
		cursor.Up(len(mpb.bars))
	}

	for _, pb := range mpb.bars {
		str := pb.String()
		_, _ = mpb.out.Write([]byte(str))
		_, _ = mpb.out.Write([]byte("\n"))
		// _, _ = fmt.Fprintf(mpb.out, "%s%s\n", indentChars, str)
		if !pb.completed {
			done = false
			cnt++
		}
	}
	// _, _ = fmt.Fprintf(tui, "%v tasks activate [%v, %v, %v lines]\n", cnt, width, tui.Height(), len(mpb.bars)+1)
	// _ = tui.FlushN(len(mpb.bars) + 1)

	// if first {
	// 	cursor.Save()
	// }

	if done {
		// mpb.out.Flush()
		atomic.CompareAndSwapInt32(&mpb.dirtyFlag, 1, 0)
		if mpb.onDone != nil {
			cb := mpb.onDone
			mpb.onDone = nil

			mpb.sigRedraw <- struct{}{}

			cb(mpb)
		}
	}
}

func (mpb *mpbar) Write(data []byte) (n int, err error) {
	n, err = mpb.out.Write(data)
	// _ = mpb.out.Flush()
	return
}
