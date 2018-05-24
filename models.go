package je

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	_ = iota
	STATE_CREATED
	STATE_WAITING
	STATE_RUNNING
	STATE_STOPPED
	STATE_ERRORED
)

// State ...
type State int

func (s State) String() string {
	switch s {
	case STATE_CREATED:
		return "CREATED"
	case STATE_WAITING:
		return "WAITING"
	case STATE_RUNNING:
		return "RUNNING"
	case STATE_STOPPED:
		return "STOPPED"
	case STATE_ERRORED:
		return "ERRORED"
	default:
		return "???"
	}
}

// Job ...
type Job struct {
	ID        int      `storm:"id,increment"`
	Name      string   `storm:"index"`
	Args      []string `storm:"index"`
	State     State    `storm:"index"`
	Status    int      `storm:"index"`
	Input     string
	CreatedAt time.Time `storm:"index"`
	StartedAt time.Time `storm:"index"`
	StoppedAt time.Time `storm:"index"`
	ErroredAt time.Time `storm:"index"`

	done chan bool
}

func NewJob(name string, args []string, input io.Reader) (job *Job, err error) {
	inputBytes, err := ioutil.ReadAll(input)
	if err != nil {
		log.Errorf("error reading input for new job: %s", err)
		return job, nil
	}

	job = &Job{
		Name:      name,
		Args:      args,
		Input:     string(inputBytes),
		CreatedAt: time.Now(),

		done: make(chan bool, 1),
	}
	err = db.Save(job)
	return
}

func (j *Job) Enqueue() error {
	j.State = STATE_WAITING
	return db.Save(j)
}

func (j *Job) Start() error {
	j.State = STATE_RUNNING
	j.StartedAt = time.Now()
	return db.Save(j)
}

func (j *Job) Stop() error {
	j.done <- true
	j.State = STATE_STOPPED
	j.StoppedAt = time.Now()
	return db.Save(j)
}

func (j *Job) Error(err error) error {
	j.State = STATE_ERRORED
	j.ErroredAt = time.Now()
	return db.Save(j)
}

func (j *Job) Wait() {
	<-j.done
}

func (j *Job) Logs() (io.ReadCloser, error) {
	return os.Open(fmt.Sprintf("%d.log", j.ID))
}

func (j *Job) Output() (io.ReadCloser, error) {
	return os.Open(fmt.Sprintf("%d.out", j.ID))
}

func (j *Job) Execute() error {
	j.Start()
	defer j.Stop()

	cmd := exec.Command(j.Name, j.Args...)
	cmd.Stdin = bytes.NewBufferString(j.Input)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("error reading logs from job #%d: %s", j.ID, err)
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("error reading output from job #%d: %s", j.ID, err)
		return err
	}

	if err = cmd.Start(); err != nil {
		log.Errorf("error starting job #%d: %s", j.ID, err)
		return err
	}

	logf, err := os.OpenFile(fmt.Sprintf("%d.log", j.ID), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf("error creating logfile for job #%s: %s", j.ID, err)
		return err
	}

	outf, err := os.OpenFile(fmt.Sprintf("%d.out", j.ID), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf("error creating output file for job #%s: %s", j.ID, err)
		return err
	}

	// TODO: Check if written < len(res.Log)?
	go func() {
		defer logf.Close()
		_, err = io.Copy(logf, stderr)
	}()

	go func() {
		defer outf.Close()
		_, err = io.Copy(outf, stdout)
	}()

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				j.Status = status.ExitStatus()
			}
		}
	}

	return nil
}
