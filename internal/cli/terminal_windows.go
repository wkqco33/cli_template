//go:build windows

package cli

import "syscall"

// isTerminalFD fd가 콘솔인지 확인한다.
func isTerminalFD(fd int) bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Handle(fd), &mode) == nil
}
