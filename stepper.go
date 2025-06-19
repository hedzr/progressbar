// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"bytes"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/hedzr/is/term/color"
)

func MaxSteppers() int { return len(steppers) }

type BarT interface {
	StringV1(pb *pbar) string
	BytesV1(pb *pbar) []byte

	String(bar MiniResizeableBar) string
	Bytes(bar MiniResizeableBar) []byte

	Percent() string   // just for stepper
	PercentF() float64 // return 0.905
	PercentI() int     // '90.5' -> return 91

	Resumeable() bool
	SetResumeable(resumeable bool)
	SetInitialValue(initial int64)

	SetSchema(schema string)
	SetWidth(w int)
	SetIndentChars(s string)
	SetPrependText(s string)
	SetAppendText(s string)
	SetExtraTailSpaces(howMany int)

	SetBaseColor(clr color.Color)
	SetHighlightColor(clr color.Color)
}

type StepperOpt func(s BarT)

func WithStepperResumeable(resumeable bool) StepperOpt {
	return func(s BarT) {
		s.SetResumeable(resumeable)
	}
}

func WithStepperInitialValue(initial int64) StepperOpt {
	return func(s BarT) {
		s.SetInitialValue(initial)
	}
}

func WithStepperSchema(schema string) StepperOpt {
	return func(s BarT) {
		s.SetSchema(schema)
	}
}

func WithStepperWidth(width int) StepperOpt {
	return func(s BarT) {
		s.SetWidth(width)
	}
}

func WithStepperIndentChars(text string) StepperOpt {
	return func(s BarT) {
		s.SetIndentChars(text)
	}
}

func WithStepperPrependText(text string) StepperOpt {
	return func(s BarT) {
		s.SetPrependText(text)
	}
}

func WithStepperAppendText(text string) StepperOpt {
	return func(s BarT) {
		s.SetAppendText(text)
	}
}

func WithStepperTailSpace(count int) StepperOpt {
	return func(s BarT) {
		s.SetExtraTailSpaces(count)
	}
}

func WithStepperBaseColor(clr color.Color) StepperOpt {
	return func(s BarT) {
		s.SetBaseColor(clr)
	}
}

func WithStepperHighlightColor(clr color.Color) StepperOpt {
	return func(s BarT) {
		s.SetHighlightColor(clr)
	}
}

var steppers = map[int]*stepper{
	// 0: python installer style
	0: {unread: "━", read: "━", leftHalf: "╺", rightHalf: "╸", clrBase: color.FgDarkGray, clrHighlight: color.NewColor16m(173, 147, 77, false), clrHighlight16M: color.FgYellow},

	// "▏", "▎", "▍", "▌", "▋", "▊", "▉"

	1: {unread: "▒", read: "▉", leftHalf: "▒", rightHalf: "▌", clrBase: color.FgDarkGray, clrHighlight: color.FgLightCyan},
	2: {unread: "-", read: "+", leftHalf: "+", rightHalf: "+", clrBase: color.FgDarkGray, clrHighlight: color.FgYellow},
	3: {unread: "&nbsp;", read: "=", leftHalf: ">", rightHalf: ">", clrBase: color.FgDarkGray, clrHighlight: color.FgYellow},
}

type stepper struct {
	tr               color.Translator
	tmpl             *template.Template
	unread           string
	read             string
	leftHalf         string
	rightHalf        string
	indentL          string
	prepend          string
	append           string
	schema           string
	clrBase          color.Color
	clrHighlight     color.Color
	clrHighlight16M  color.Color
	barWidth         int
	safetyTailSpaces int
	percent          float64
	initial          int64
	resumeable       bool
}

func (s *stepper) SetInitialValue(initial int64) {
	s.initial = initial
}

func (s *stepper) SetResumeable(resumeable bool) {
	s.resumeable = resumeable
}

func (s *stepper) Resumeable() bool {
	return s.resumeable
}

func (s *stepper) SetBaseColor(clr color.Color) {
	s.clrBase = clr
}

func (s *stepper) SetHighlightColor(clr color.Color) {
	s.clrHighlight = clr
}

func (s *stepper) SetSchema(schema string) {
	s.schema = schema
	s.updateSchema()
}

func (s *stepper) SetWidth(w int) {
	s.barWidth = w
}

func (s *stepper) SetIndentChars(str string) {
	s.indentL = str
}

func (s *stepper) SetPrependText(str string) {
	s.prepend = str
}

func (s *stepper) SetAppendText(str string) {
	s.append = str
}

func (s *stepper) SetExtraTailSpaces(howMany int) {
	s.safetyTailSpaces = howMany
}

func (s *stepper) init(opts ...StepperOpt) *stepper {
	if s.tr == nil {
		s.tr = color.GetCPT()
	}
	if s.tmpl == nil {
		s.updateSchema()
	}
	if s.safetyTailSpaces == 0 {
		s.safetyTailSpaces = 8
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *stepper) updateSchema() *stepper {
	if s.schema == "" {
		s.schema = defaultSchema
	}
	if s.barWidth == 0 {
		s.barWidth = barWidth
	}
	s.tmpl = template.Must(template.New("bar-build").Parse(s.schema))
	return s
}

func (s *stepper) buildBar(pb MiniResizeableBar, pos, barWidth int, half bool) string {
	var sb bytes.Buffer
	var rightPart string
	if pos > 0 {
		leftPart := strings.Repeat(s.read, pos)
		s.tr.ColoredFast(&sb, s.clrHighlight, s.tr.Translate(leftPart, color.Reset))
	}
	if !pb.Completed() {
		if half {
			s.tr.ColoredFast(&sb, s.clrBase, s.leftHalf)
		} else {
			s.tr.ColoredFast(&sb, s.clrHighlight, s.rightHalf)
		}
	}
	if barWidth > pos {
		rightPart = strings.Repeat(s.unread, barWidth-pos-1)
		s.tr.ColoredFast(&sb, s.clrHighlight, s.tr.Translate(rightPart, color.Reset))
	}
	return sb.String()
}

// func (s *stepper) buildPrepend(pb *pbar, pos, barWidth int, half bool) string {
// 	var sb bytes.Buffer
// 	return sb.String()
// }
//
// func (s *stepper) buildAppend(pb *pbar, pos, barWidth int, half bool) string {
// 	var sb bytes.Buffer
// 	return sb.String()
// }

func (s *stepper) String(bar MiniResizeableBar) string {
	min, max, progress := bar.State()
	s.percent = float64(progress) / float64(max-min)
	if s.percent > 1 {
		s.percent = 1
	}

	dur := bar.Dur()

	read, suffix := humanizeBytes(float64(progress))
	total, suffix1 := humanizeBytes(float64(max))
	speed, suffix2 := humanizeBytes(float64(progress) / dur.Seconds())

	pos1 := int(s.percent * float64(s.barWidth) * 2)
	half := pos1%2 == 0
	pos := pos1 / 2

	var sb bytes.Buffer
	data := &SchemaData{
		Indent:       s.indentL,
		Prepend:      s.prepend,
		Bar:          s.buildBar(bar, pos, s.barWidth, half),
		Percent:      fltfmtpercent(s.percent), // fmt.Sprintf("%.1f%%", percent),
		PercentFloat: s.percent,                // percent = 61 => 61.0%
		Title:        bar.Title(),              //
		Current:      read + suffix,            // fmt.Sprintf("%v%v", read, suffix),
		Total:        total + suffix1,          // fmt.Sprintf("%v%v", total, suffix1),
		Speed:        speed + suffix2 + "/s",   // fmt.Sprintf("%v%v/s", speed, suffix2),
		Elapsed:      durfmt(dur),              // fmt.Sprintf("%v", dur), //nolint:gocritic
		ElapsedTime:  dur,
		Append:       s.append,
	}

	s.init()

	bar.SchemaDataPrepared(data)

	err := s.tmpl.Execute(&sb, data)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// str := sb.Bytes()
	// return str

	if s.safetyTailSpaces > 0 {
		sb.WriteString(strings.Repeat(" ", s.safetyTailSpaces))
	}

	str := s.tr.Translate(sb.String(), color.Reset)
	return str
}

func (s *stepper) Bytes(bar MiniResizeableBar) []byte {
	return []byte(s.String(bar))
}

func (s *stepper) BytesV1(pb *pbar) []byte {
	s.percent = float64(max(s.initial, pb.read)) / float64(pb.max-pb.min)
	if s.percent > 1 {
		s.percent = 1
	}

	if !pb.completed {
		pb.stopTime = time.Now()
	}
	dur := pb.stopTime.Sub(pb.startTime)

	read, suffix := humanizeBytes(float64(pb.read))
	total, suffix1 := humanizeBytes(float64(pb.max))
	speed, suffix2 := humanizeBytes(float64(pb.read) / dur.Seconds())

	pos1 := int(s.percent * float64(s.barWidth) * 2)
	half := pos1%2 == 0
	pos := pos1 / 2

	var sb bytes.Buffer
	data := &SchemaData{
		Indent:       s.indentL,
		Prepend:      s.prepend,
		Bar:          s.buildBar(pb, pos, s.barWidth, half),
		Percent:      fltfmtpercent(s.percent), // fmt.Sprintf("%.1f%%", percent),
		PercentFloat: s.percent,                // percent = 61 => 61.0%
		Title:        pb.title,                 //
		Current:      read + suffix,            // fmt.Sprintf("%v%v", read, suffix),
		Total:        total + suffix1,          // fmt.Sprintf("%v%v", total, suffix1),
		Speed:        speed + suffix2 + "/s",   // fmt.Sprintf("%v%v/s", speed, suffix2),
		Elapsed:      durfmt(dur),              // fmt.Sprintf("%v", dur), //nolint:gocritic
		ElapsedTime:  dur,
		Append:       s.append,
	}

	s.init()

	if pb.onDataPrepared != nil {
		pb.onDataPrepared(pb, data)
	}

	err := s.tmpl.Execute(&sb, data)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// str := sb.Bytes()
	// return str

	if s.safetyTailSpaces > 0 {
		sb.WriteString(strings.Repeat(" ", s.safetyTailSpaces))
	}

	str := s.tr.Translate(sb.String(), color.Reset)
	return []byte(str)
}

func (s *stepper) StringV1(pb *pbar) string {
	return string(s.BytesV1(pb))
}

func (s *stepper) Percent() string {
	return fltfmtpercent(s.percent)
}

func (s *stepper) PercentF() float64 {
	return s.percent
}

func (s *stepper) PercentI() int {
	return int(s.percent*100 + 0.5)
}

const (
	defaultSchema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <font color="green">{{.Title}}</font> | {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}`
	barWidth      = 30
	indentChars   = `    `
)
