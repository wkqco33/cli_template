//go:build linux || darwin

package cli

import (
	"syscall"
	"unsafe"
)

// isTerminalFD fd가 TTY인지 확인한다. /dev/null 같은 character device를
// TTY로 잘못 판단하지 않도록 termios 조회 결과를 사용한다.
func isTerminalFD(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(ioctlGetTermios),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
	return err == 0
}
