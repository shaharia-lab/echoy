//go:build !windows
// +build !windows

package daemon

import (
	"fmt"
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

// setSocketUmask sets the umask for socket creation on Unix systems and returns
// a function to restore the original umask
func setSocketUmask(d *Daemon) func() {
	oldMask := syscall.Umask(0o002)
	d.logger.Debug("Set umask", "new_mask", fmt.Sprintf("%04o", 0o002), "old_mask", fmt.Sprintf("%04o", oldMask))

	return func() {
		syscall.Umask(oldMask)
		d.logger.Debug("Restored umask", "old_mask", fmt.Sprintf("%04o", oldMask))
	}
}
