// Copyright © 2022 Atonal Authors
//

package progressbar

import "io"

type Opt func(pb *pbar)

func WithBarSpinner(whichOne int) Opt {
	return func(pb *pbar) {
		if s, ok := spinners[whichOne]; ok {
			pb.stepper = s
		}
	}
}

func WithBarStepper(whichOne int) Opt {
	return func(pb *pbar) {
		if s, ok := steppers[whichOne]; ok {
			pb.stepper = s
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

type (
	OnDone func(mpb MultiPB)
	MOpt   func(mpb *mpbar)
)

func WithOnDone(cb OnDone) MOpt {
	return func(mpb *mpbar) {
		mpb.onDone = cb
	}
}

func WithOutput(out io.Writer) MOpt {
	return func(mpb *mpbar) {
		mpb.out = out
	}
}
