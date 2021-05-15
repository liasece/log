// +build !windows

package encoder

import (
	"io"
)

// CheckIfTerminal check the terminal for suport unix type consloe color
func CheckIfTerminal(w io.Writer) bool {
	return false
}
