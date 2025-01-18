// Copyright © 2022 Atonal Authors
//

package progressbar

import (
	"bytes"
	"log"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/hedzr/progressbar/tool"
)

func MaxSpinners() int { return len(spinners) }

// The following spinner templates are modified from https://github.com/schollz/progressbar
var spinners = map[int]*spinner{
	0:  {chars: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}},
	1:  {chars: []string{"▁", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▁"}},
	2:  {chars: []string{"▖", "▘", "▝", "▗"}},
	3:  {chars: []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}},
	4:  {chars: []string{"◢", "◣", "◤", "◥"}},
	5:  {chars: []string{"◰", "◳", "◲", "◱"}},
	6:  {chars: []string{"◴", "◷", "◶", "◵"}},
	7:  {chars: []string{"◐", "◓", "◑", "◒"}},
	8:  {chars: []string{".", "o", "O", "@", "*"}},
	9:  {chars: []string{"|", "/", "-", "\\"}},
	10: {chars: []string{"◡◡", "⊙⊙", "◠◠"}},
	11: {chars: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}},
	12: {chars: []string{">))'>", " >))'>", "  >))'>", "   >))'>", "    >))'>", "   <'((<", "  <'((<", " <'((<"}},
	13: {chars: []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}},
	14: {chars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}},
	15: {chars: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}},
	16: {chars: []string{"▉", "▊", "▋", "▌", "▍", "▎", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}},
	17: {chars: []string{"■", "□", "▪", "▫"}},
	18: {chars: []string{"←", "↑", "→", "↓"}},
	19: {chars: []string{"╫", "╪"}},
	20: {chars: []string{"⇐", "⇖", "⇑", "⇗", "⇒", "⇘", "⇓", "⇙"}},
	21: {chars: []string{"⠁", "⠁", "⠉", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠤", "⠄", "⠄", "⠤", "⠠", "⠠", "⠤", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋", "⠉", "⠈", "⠈"}},
	22: {chars: []string{"⠈", "⠉", "⠋", "⠓", "⠒", "⠐", "⠐", "⠒", "⠖", "⠦", "⠤", "⠠", "⠠", "⠤", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋", "⠉", "⠈"}},
	23: {chars: []string{"⠁", "⠉", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠤", "⠄", "⠄", "⠤", "⠴", "⠲", "⠒", "⠂", "⠂", "⠒", "⠚", "⠙", "⠉", "⠁"}},
	24: {chars: []string{"⠋", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋"}},
	25: {chars: []string{"ｦ", "ｧ", "ｨ", "ｩ", "ｪ", "ｫ", "ｬ", "ｭ", "ｮ", "ｯ", "ｱ", "ｲ", "ｳ", "ｴ", "ｵ", "ｶ", "ｷ", "ｸ", "ｹ", "ｺ", "ｻ", "ｼ", "ｽ", "ｾ", "ｿ", "ﾀ", "ﾁ", "ﾂ", "ﾃ", "ﾄ", "ﾅ", "ﾆ", "ﾇ", "ﾈ", "ﾉ", "ﾊ", "ﾋ", "ﾌ", "ﾍ", "ﾎ", "ﾏ", "ﾐ", "ﾑ", "ﾒ", "ﾓ", "ﾔ", "ﾕ", "ﾖ", "ﾗ", "ﾘ", "ﾙ", "ﾚ", "ﾛ", "ﾜ", "ﾝ"}},
	26: {chars: []string{".", "..", "..."}},
	27: {chars: []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▉", "▊", "▋", "▌", "▍", "▎", "▏", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█", "▇", "▆", "▅", "▄", "▃", "▂", "▁"}},
	28: {chars: []string{".", "o", "O", "°", "O", "o", "."}},
	29: {chars: []string{"+", "x"}},
	30: {chars: []string{"v", "<", "^", ">"}},
	31: {chars: []string{">>--->", " >>--->", "  >>--->", "   >>--->", "    >>--->", "    <---<<", "   <---<<", "  <---<<", " <---<<", "<---<<"}},
	32: {chars: []string{"|", "||", "|||", "||||", "|||||", "|||||||", "||||||||", "|||||||", "||||||", "|||||", "||||", "|||", "||", "|"}},
	33: {chars: []string{"[          ]", "[=         ]", "[==        ]", "[===       ]", "[====      ]", "[=====     ]", "[======    ]", "[=======   ]", "[========  ]", "[========= ]", "[==========]"}},
	34: {chars: []string{"(*---------)", "(-*--------)", "(--*-------)", "(---*------)", "(----*-----)", "(-----*----)", "(------*---)", "(-------*--)", "(--------*-)", "(---------*)"}},
	35: {chars: []string{"█▒▒▒▒▒▒▒▒▒", "███▒▒▒▒▒▒▒", "█████▒▒▒▒▒", "███████▒▒▒", "██████████"}},
	36: {chars: []string{"[                    ]", "[=>                  ]", "[===>                ]", "[=====>              ]", "[======>             ]", "[========>           ]", "[==========>         ]", "[============>       ]", "[==============>     ]", "[================>   ]", "[==================> ]", "[===================>]"}},
	37: {chars: []string{"ဝ", "၀"}},
	38: {chars: []string{"▌", "▀", "▐▄"}},
	39: {chars: []string{"🌍", "🌎", "🌏"}},
	40: {chars: []string{"◜", "◝", "◞", "◟"}},
	41: {chars: []string{"⬒", "⬔", "⬓", "⬕"}},
	42: {chars: []string{"⬖", "⬘", "⬗", "⬙"}},
	43: {chars: []string{"[>>>          >]", "[]>>>>        []", "[]  >>>>      []", "[]    >>>>    []", "[]      >>>>  []", "[]        >>>>[]", "[>>          >>]"}},
	44: {chars: []string{"♠", "♣", "♥", "♦"}},
	45: {chars: []string{"➞", "➟", "➠", "➡", "➠", "➟"}},
	46: {chars: []string{"  |  ", ` \   `, "_    ", ` \   `, "  |  ", "   / ", "    _", "   / "}},
	47: {chars: []string{"  . . . .", ".   . . .", ". .   . .", ". . .   .", ". . . .  ", ". . . . ."}},
	48: {chars: []string{" |     ", "  /    ", "   _   ", `    \  `, "     | ", `    \  `, "   _   ", "  /    "}},
	49: {chars: []string{"⎺", "⎻", "⎼", "⎽", "⎼", "⎻"}},
	50: {chars: []string{"▹▹▹▹▹", "▸▹▹▹▹", "▹▸▹▹▹", "▹▹▸▹▹", "▹▹▹▸▹", "▹▹▹▹▸"}},
	51: {chars: []string{"[    ]", "[   =]", "[  ==]", "[ ===]", "[====]", "[=== ]", "[==  ]", "[=   ]"}},
	52: {chars: []string{"( ●    )", "(  ●   )", "(   ●  )", "(    ● )", "(     ●)", "(    ● )", "(   ●  )", "(  ●   )", "( ●    )"}},
	53: {chars: []string{"✶", "✸", "✹", "✺", "✹", "✷"}},
	54: {chars: []string{"▐|\\____________▌", "▐_|\\___________▌", "▐__|\\__________▌", "▐___|\\_________▌", "▐____|\\________▌", "▐_____|\\_______▌", "▐______|\\______▌", "▐_______|\\_____▌", "▐________|\\____▌", "▐_________|\\___▌", "▐__________|\\__▌", "▐___________|\\_▌", "▐____________|\\▌", "▐____________/|▌", "▐___________/|_▌", "▐__________/|__▌", "▐_________/|___▌", "▐________/|____▌", "▐_______/|_____▌", "▐______/|______▌", "▐_____/|_______▌", "▐____/|________▌", "▐___/|_________▌", "▐__/|__________▌", "▐_/|___________▌", "▐/|____________▌"}},
	55: {chars: []string{"▐⠂       ▌", "▐⠈       ▌", "▐ ⠂      ▌", "▐ ⠠      ▌", "▐  ⡀     ▌", "▐  ⠠     ▌", "▐   ⠂    ▌", "▐   ⠈    ▌", "▐    ⠂   ▌", "▐    ⠠   ▌", "▐     ⡀  ▌", "▐     ⠠  ▌", "▐      ⠂ ▌", "▐      ⠈ ▌", "▐       ⠂▌", "▐       ⠠▌", "▐       ⡀▌", "▐      ⠠ ▌", "▐      ⠂ ▌", "▐     ⠈  ▌", "▐     ⠂  ▌", "▐    ⠠   ▌", "▐    ⡀   ▌", "▐   ⠠    ▌", "▐   ⠂    ▌", "▐  ⠈     ▌", "▐  ⠂     ▌", "▐ ⠠      ▌", "▐ ⡀      ▌", "▐⠠       ▌"}},
	56: {chars: []string{"¿", "?"}},
	57: {chars: []string{"⢹", "⢺", "⢼", "⣸", "⣇", "⡧", "⡗", "⡏"}},
	58: {chars: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"}},
	59: {chars: []string{".  ", ".. ", "...", " ..", "  .", "   "}},
	60: {chars: []string{".", "o", "O", "°", "O", "o", "."}},
	61: {chars: []string{"▓", "▒", "░"}},
	62: {chars: []string{"▌", "▀", "▐", "▄"}},
	63: {chars: []string{"⊶", "⊷"}},
	64: {chars: []string{"▪", "▫"}},
	65: {chars: []string{"□", "■"}},
	66: {chars: []string{"▮", "▯"}},
	67: {chars: []string{"-", "=", "≡"}},
	68: {chars: []string{"d", "q", "p", "b"}},
	69: {chars: []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"}},
	70: {chars: []string{"🌑 ", "🌒 ", "🌓 ", "🌔 ", "🌕 ", "🌖 ", "🌗 ", "🌘 "}},
	71: {chars: []string{"☗", "☖"}},
	72: {chars: []string{"⧇", "⧆"}},
	73: {chars: []string{"◉", "◎"}},
	74: {chars: []string{"㊂", "㊀", "㊁"}},
	75: {chars: []string{"⦾", "⦿"}},

	// https://github.com/schollz/progressbar/blob/master/spinners.go
}

type spinner struct {
	tool.ColorTranslator
	onDraw           func(pb *pbar)
	tmpl             *template.Template
	indentL          string
	prepend          string
	append           string
	schema           string
	chars            []string
	barWidth         int
	safetyTailSpaces int
	gauge            int32
	clrBase          int
	clrHighlight     int
}

func (s *spinner) SetBaseColor(clr int) {
	s.clrBase = clr
}

func (s *spinner) SetHighlightColor(clr int) {
	s.clrHighlight = clr
}

func (s *spinner) SetSchema(schema string) {
	s.schema = schema
	s.updateSchema()
}

func (s *spinner) SetWidth(w int) {
	s.barWidth = w
}

func (s *spinner) SetIndentChars(str string) {
	s.indentL = str
}

func (s *spinner) SetPrependText(str string) {
	s.prepend = str
}

func (s *spinner) SetAppendText(str string) {
	s.append = str
}

func (s *spinner) SetExtraTailSpaces(howMany int) {
	s.safetyTailSpaces = howMany
}

func (s *spinner) init() *spinner {
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

func (s *spinner) updateSchema() *spinner {
	if s.schema == "" {
		s.schema = defaultSchema
	}
	if s.barWidth == 0 {
		s.barWidth = barWidth
	}
	s.tmpl = template.Must(template.New("bar-build").Parse(s.schema))
	return s
}

func (s *spinner) buildBar(pb *pbar, pos, barWidth int, half bool) string {
	str := s.chars[pos]
	if len(str) < s.barWidth {
		str += strings.Repeat(" ", s.barWidth-len(str))
	}
	return str
}

func (s *spinner) String(pb *pbar) string {
	return string(s.Bytes(pb))
}

func (s *spinner) Bytes(pb *pbar) []byte {
	// defer pb.locker()()

	cnt := int(atomic.AddInt32(&s.gauge, 1)) % len(s.chars)

	var percent float64
	percent = float64(pb.read) / float64(pb.max-pb.min)
	// if percent >= 100 {
	// 	pb.completed = true
	// }

	if !pb.completed {
		pb.stopTime = time.Now()
	}
	dur := pb.stopTime.Sub(pb.startTime)

	read, suffix := humanizeBytes(float64(pb.read))
	total, suffix1 := humanizeBytes(float64(pb.max))
	speed, suffix2 := humanizeBytes(float64(pb.read) / dur.Seconds())

	data := &SchemaData{
		Indent:       s.indentL,
		Prepend:      s.prepend,
		Bar:          s.buildBar(pb, cnt, s.barWidth, false),
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

	var sb bytes.Buffer
	err := s.tmpl.Execute(&sb, data)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	if s.safetyTailSpaces > 0 {
		sb.WriteString(strings.Repeat(" ", s.safetyTailSpaces))
	}

	str := s.Translate(sb.String(), 0)
	return []byte(str)

	// if pb.completed {
	// 	return []byte(fmt.Sprintf("\r%s%s %.1fs (%v) %s Done.",
	// 		indentChars, s.chars[cnt], percent, dur, pb.title))
	// }
	//
	// str := []byte(fmt.Sprintf("\r%s%s %.1f%% %s (%v%v/%v%v, %v%v/s, %v)",
	// 	indentChars, s.chars[cnt], percent, pb.title,
	// 	read, suffix, total, suffix1, speed, suffix2, dur))
	// return str
}
