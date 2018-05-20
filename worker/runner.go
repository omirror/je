package worker

import (
	"io"
	"os/exec"
	"syscall"
)

// Run ...
func Run(binary string) (*Result, error) {
	res := NewResult()

	cmd := exec.Command(binary)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return res, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return res, err
	}

	if err = cmd.Start(); err != nil {
		return res, err
	}

	// TODO: Check written < len(stderr)
	_, err = io.Copy(res.Log, stderr)
	if err != nil {
		return res, err
	}

	// TODO: Check written < len(stdout)
	_, err = io.Copy(res.Out, stdout)
	if err != nil {
		return res, err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				res.Status = status.ExitStatus()
			}
		} else {
			return res, err
		}
	}

	return res, nil
}
