//go:build windows

package utils

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
)

func Shellout(command string, cfg *ExecConfig, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if cfg != nil {
		if cfg.Dir != "" {
			cmd.Dir = cfg.Dir
		}
		if len(cfg.Env) > 0 {
			cmd.Env = append(os.Environ(), cfg.Env...)
		}
	}

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
