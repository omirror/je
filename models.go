package je

import (
	"time"

	"git.mills.io/prologic/je/worker"
)

// Job ...
type Job struct {
	ID        int    `storm:"id,increment"`
	Name      string `storm:"index"`
	Status    int    `storm:"index"`
	Response  string
	CreatedAt time.Time `storm:"index"`
	StartedAt time.Time `storm:"index"`
	EndedAt   time.Time `storm:"index"`
}

func NewJob(name string) (job *Job, err error) {
	job = &Job{Name: name, CreatedAt: time.Now()}
	err = db.Save(job)
	return
}

func (j *Job) Start() error {
	j.StartedAt = time.Now()
	return db.Save(j)
}

func (j *Job) Finish(res *worker.Result) error {
	j.Status = res.Status()
	j.Response = string(res.Response())
	j.EndedAt = time.Now()
	return db.Save(j)
}
