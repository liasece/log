// +build !appengine,!js,windows

package encoder

import (
	"io"
	"os"
	"syscall"

	sequences "github.com/konsorten/go-windows-terminal-sequences"
)

func initTerminal(w io.Writer) error {
	switch v := w.(type) {
	case *os.File:
		return sequences.EnableVirtualTerminalProcessing(syscall.Handle(v.Fd()), true)
	}
	return nil
}

// CheckIfTerminal check the terminal for suport unix type consloe color
func CheckIfTerminal(w io.Writer) bool {
	var ret bool
	switch v := w.(type) {
	case *os.File:
		var mode uint32
		err := syscall.GetConsoleMode(syscall.Handle(v.Fd()), &mode)
		ret = (err == nil)
	default:
		ret = false
	}
	if ret {
		err := initTerminal(w)
		return err == nil
	}
	return ret
}
