// Copyright © 2022 Atonal Authors
//

//go:build !windows && !plan9 && !appengine && !wasm
// +build !windows,!plan9,!appengine,!wasm

package cursor

import (
	"io"
	"os"
	"strconv"
)

// Up moves cursor up by n
func Up(n int) {
	var bb = []byte(aecHideCursor)
	safeWrite(bb[0:2])
	var ss = strconv.Itoa(n)
	safeWrite([]byte(ss))
	var A = []byte{'A'}
	safeWrite(A)
	// _, _ = fmt.Fprintf(Out, "%s[%dA", escape, n)
}

// Left moves cursor left by n
func Left(n int) {
	var bb = []byte(aecHideCursor)
	safeWrite(bb[0:2])
	var ss = strconv.Itoa(n)
	safeWrite([]byte(ss))
	var D = []byte{'D'}
	safeWrite(D)
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

var escape = []byte{'\x1b'}

const aecHideCursor = "\x1b[?25l"
const aecShowCursor = "\x1b[?25h"
