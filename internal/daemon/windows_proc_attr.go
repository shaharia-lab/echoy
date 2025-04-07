//go:build windows
// +build windows

package daemon

import (
	"os/exec"
	"syscall"
)

// setPlatformProcAttr sets platform-specific process attributes for Windows
func setPlatformProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
}
