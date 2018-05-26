package je

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
