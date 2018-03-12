package worker

import (
	"io/ioutil"
	"os/exec"
	"syscall"
)

// Result ...
type Result struct {
	status   int
	logs     []byte
	response []byte
}

// String ...
func (res *Result) String() string {
	return string(res.logs[:])
}

// Status ...
func (res *Result) Status() int {
	return res.status
}

// Logs ...
func (res *Result) Logs() []byte {
	return res.logs
}

// Response ...
func (res *Result) Response() []byte {
	return res.response
}

// Run ...
func Run(binary string) (res *Result, err error) {
	res = &Result{}

	cmd := exec.Command(binary)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return res, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return res, err
	}

	if err := cmd.Start(); err != nil {
		return res, err
	}

	logs, err := ioutil.ReadAll(stderr)
	if err != nil {
		return res, err
	}

	//res.logs = make([]byte, len(logs))
	res.logs = logs[:]

	response, err := ioutil.ReadAll(stdout)
	if err != nil {
		return res, err
	}
	//res.response = make([]byte, len(response))
	res.response = response[:]

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				res.status = status.ExitStatus()
			}
		} else {
			return res, err
		}
	}

	return res, nil
}
