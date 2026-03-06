//go:build !windows

package testutil

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// processGroupAttr returns SysProcAttr that places the child in its own
// process group, allowing clean shutdown of Chrome children.
func processGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

// TerminateProcessGroup sends SIGTERM to the process group, escalating to
// SIGKILL if the process doesn't exit within the timeout.
func TerminateProcessGroup(cmd *exec.Cmd, timeout time.Duration) {
	if cmd.Process == nil {
		return
	}

	if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = cmd.Process.Signal(os.Interrupt)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
		return
	case <-time.After(timeout):
		if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
		_ = cmd.Process.Kill()
		<-done
	}
}
