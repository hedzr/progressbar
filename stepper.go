// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"bytes"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/hedzr/progressbar/tool"
)

func MaxSteppers() int { return len(steppers) }

type BarT interface {
	String(pb *pbar) string
	Bytes(pb *pbar) []byte

	SetSchema(schema string)
	SetWidth(w int)
	SetIndentChars(s string)
	SetPrependText(s string)
	SetAppendText(s string)
	SetExtraTailSpaces(howMany int)

	SetBaseColor(clr int)
	SetHighlightColor(clr int)
}

var steppers = map[int]*stepper{
	// 0: python installer style
	0: {unread: "━", read: "━", leftHalf: "╺", rightHalf: "╸", clrBase: tool.FgDarkGray, clrHighlight: tool.FgYellow},

	// "▏", "▎", "▍", "▌", "▋", "▊", "▉"

	1: {unread: "▒", read: "▉", leftHalf: "▒", rightHalf: "▌", clrBase: tool.FgDarkGray, clrHighlight: tool.FgLightCyan},
	2: {unread: "-", read: "+", leftHalf: "+", rightHalf: "+", clrBase: tool.FgDarkGray, clrHighlight: tool.FgYellow},
	3: {unread: "&nbsp;", read: "=", leftHalf: ">", rightHalf: ">", clrBase: tool.FgDarkGray, clrHighlight: tool.FgYellow},
}

type stepper struct {
	tool.ColorTranslator
	tmpl             *template.Template
	unread           string
	read             string
	leftHalf         string
	rightHalf        string
	indentL          string
	prepend          string
	append           string
	schema           string
	clrBase          int
	clrHighlight     int
	barWidth         int
	safetyTailSpaces int
}

func (s *stepper) SetBaseColor(clr int) {
	s.clrBase = clr
}

func (s *stepper) SetHighlightColor(clr int) {
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

func (s *stepper) init() *stepper {
	if s.ColorTranslator == nil {
		s.ColorTranslator = tool.NewCPT()
	}
	if s.tmpl == nil {
		s.updateSchema()
	}
	if s.safetyTailSpaces == 0 {
		s.safetyTailSpaces = 8
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

func (s *stepper) buildBar(pb *pbar, pos, barWidth int, half bool) string {
	var sb bytes.Buffer
	var rightPart string
	if pos > 0 {
		leftPart := strings.Repeat(s.read, pos)
		sb.WriteString(s.Colorize(s.Translate(leftPart, 0), s.clrHighlight))
	}
	if !pb.completed {
		if half {
			sb.WriteString(s.Colorize(s.leftHalf, s.clrBase))
		} else {
			sb.WriteString(s.Colorize(s.rightHalf, s.clrHighlight))
		}
	}
	if barWidth > pos {
		rightPart = strings.Repeat(s.unread, barWidth-pos-1)
		sb.WriteString(s.Colorize(s.Translate(rightPart, 0), s.clrHighlight))
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

func (s *stepper) Bytes(pb *pbar) []byte {
	percent := float64(pb.read) / float64(pb.max-pb.min)

	if !pb.completed {
		pb.stopTime = time.Now()
	}
	dur := pb.stopTime.Sub(pb.startTime)

	read, suffix := humanizeBytes(float64(pb.read))
	total, suffix1 := humanizeBytes(float64(pb.max))
	speed, suffix2 := humanizeBytes(float64(pb.read) / dur.Seconds())

	pos1 := int(percent * float64(s.barWidth) * 2)
	half := pos1%2 == 0
	pos := pos1 / 2

	var sb bytes.Buffer
	data := &SchemaData{
		Indent:       s.indentL,
		Prepend:      s.prepend,
		Bar:          s.buildBar(pb, pos, s.barWidth, half),
		Percent:      fltfmtpercent(percent), // fmt.Sprintf("%.1f%%", percent),
		PercentFloat: percent,                // percent = 61 => 61.0%
		Title:        pb.title,               //
		Current:      read + suffix,          // fmt.Sprintf("%v%v", read, suffix),
		Total:        total + suffix1,        // fmt.Sprintf("%v%v", total, suffix1),
		Speed:        speed + suffix2 + "/s", // fmt.Sprintf("%v%v/s", speed, suffix2),
		Elapsed:      durfmt(dur),            // fmt.Sprintf("%v", dur), //nolint:gocritic
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

	str := s.Translate(sb.String(), 0)
	return []byte(str)
}

func (s *stepper) String(pb *pbar) string {
	return string(s.Bytes(pb))
}

const (
	defaultSchema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <font color="green">{{.Title}}</font> | {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}`
	barWidth      = 30
	indentChars   = `    `
)
