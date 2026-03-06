//go:build windows

package testutil

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

func processGroupAttr() *syscall.SysProcAttr {
	// CREATE_NEW_PROCESS_GROUP so we can kill the tree on shutdown.
	return &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// TerminateProcessGroup kills the process and its children via taskkill /T.
// Falls back to Process.Kill if taskkill isn't available.
func TerminateProcessGroup(cmd *exec.Cmd, timeout time.Duration) {
	if cmd.Process == nil {
		return
	}

	// taskkill /T /F /PID kills the process tree
	kill := exec.Command("taskkill", "/T", "/F", "/PID", // #nosec G204 -- PID is from our own cmd.Process
		fmt.Sprintf("%d", cmd.Process.Pid))
	if err := kill.Run(); err != nil {
		_ = cmd.Process.Kill()
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
	}
}
