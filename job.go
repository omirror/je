package je

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hpcloud/tail"
)

// Job ...
type Job struct {
	sync.RWMutex

	ID        uint64    `storm:"id,increment"`
	Name      string    `storm:"index"`
	Args      []string  `storm:"index"`
	Worker    string    `storm:"index"`
	State     State     `storm:"index"`
	Status    int       `storm:"index"`
	CreatedAt time.Time `storm:"index"`
	StartedAt time.Time `storm:"index"`
	StoppedAt time.Time `storm:"index"`
	KilledAt  time.Time `storm:"index"`
	ErroredAt time.Time `storm:"index"`

	cmd  *exec.Cmd
	done chan bool
}

func NewJob(name string, args []string) (job *Job, err error) {
	job = &Job{
		Name:      name,
		Args:      args,
		CreatedAt: time.Now(),

		done: make(chan bool, 1),
	}
	err = db.Save(job)
	return
}

func (j *Job) Id() uint64 {
	return j.ID
}

func (j *Job) Enqueue() error {
	j.State = STATE_WAITING
	return db.Save(j)
}

func (j *Job) Start(worker string) error {
	j.Worker = worker
	j.State = STATE_RUNNING
	j.StartedAt = time.Now()
	return db.Save(j)
}

func (j *Job) Kill(force bool) (err error) {
	if force {
		err = j.cmd.Process.Kill()
		if err != nil {
			log.Errorf("error killing job #%d: %s", j.ID, err)
			return
		}

		j.done <- true
		j.State = STATE_KILLED
		j.KilledAt = time.Now()
		return db.Save(j)
	}
	return j.cmd.Process.Signal(os.Interrupt)
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

func (j *Job) Input() (io.Reader, error) {
	return os.Open(fmt.Sprintf("%d.in", j.ID))
}

func (j *Job) SetInput(input io.Reader) error {
	inf, err := os.OpenFile(fmt.Sprintf("%d.in", j.ID), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf("error creating input file for job #%s: %s", j.ID, err)
		return err
	}
	defer inf.Close()

	io.Copy(inf, input)

	return nil
}

func (j *Job) Logs() (io.ReadCloser, error) {
	return os.Open(fmt.Sprintf("%d.log", j.ID))
}

func (j *Job) LogsTail(ctx context.Context) (lines chan string, errors chan error) {
	lines = make(chan string)
	errors = make(chan error)

	t, err := tail.TailFile(
		fmt.Sprintf("%d.log", j.ID),
		tail.Config{Follow: true},
	)
	if err != nil {
		log.Errorf("error tailing output for job #%d: %s", err)
		errors <- err
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line := <-t.Lines:
				if line.Err != nil {
					errors <- line.Err
				} else {
					lines <- line.Text
				}
			}
		}
	}()
	return
}

func (j *Job) Output() (io.ReadCloser, error) {
	return os.Open(fmt.Sprintf("%d.out", j.ID))
}

func (j *Job) OutputTail(ctx context.Context) (lines chan string, errors chan error) {
	lines = make(chan string)
	errors = make(chan error)

	t, err := tail.TailFile(
		fmt.Sprintf("%d.out", j.ID),
		tail.Config{Follow: true},
	)
	if err != nil {
		log.Errorf("error tailing output for job #%d: %s", err)
		errors <- err
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line := <-t.Lines:
				if line.Err != nil {
					errors <- line.Err
				} else {
					lines <- line.Text
				}
			}
		}
	}()
	return
}

func (j *Job) Execute() error {
	stdin, err := j.Input()
	if err != nil {
		log.Errorf("error reading input for job #%d: %s", j.ID, err)
		return err
	}

	cmd := exec.Command(j.Name, j.Args...)
	cmd.Stdin = stdin

	j.Lock()
	j.cmd = cmd
	j.Unlock()

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
