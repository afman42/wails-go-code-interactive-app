//go:build linux

package utils

import (
	"bytes"
	"os"
	"os/exec"
)

func Shellout(command string, cfg *ExecConfig, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
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
