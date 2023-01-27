// Copyright © 2022 Atonal Authors
//

package progressbar

import "io"

type Opt func(pb *pbar)

func WithBarSpinner(whichOne int) Opt {
	return func(pb *pbar) {
		if s, ok := spinners[whichOne]; ok {
			pb.stepper = s.init()
		}
	}
}

func WithBarStepper(whichOne int) Opt {
	return func(pb *pbar) {
		if s, ok := steppers[whichOne]; ok {
			pb.stepper = s.init()
		}
	}
}

func WithBarUpperBound(ub int64) Opt {
	return func(pb *pbar) {
		pb.max = ub
	}
}

func WithBarWidth(w int) Opt {
	return func(pb *pbar) {
		pb.stepper.SetWidth(w)
	}
}

// WithBarTextSchema allows cha
//
//	"{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | {{.Title}} | {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}"
func WithBarTextSchema(schema string) Opt {
	return func(pb *pbar) {
		pb.stepper.SetSchema(schema)
	}
}

func WithBarIndentChars(str string) Opt {
	return func(pb *pbar) {
		pb.stepper.SetIndentChars(str)
	}
}

func WithBarPrependText(str string) Opt {
	return func(pb *pbar) {
		pb.stepper.SetPrependText(str)
	}
}

func WithBarAppendText(str string) Opt {
	return func(pb *pbar) {
		pb.stepper.SetAppendText(str)
	}
}

// WithBarExtraTailSpaces specifies how many spaces will be printed
// at end of each bar. These spaces can wipe out the dirty tail of
// line.
//
// Default is 8 (spaces). You may specify -1 to disable extra
// spaces to be printed.
func WithBarExtraTailSpaces(howMany int) Opt {
	return func(pb *pbar) {
		pb.stepper.SetExtraTailSpaces(howMany)
	}
}

func WithBarWorker(w Worker) Opt {
	return func(pb *pbar) {
		pb.worker = w
	}
}

func WithBarOnCompleted(cb OnCompleted) Opt {
	return func(pb *pbar) {
		pb.onComp = cb
	}
}

func WithBarOnStart(cb OnStart) Opt {
	return func(pb *pbar) {
		pb.onStart = cb
	}
}

func WithBarOnDataPrepared(cb OnDataPrepared) Opt {
	return func(pb *pbar) {
		pb.onDataPrepared = cb
	}
}

type (
	OnDone func(mpb MultiPB)
	MOpt   func(mpb *mpbar)
)

func WithOnDone(cb OnDone) MOpt {
	return func(mpb *mpbar) {
		mpb.onDone = cb
	}
}

func WithOutputDevice(out io.Writer) MOpt {
	return func(mpb *mpbar) {
		mpb.out = out
	}
}
