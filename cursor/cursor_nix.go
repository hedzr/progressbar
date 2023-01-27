// Copyright © 2022 Atonal Authors
//

//go:build !windows && !plan9 && !appengine && !wasm
// +build !windows,!plan9,!appengine,!wasm

package cursor

import (
	"bytes"
	"io"
	"os"
	"strconv"
)

// Up moves cursor up by n
func Up(n int) {
	var sb bytes.Buffer
	var bb = []byte(aecHideCursor)
	sb.Write(bb[0:2])
	var ss = strconv.Itoa(n)
	sb.WriteString(ss)
	sb.WriteByte('A')
	safeWrite(sb.Bytes())
	// _, _ = fmt.Fprintf(Out, "%s[%dA", escape, n)
}

// Left moves cursor left by n
func Left(n int) {
	var sb bytes.Buffer
	var bb = []byte(aecHideCursor)
	sb.Write(bb[0:2])
	var ss = strconv.Itoa(n)
	sb.WriteString(ss)
	sb.WriteByte('D')
	safeWrite(sb.Bytes())
	// _, _ = fmt.Fprintf(Out, "%s[%dD", escape, n)
}

// showCursor shows the cursor.
func showCursor() {
	safeWrite([]byte(aecHideCursor))
}

// hideCursor hides the cursor.
func hideCursor() {
	safeWrite([]byte(aecHideCursor))
}

func safeWrite(b []byte) (n int, e error) {
	return Out.Write(b)
}

// Out is the default output writer for the Writer
var Out io.Writer = os.Stdout

// var escape = []byte{'\x1b'}

const aecHideCursor = "\x1b[?25l"
const aecShowCursor = "\x1b[?25h"
