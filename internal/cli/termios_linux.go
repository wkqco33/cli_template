//go:build linux

package cli

import "syscall"

// ioctlGetTermios 터미널 여부를 조회하는 ioctl 상수 (Linux)
const ioctlGetTermios = syscall.TCGETS
