package je

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// Job ...
type Job struct {
	sync.RWMutex

	ID          ID        `json:"id"`
	Name        string    `json:"name"`
	Args        []string  `json:"args"`
	Interactive bool      `json:"interactive"`
	Worker      string    `json:"worker"`
	State       State     `json:"state"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created"`
	StartedAt   time.Time `json:"started"`
	StoppedAt   time.Time `json:"stopped"`
	KilledAt    time.Time `json:"killed"`
	ErroredAt   time.Time `json:"errored"`

	input io.WriteCloser
	cmd   *exec.Cmd
	done  chan bool
}

func NewJob(name string, args []string, interactive bool) (job *Job, err error) {
	job = &Job{
		ID:          db.NextId(),
		Name:        name,
		Args:        args,
		Interactive: interactive,
		CreatedAt:   time.Now(),

		done: make(chan bool, 1),
	}
	err = db.Save(job)
	if err == nil {
		metrics.Counter("job", "count").Inc()
		metrics.GaugeVec("job", "stats").WithLabelValues(job.ID.String(), job.Name, "created").Inc()
	}
	return
}

func (j *Job) Id() ID {
	j.RLock()
	defer j.RUnlock()
	return j.ID
}

func (j *Job) Enqueue() error {
	j.Lock()
	defer j.Unlock()
	j.State = STATE_WAITING
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "created").Dec()
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "waiting").Inc()
	return db.Save(j)
}

func (j *Job) Start(worker string) error {
	j.Lock()
	defer j.Unlock()
	j.Worker = worker
	j.State = STATE_RUNNING
	j.StartedAt = time.Now()
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "waiting").Dec()
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "running").Inc()
	return db.Save(j)
}

func (j *Job) Kill(force bool) (err error) {
	j.Lock()
	defer j.Unlock()
	if force {
		err = j.cmd.Process.Kill()
		if err != nil {
			log.Errorf("error killing job #%d: %s", j.ID, err)
			return
		}

		j.done <- true
		j.State = STATE_KILLED
		j.KilledAt = time.Now()
		metrics.SummaryVec("job", "duration").WithLabelValues(j.Name).Observe(j.KilledAt.Sub(j.StartedAt).Seconds())
		metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "running").Dec()
		metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "killed").Inc()
		return db.Save(j)
	}
	return j.cmd.Process.Signal(os.Interrupt)
}

func (j *Job) Stop() error {
	j.Lock()
	defer j.Unlock()
	j.done <- true
	j.State = STATE_STOPPED
	j.StoppedAt = time.Now()
	metrics.SummaryVec("job", "duration").WithLabelValues(j.Name).Observe(j.StoppedAt.Sub(j.StartedAt).Seconds())
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "running").Dec()
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "stopped").Inc()
	return db.Save(j)
}

func (j *Job) Error(err error) error {
	j.Lock()
	defer j.Unlock()
	j.State = STATE_ERRORED
	j.ErroredAt = time.Now()
	metrics.SummaryVec("job", "duration").WithLabelValues(j.Name).Observe(j.ErroredAt.Sub(j.StartedAt).Seconds())
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "running").Dec()
	metrics.GaugeVec("job", "stats").WithLabelValues(j.ID.String(), j.Name, "errored").Inc()
	return db.Save(j)
}

func (j *Job) Wait() {
	<-j.done
}

func (j *Job) Close() error {
	if !j.Interactive {
		return fmt.Errorf("cannot write to a non-interactive job")
	}

	return j.input.Close()
}

func (j *Job) Write(input io.Reader) (int64, error) {
	if !j.Interactive {
		return 0, fmt.Errorf("cannot write to a non-interactive job")
	}

	return io.Copy(j.input, input)
}

func (j *Job) Killed() bool {
	j.RLock()
	defer j.RUnlock()
	return j.State == STATE_KILLED
}

func (j *Job) Execute() (err error) {
	cmd := exec.Command(j.Name, j.Args...)

	if j.Interactive {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Errorf("error creating input for job #%d: %s", j.ID, err)
			return err
		}
		defer stdin.Close()
		j.Lock()
		j.input = stdin
		j.Unlock()
	} else {
		stdin, err := data.Read(j.ID, DATA_INPUT)
		if err != nil {
			log.Errorf("error reading input for job #%d: %s", j.ID, err)
			return err
		}
		defer stdin.Close()
		cmd.Stdin = stdin
	}

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

	logs, err := data.Write(j.ID, DATA_LOGS)
	if err != nil {
		log.Errorf("error creating logs for job #%s: %s", j.ID, err)
		return err
	}
	// TODO: Check for errors? Retry RINTR?
	defer logs.Close()

	output, err := data.Write(j.ID, DATA_OUTPUT)
	if err != nil {
		log.Errorf("error creating output for job #%s: %s", j.ID, err)
		return err
	}
	// TODO: Check for errors? Retry RINTR?
	defer output.Close()

	if err = cmd.Start(); err != nil {
		log.Errorf("error starting job #%d: %s", j.ID, err)
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := io.Copy(logs, stderr)
		log.Debugf("written %d bytes of logs for job #%d", n, j.ID)
		if err != nil {
			log.Errorf("error writing logs for job #%d: %s", j.ID, err)
		}
		err = stderr.Close()
		if err != nil {
			log.Errorf("error closing stderr for job #%d: %s", j.ID, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := io.Copy(output, stdout)
		log.Debugf("written %d bytes of output for job #%d", n, j.ID)
		if err != nil {
			log.Errorf("error writing output for job #%d: %s", j.ID, err)
		}
		err = stdout.Close()
		if err != nil {
			log.Errorf("error closing stdout for job #%d: %s", j.ID, err)
		}
	}()

	// Wait for all io.Copy()s to complete
	wg.Wait()

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
