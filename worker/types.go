package worker

import (
	"bytes"
)

// Result holds the status, output and log of a job
type Result struct {
	Status int
	Log    *bytes.Buffer
	Out    *bytes.Buffer
}

// NewResult returns a new instance of Result with Log and Out buffers
// pre-initialized ready for use
func NewResult() *Result {
	return &Result{
		Status: -1,
		Log:    new(bytes.Buffer),
		Out:    new(bytes.Buffer),
	}
}
