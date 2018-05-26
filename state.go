package je

import (
	"strconv"
	"strings"
)

const (
	_ = iota
	STATE_CREATED
	STATE_WAITING
	STATE_RUNNING
	STATE_STOPPED
	STATE_KILLED
	STATE_ERRORED
)

// State ...
type State int

func ParseState(s string) State {
	switch strings.ToLower(s) {
	case "created":
		return State(STATE_CREATED)
	case "waiting":
		return State(STATE_WAITING)
	case "running":
		return State(STATE_RUNNING)
	case "stopped":
		return State(STATE_STOPPED)
	case "killed":
		return State(STATE_KILLED)
	case "errored":
		return State(STATE_ERRORED)
	default:
		i, err := strconv.Atoi(s)
		if err != nil {
			return State(0)
		}
		return State(i)
	}
}

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
	case STATE_KILLED:
		return "KILLED"
	case STATE_ERRORED:
		return "ERRORED"
	default:
		return "???"
	}
}
