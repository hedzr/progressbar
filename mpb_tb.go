package progressbar

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (s *TaskBar) repaint(w Writer, i int, pc *paintCtx) {
	n, _ := w.WriteString(s.Name)
	_, _ = w.WriteString(strings.Repeat(" ", 32-n))
	progress := s.Progress()
	_, _ = w.WriteString(strconv.Itoa(int(progress)))
	_, _ = w.WriteString("\n")
	_, _ = i, pc
}

// String implements PB.
func (s *TaskBar) String() string {
	return s.stepper.String(s)
}

func (s *TaskBar) Bytes() []byte {
	return []byte(s.String())
}

func (s *TaskBar) State() (min, max, pos int64) {
	pos = atomic.LoadInt64(&s.progress)
	min = atomic.LoadInt64(&s.min)
	max = atomic.LoadInt64(&s.max)
	return
}

func (s *TaskBar) Completed() bool {
	pos := atomic.LoadInt64(&s.progress)
	max := atomic.LoadInt64(&s.max)
	return pos >= max
}

func (s *TaskBar) Min() int64      { return atomic.LoadInt64(&s.min) }
func (s *TaskBar) Max() int64      { return atomic.LoadInt64(&s.max) }
func (s *TaskBar) Progress() int64 { return atomic.LoadInt64(&s.progress) }
func (s *TaskBar) Increase(delta int64) (done bool) {
	val := atomic.AddInt64(&s.progress, delta)
	return val >= s.Max()
}

func (s *TaskBar) Done() (progress, max int64, done bool) {
	_, max, progress = s.State()
	// progress = atomic.LoadInt64(&s.progress)
	// max = atomic.LoadInt64(&s.max)
	done = progress >= max
	return
}

// func (s *TaskV2) runner(ctx context.Context, pc *paintCtx) {
// 	defer func() {
// 		//
// 	}()
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-pc.chPaint:
// 			s.repaint(pc)
// 		}
// 	}
// }

// Bounds implements PB.
func (s *TaskBar) Bounds() (lb int64, ub int64, progress int64) {
	lb, ub, progress = s.State()
	return
}

// Close implements PB.
func (s *TaskBar) Close() {
	if s.downloader != nil {
		s.downloader.Close()
	}
}

// LowerBound implements PB.
func (s *TaskBar) LowerBound() (lb int64) {
	lb = atomic.LoadInt64(&s.min)
	return
}

// Percent implements PB.
func (s *TaskBar) Percent() string {
	lb, ub, progress := s.State()
	percent := float64(progress) / float64(ub-lb)
	return fltfmtpercent(percent)
}

// PercentF implements PB.
func (s *TaskBar) PercentF() float64 {
	lb, ub, progress := s.State()
	percent := float64(progress) / float64(ub-lb)
	return percent
}

// PercentI implements PB.
func (s *TaskBar) PercentI() int {
	lb, ub, progress := s.State()
	percent := float64(progress) / float64(ub-lb)
	return int(percent*100 + 0.5)
}

// Resumeable implements PB.
func (s *TaskBar) Resumeable() bool {
	return true
}

// SetInitialValue implements PB.
func (s *TaskBar) SetInitialValue(initial int64) {
	atomic.StoreInt64(&s.progress, initial)
}

// SetResumeable implements PB.
func (s *TaskBar) SetResumeable(resumeable bool) {
}

// Step implements PB.
func (s *TaskBar) Step(delta int64) {
	_ = s.Increase(delta)
}

func (pb *TaskBar) Dur() (dur time.Duration) {
	if _, _, done := pb.Done(); !done {
		pb.stopTime = time.Now()
	}
	dur = pb.stopTime.Sub(pb.startTime)
	return
}

func (s *TaskBar) startNow() {
	now := time.Now()
	s.startTime = now.Add(-1 * time.Millisecond)
	s.stopTime = now
}

func (pb *TaskBar) Title() string { return pb.Name }

func (pb *TaskBar) SchemaDataPrepared(data *SchemaData) {
	if pb.onDataPrepared != nil {
		pb.onDataPrepared(pb, data)
	}
}

// UpdateRange implements PB.
func (s *TaskBar) UpdateRange(min, max int64) {
	atomic.StoreInt64(&s.min, min)
	atomic.StoreInt64(&s.max, max)
}

// UpperBound implements PB.
func (s *TaskBar) UpperBound() (ub int64) {
	ub = atomic.LoadInt64(&s.max)
	return
}

// Write implements PB.
func (s *TaskBar) Write(p []byte) (n int, err error) {
	n = len(p)
	// _ = s.Increase(int64(n))
	atomic.AddInt64(&s.progress, int64(n))
	s.dad.Repaint()
	return
}
