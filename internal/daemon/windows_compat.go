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

// setSocketUmask is a no-op on Windows as umask is not supported
func setSocketUmask(d *Daemon) func() {
	d.logger.Debug("Umask operations not applicable on Windows")
	return func() {}
}
