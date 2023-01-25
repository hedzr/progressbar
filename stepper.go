// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"bytes"
	"log"
	"math"
	"strings"
	"text/template"
	"time"

	"github.com/hedzr/progressbar/tool"
)

func MaxSteppers() int { return len(steppers) }

type barT interface {
	String(pb *pbar) string
	Bytes(pb *pbar) []byte

	SetSchema(schema string)
	SetWidth(w int)
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
	tmpl         *template.Template
	unread       string
	read         string
	leftHalf     string
	rightHalf    string
	indentL      string
	prepend      string
	append       string
	schema       string
	clrBase      int
	clrHighlight int
	barWidth     int
}

func (s *stepper) SetSchema(schema string) {
	s.schema = schema
}

func (s *stepper) SetWidth(w int) {
	s.barWidth = w
}

func (s *stepper) init() {
	if s.tmpl == nil {
		if s.schema == "" {
			s.schema = defaultSchema
		}
		if s.barWidth == 0 {
			s.barWidth = barWidth
		}
		s.tmpl = template.Must(template.New("bar-build").Parse(s.schema))
	}
}

func (s *stepper) buildBar(pb *pbar, pos, barWidth int, half bool) string {
	var sb bytes.Buffer
	cpt := tool.GetCPT()
	if pos > 0 {
		leftPart := strings.Repeat(s.read, pos)
		sb.WriteString(cpt.Colorize(cpt.Translate(leftPart, 0), s.clrHighlight))
	}
	if !pb.completed {
		var rightPart string
		if half {
			sb.WriteString(cpt.Colorize(s.leftHalf, s.clrBase))
		} else {
			sb.WriteString(cpt.Colorize(s.rightHalf, s.clrHighlight))
		}
		if barWidth > pos {
			rightPart = strings.Repeat(s.unread, barWidth-pos-1)
			sb.WriteString(cpt.Colorize(cpt.Translate(rightPart, 0), s.clrBase))
		}
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

func (s *stepper) String(pb *pbar) string {
	return string(s.Bytes(pb))
}

func (s *stepper) Bytes(pb *pbar) []byte {
	percent := float64(pb.read) / float64(pb.max-pb.min) * 100

	if !pb.completed {
		pb.stopTime = time.Now()
	}
	dur := pb.stopTime.Sub(pb.startTime)

	read, suffix := humanizeBytes(float64(pb.read))
	total, suffix1 := humanizeBytes(float64(pb.max))
	speed, suffix2 := humanizeBytes(float64(pb.read) / dur.Seconds())

	pos1 := int(percent * 60 / 100)
	half := pos1%2 == 0
	pos := pos1 / 2

	cpt := tool.GetCPT()
	var sb bytes.Buffer

	data := &schemaData{
		Indent:  s.indentL,
		Prepend: s.prepend,
		Bar:     s.buildBar(pb, pos, s.barWidth, half),
		Percent: fltfmt(percent), // fmt.Sprintf("%.1f%%", percent),
		Title:   pb.title,        //
		Current: read + suffix,   // fmt.Sprintf("%v%v", read, suffix),
		Total:   total + suffix1, // fmt.Sprintf("%v%v", total, suffix1),
		Speed:   speed + suffix2, // fmt.Sprintf("%v%v/s", speed, suffix2),
		Elapsed: durfmt(dur),     // fmt.Sprintf("%v", dur), //nolint:gocritic
		Append:  s.append,
	}

	s.init()
	err := s.tmpl.Execute(&sb, data)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// str := sb.Bytes()
	// return str

	str := cpt.Translate(sb.String(), 0)
	return []byte(str)
}

func humanizeBytes(s float64) (value, suffix string) {
	sizes := []string{" B", " kB", " MB", " GB", " TB", " PB", " EB"}
	base := 1024.0
	if s < 10 {
		// return fmt.Sprintf("%2.0f", s), "B"
		return fltfmt(s), "B"
	}

	e := math.Floor(logn(s, base))
	suffix = sizes[int(e)]
	val := math.Floor(s/math.Pow(base, e)*10+0.5) / 10

	// f := "%.0f"
	// if val < 10 {
	// 	f = "%.1f"
	// }
	// value = fmt.Sprintf(f, val)

	value = fltfmt(val)
	return
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

type schemaData struct {
	Indent  string
	Prepend string
	Bar     string
	Percent string
	Title   string
	Current string
	Total   string
	Elapsed string
	Speed   string
	Append  string
}

const (
	defaultSchema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <font color="green">{{.Title}}</font> | {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}`
	barWidth      = 30
	indentChars   = `    `
)
