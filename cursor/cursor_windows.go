// Copyright © 2022 Atonal Authors
//

//go:build windows
// +build windows

package cursor

import (
	"syscall"
	"unsafe"
)

type (
	SHORT int16
	WORD  uint16

	SMALL_RECT struct {
		Left   SHORT
		Top    SHORT
		Right  SHORT
		Bottom SHORT
	}

	COORD struct {
		X SHORT
		Y SHORT
	}

	CONSOLE_SCREEN_BUFFER_INFO struct {
		Size              COORD
		CursorPosition    COORD
		Attributes        WORD
		Window            SMALL_RECT
		MaximumWindowSize COORD
	}
)

var (
	getConsoleScreenBufferInfoProc *syscall.LazyProc
	getConsoleCursorPositionProc   *syscall.LazyProc
	setConsoleCursorPositionProc   *syscall.LazyProc
	hideCursorProc                 *syscall.LazyProc
	showCursorProc                 *syscall.LazyProc
)

func initSyscall() {
	if getConsoleScreenBufferInfo == nil {
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		getConsoleScreenBufferInfoProc = kernel32.NewProc("GetConsoleScreenBufferInfo")
		getConsoleCursorPositionProc = kernel32.NewProc("GetConsoleCursorPosition")
		setConsoleCursorPositionProc = kernel32.NewProc("SetConsoleCursorPosition")
		hideCursorProc = kernel32.NewProc("HideCursor")
		showCursorProc = kernel32.NewProc("ShowCursor")
	}
}

// checkError evaluates the results of a Windows API call and returns the error if it failed.
func checkError(r1, r2 uintptr, err error) error {
	// Windows APIs return non-zero to indicate success
	if r1 != 0 {
		return nil
	}

	// Return the error if provided, otherwise default to EINVAL
	if err != nil {
		return err
	}
	return syscall.EINVAL
}

// coordToPointer converts a COORD into a uintptr (by fooling the type system).
func coordToPointer(c COORD) uintptr {
	// Note: This code assumes the two SHORTs are correctly laid out; the "cast" to uint32 is just to get a pointer to pass.
	return uintptr(*((*uint32)(unsafe.Pointer(&c))))
}

func getStdHandle(stdhandle int) (uintptr, error) {
	handle, err := syscall.GetStdHandle(stdhandle)
	if err != nil {
		return 0, err
	}
	return uintptr(handle), nil
}

// GetConsoleScreenBufferInfo retrieves information about the specified console screen buffer.
// See http://msdn.microsoft.com/en-us/library/windows/desktop/ms683171(v=vs.85).aspx.
func getConsoleScreenBufferInfo(handle uintptr) (info *CONSOLE_SCREEN_BUFFER_INFO, err error) {
	info = &CONSOLE_SCREEN_BUFFER_INFO{}
	err = checkError(getConsoleScreenBufferInfoProc.Call(handle, uintptr(unsafe.Pointer(info)), 0))
	return
}

// SetConsoleCursorPosition location of the console cursor.
// See https://msdn.microsoft.com/en-us/library/windows/desktop/ms686025(v=vs.85).aspx.
func setConsoleCursorPosition(handle uintptr, coord COORD) error {
	r1, r2, err := setConsoleCursorPositionProc.Call(handle, coordToPointer(coord))
	// use(coord)
	return checkError(r1, r2, err)
}

func getConsoleCursorPosition(handle uintptr) (coord COORD, err error) {
	err = checkError(getConsoleCursorPositionProc.Call(handle, coordToPointer(coord)))
	return
}

func hideCursor() (err error) {
	err = checkError(hideCursorProc.Call())
	return
}

func showCursor() (err error) {
	err = checkError(showCursorProc.Call())
	return
}

// Up moves cursor up by n
func Up(n int) {
	var err error

	var stdoutHandle uintptr
	stdoutHandle, err = getStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}

	consoleInfo, err := getConsoleScreenBufferInfo(stdoutHandle)
	if err != nil {
		return
	}

	y := consoleInfo.CursorPosition.Y - SHORT(n)
	setConsoleCursorPosition(stdoutHandle, COORD{X: consoleInfo.CursorPosition.X, Y: y})
}

// Left moves cursor left by n
func Left(n int) {
	var err error

	var stdoutHandle uintptr
	stdoutHandle, err = getStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}

	consoleInfo, err := getConsoleScreenBufferInfo(stdoutHandle)
	if err != nil {
		return
	}

	x := consoleInfo.CursorPosition.X - SHORT(n)
	setConsoleCursorPosition(stdoutHandle, COORD{X: x, Y: consoleInfo.CursorPosition.Y})
}
