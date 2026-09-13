//go:build !linux && !darwin && !windows

package cli

// isTerminalFD 지원하지 않는 플랫폼에서는 프롬프트를 띄우지 않는다.
func isTerminalFD(int) bool { return false }
