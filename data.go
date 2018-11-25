package je

import (
	"context"
	"io"
)

const (
	DATA_INPUT DataType = iota
	DATA_OUTPUT
	DATA_LOGS
)

type DataType int

func (dt DataType) String() string {
	switch dt {
	case 0:
		return "in"
	case 1:
		return "out"
	case 2:
		return "log"
	default:
		return "???"
	}
}

type Data interface {
	Read(id ID, dtype DataType) (io.ReadCloser, error)
	Write(id ID, dtype DataType) (io.WriteCloser, error)
	Tail(id ID, dtype DataType, ctx context.Context) (chan string, chan error)
}
