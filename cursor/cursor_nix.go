// Copyright © 2022 Atonal Authors
//

//go:build !windows && !plan9 && !appengine && !wasm
// +build !windows,!plan9,!appengine,!wasm

package cursor

import (
	"fmt"
	"io"
	"os"
)

// Out is the default output writer for the Writer
var Out = io.Writer(os.Stdout)

// Up moves cursor up by n
func Up(n int) {
	_, _ = fmt.Fprintf(Out, "%s[%dA", escape, n)
}

// Left moves cursor left by n
func Left(n int) {
	_, _ = fmt.Fprintf(Out, "%s[%dD", escape, n)
}

// showCursor shows the cursor.
func showCursor() {
	_, _ = fmt.Fprintf(Out, "%s[?25h", escape)
}

// hideCursor hides the cursor.
func hideCursor() {
	_, _ = fmt.Fprintf(Out, "%s[?25l", escape)
}

const escape = "\x1b"
