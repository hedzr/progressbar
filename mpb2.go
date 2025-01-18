// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"os"
	"sync/atomic"

	"github.com/hedzr/progressbar/cursor"
)

type GroupedPB interface {
	MultiPB

	AddToGroup(group string, maxBytes int64, title string, opts ...Opt) (index int)
	RemoveFromGroup(group string, index int)

	// io.Writer
	// Close()

	// Add(maxBytes int64, title string, opts ...Opt) (index int)
	// Remove(index int)

	// Redraw()
	// SignalExit() <-chan struct{}
}

func multiBar2(opts ...MOpt) *mpbar2 {
	mpb2 := &mpbar2{
		mpbar: &mpbar{
			out:       os.Stdout,
			sigRedraw: make(chan struct{}, 16),
			sigExit:   make(chan struct{}, 16),
		},
	}

	for _, opt := range opts {
		opt(mpb2.mpbar)
	}

	go mpb2.run()
	return mpb2
}

type mpbar2 struct {
	*mpbar
	gb []*barsGroup
}

type barsGroup struct {
	bars  []*pbar
	title string
}

func (bars *barsGroup) Match(title string) bool {
	return bars.title == title
}

func (bars *barsGroup) Close() {
	for _, pb := range bars.bars {
		pb.Close()
	}
	bars.bars = nil
}

func (bars *barsGroup) Add(mpb *mpbar2, maxBytes int64, title string, opts ...Opt) (index int) {
	pb := defaultBytes(mpb, maxBytes, title, opts...).(*pbar) //nolint:errcheck //the call is always ok
	pb.stepper.SetIndentChars(indentChars)

	mpb.rw.Lock()
	bars.bars = append(bars.bars, pb)
	mpb.rw.Unlock()
	return len(bars.bars) - 1
}

func (mpb *mpbar2) Close() {
	if atomic.CompareAndSwapInt32(&mpb.closed, 0, 1) {
		close(mpb.sigExit)

		if mpb.gb != nil {
			mpb.rw.Lock()
			defer mpb.rw.Unlock()

			close(mpb.sigRedraw)
			mpb.sigRedraw = nil

			for _, bars := range mpb.gb {
				bars.Close()
			}
			mpb.bars = nil

			if mpb.bars != nil {
				for _, pb := range mpb.bars {
					pb.Close()
				}
				mpb.bars = nil
			}
		} else if mpb.bars != nil {
			mpb.rw.Lock()
			defer mpb.rw.Unlock()

			close(mpb.sigRedraw)
			mpb.sigRedraw = nil

			for _, pb := range mpb.bars {
				pb.Close()
			}
			mpb.bars = nil
		} else {
			mpb.rw.Lock()
			defer mpb.rw.Unlock()

			close(mpb.sigRedraw)
			mpb.sigRedraw = nil
		}
	}
}

func (mpb *mpbar2) Redraw() {
	if atomic.LoadInt32(&mpb.closed) == 0 {
		var sig chan struct{}
		mpb.rw.RLock()
		sig = mpb.sigRedraw
		mpb.rw.RUnlock()
		if sig != nil {
			sig <- struct{}{}
		}
	}
}

func (mpb *mpbar2) SignalExit() <-chan struct{} { return mpb.sigExit }

func (mpb *mpbar2) AddToGroup(group string, maxBytes int64, title string, opts ...Opt) (index int) {
	var found *barsGroup

	mpb.rw.RLock()
	for _, it := range mpb.gb {
		if yes := it.Match(group); yes {
			found = it
			break
		}
	}
	mpb.rw.RUnlock()

	mpb.rw.Lock()
	if found == nil {
		found = &barsGroup{title: group}
		mpb.gb = append(mpb.gb, found)
	}
	mpb.rw.Unlock()

	return found.Add(mpb, maxBytes, title, opts...)
}

func (mpb *mpbar2) Add(maxBytes int64, title string, opts ...Opt) (index int) {
	pb := defaultBytes(mpb, maxBytes, title, opts...).(*pbar) //nolint:errcheck //the call is always ok
	pb.stepper.SetIndentChars(indentChars)

	mpb.rw.Lock()
	mpb.bars = append(mpb.bars, pb)
	mpb.rw.Unlock()
	return len(mpb.bars) - 1
}

func (mpb *mpbar2) RemoveFromGroup(group string, index int) {
	var ix int
	var found *barsGroup

	mpb.rw.RLock()
	for i, it := range mpb.gb {
		if yes := it.Match(group); yes {
			ix, found = i, it
			break
		}
	}
	mpb.rw.RUnlock()

	if found != nil && ix >= 0 {
		if index > 0 && index < len(found.bars) {
			mpb.rw.Lock()
			found.bars = append(found.bars[0:index], found.bars[index+1:]...)
			defer mpb.rw.Unlock()
		}
	}
}

func (mpb *mpbar2) Remove(index int) {
	mpb.rw.Lock()
	defer mpb.rw.Unlock()
	if index > 0 && index < len(mpb.bars) {
		mpb.bars = append(mpb.bars[0:index], mpb.bars[index+1:]...)
	}
}

func (mpb *mpbar2) run() {
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

func (mpb *mpbar2) outFlush() {
	if ss, ok := mpb.out.(interface{ Sync() error }); ok {
		_ = ss.Sync()
	}
	if ss, ok := mpb.out.(interface{ Flush() error }); ok {
		_ = ss.Flush()
	}
}

func (mpb *mpbar2) redrawNow() {
	if !mpb.rw.TryRLock() {
		return
	}

	defer mpb.rw.RUnlock()

	if atomic.LoadInt32(&mpb.closed) == 1 {
		mpb.outFlush()
		return
	}

	if len(mpb.gb) > 0 {
		totalRows := 0
		for _, gv := range mpb.gb {
			totalRows++
			totalRows += len(gv.bars)
		}

		var first = atomic.CompareAndSwapInt32(&mpb.dirtyFlag, 0, 1)
		if !first {
			cursor.Left(1000)
			cursor.Up(totalRows - mpb.lines)
		}

		shouldBeDone, doneAll, rows := 0, 0, 0
		for _, gv := range mpb.gb {
			// if rows >= mpb.lines {
			// _, _ = mpb.out.Write([]byte(fmt.Sprintf("%-30s (%d items) %d/%d", gv.title, len(gv.bars), rows, totalRows)))
			_, _ = mpb.out.Write([]byte(gv.title))
			_, _ = mpb.out.Write([]byte("\n"))
			rows++

			done := true
			for _, pb := range gv.bars {
				str := pb.String()
				_, _ = mpb.out.Write([]byte(str))
				_, _ = mpb.out.Write([]byte("\n"))
				if pb.completed {
					rows++
				} else {
					done = false
					rows++
				}
			}

			if done {
				doneAll++
				// if atomic.CompareAndSwapInt32(&mpb.dirtyFlag, 1, 0) {
				// 	mpb.lines = len(gv.bars)
				// }
			}
			shouldBeDone++
		}
		if doneAll >= shouldBeDone {
			if mpb.onDone != nil {
				cb := mpb.onDone
				mpb.onDone = nil
				mpb.outFlush()
				cb(mpb)
			}
		}
	} else {
		var done = true
		var cnt = 0
		var first = atomic.CompareAndSwapInt32(&mpb.dirtyFlag, 0, 1)
		if !first {
			cursor.Left(1000)
			cursor.Up(len(mpb.bars) - mpb.lines)
		}

		for i, pb := range mpb.bars {
			if i >= mpb.lines {
				str := pb.String()
				_, _ = mpb.out.Write([]byte(str))
				_, _ = mpb.out.Write([]byte("\n"))
				// _, _ = fmt.Fprintf(mpb.out, "%s%s\n", indentChars, str)
				if !pb.completed {
					done = false
					cnt++
				}
			}
		}

		// _, _ = fmt.Fprintf(tui, "%v tasks activate [%v, %v, %v lines]\n", cnt, width, tui.Height(), len(mpb.bars)+1)
		// _ = tui.FlushN(len(mpb.bars) + 1)

		if done {
			// mpb.out.Flush()
			if atomic.CompareAndSwapInt32(&mpb.dirtyFlag, 1, 0) {
				mpb.lines = len(mpb.bars)
			}
			if mpb.onDone != nil {
				cb := mpb.onDone
				mpb.onDone = nil
				// mpb.sigRedraw <- struct{}{}
				cb(mpb)
			}
		}
	}
}

func (mpb *mpbar2) Write(data []byte) (n int, err error) {
	n, err = mpb.out.Write(data)
	// _ = mpb.out.Flush()
	return
}
