// Copyright © 2022 Atonal Authors
//

//go:build plan9 || appengine || wasm
// +build plan9 appengine wasm

// Adopted from https://github.com/jessevdk/go-flags

package cursor

func hideCursor() {
}

func showCursor() {
}

// Up moves cursor up by n
func Up(n int) {
}

// Left moves cursor left by n
func Left(n int) {
}
