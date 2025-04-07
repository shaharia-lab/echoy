//go:build !windows
// +build !windows

package daemon

import (
	"os/exec"
	"syscall"
)

// setPlatformProcAttr sets platform-specific process attributes for Unix-like systems
func setPlatformProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}
