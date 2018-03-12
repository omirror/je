package types

import (
	"time"
)

type Job struct {
	id        uint64
	name      string
	tags      map[string]string
	state     JobState
	segment   JobSegment
	creator   uint64
	createdAt time.Time
	runAt     time.Time
}

func NewJob(name string, creator uint64, segment JobSegment) (Job, error) {
	// TODO: Autoincrement and populate id
	return Job{
		id:        0,
		name:      name,
		state:     SCHEDULED,
		segment:   segment,
		creator:   creator,
		createdAt: time.Time(),
	}
}

func (job Job) SetTags(tags map[string]string) error {
	job.tags = tags[:]
}

type JobState uint8

// JobStates
const (
	SCHEDULED JobState = iota // scheduled for a future time
	WAITING                   // waiting to be picked up by a worker
	RUNNING                   // running and being worked on by a worker
	CANCELLED                 // cancelled
	COMPLETED                 // completed successfully
	FAILED                    // completed unsuccessfully with failure
	KILLED                    // forcibly killed
	PAUSED                    // paused
	ERROR                     // completedd unsuccssfully with unhandled errors
	UNKNOWN                   // unknown state due to service errors
)

func (state JobState) String() string {
	switch state {
	case SCHEDULED:
		return "Scheduled"
	case WAITING:
		return "Waiting"
	case RUNNING:
		return "Running"
	case CANCELLED:
		return "Cancelled"
	case COMPLETED:
		return "Completed"
	case FAILED:
		return "Failed"
	case KILLED:
		return "Killed"
	case PAUSED:
		return "Paused"
	case ERROR:
		return "Error"
	case UNKNOWN:
		return "Unknown"
	}
	return ""
}

type JobSegment uint8

// JobSegment
const (
	DefaultSegment JobSegment = iota
	TestSegment
	DevSegment
)

func (segment JobSegment) String() string {
	switch segment {
	case DefaultSegment:
		return "Default"
	case TestSegment:
		return "Test"
	case DevSegment:
		return "Dev"
	}
	return ""
}
