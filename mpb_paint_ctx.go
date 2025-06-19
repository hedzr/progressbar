package progressbar

type paintCtx struct {
	chPaint       <-chan struct{}
	bm            *MPBV2
	full          bool  // all tasks of all groups had done
	lastDoneCount int32 // in last group, just for trace
	lastDone      bool
}

func newPaintCtx(s *MPBV2) *paintCtx {
	pc := &paintCtx{
		s.chPaint,
		s,
		false,
		0,
		false,
	}
	return pc
}
