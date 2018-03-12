package types

import (
	"time"
)

type Message struct {
	Id         uint64
	Payload    []byte
	EnqueuedAt time.Time
	Retries    uint64
	RetriedAt  time.Time
	Acked      bool
	AckedAt    time.Time
}

func NewMessage(id uint64, payload []byte) (Message, error) {
	// TODO: Autoincrement and populate id
	return Message{
		Id:      id,
		Payload: payload[:],
	}, nil
}
