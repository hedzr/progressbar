// Copyright © 2022 Atonal Authors
//

package tool

import (
	"regexp"
	"strings"
)

// IsTtyEscaped detects a string if it contains ansi color escaped sequences
func IsTtyEscaped(s string) bool { return isTtyEscaped(s) }
func isTtyEscaped(s string) bool { return strings.Contains(s, "\x1b[") || strings.Contains(s, "\x9b[") }

// StripEscapes removes any ansi color escaped sequences from a string
func StripEscapes(str string) (strCleaned string) { return stripEscapes(str) }

// var reStripEscapesOld = regexp.MustCompile(`\x1b\[[0-9,;]+m`)

const ansi = "[\u001b\u009b][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var reStripEscapes = regexp.MustCompile(ansi)

func stripEscapes(str string) (strCleaned string) {
	strCleaned = reStripEscapes.ReplaceAllString(str, "")
	return
}

// // TrimQuotes strips first and last quote char (double quote or single quote).
// func TrimQuotes(s string) string { return exec.TrimQuotes(s) }

// StripQuotes strips single or double quotes around a string.
func StripQuotes(s string) string {
	return trimQuotes(s)
}

func trimQuotes(s string) string {
	switch {
	case s[0] == '\'':
		if s[len(s)-1] == '\'' {
			return s[1 : len(s)-1]
		}
		return s[1:]
	case s[0] == '"':
		if s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s[1:]
	case s[len(s)-1] == '\'':
		return s[0 : len(s)-1]
	case s[len(s)-1] == '"':
		return s[0 : len(s)-1]
	}
	return s
}

// func eraseMultiWSs(s string) string {
// 	return reSimp.ReplaceAllString(s, " ")
// }

// EraseAnyWSs eats any whitespaces inside the giving string s.
func EraseAnyWSs(s string) string {
	return reSimpSimp.ReplaceAllString(s, "")
}

// var reSimp = regexp.MustCompile(`[ \t][ \t]+`)
var reSimpSimp = regexp.MustCompile(`[ \t]+`)
