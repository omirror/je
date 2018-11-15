package je

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/je/codec"
	"github.com/prologic/je/codec/json"
	"github.com/prologic/msgbus/client"
)

type MessageBusQueue struct {
	client *client.Client
	codec  codec.MarshalUnmarshaler
}

func NewMessageBusQueue(uri string) (*MessageBusQueue, error) {
	client := client.NewClient(uri, nil)
	return &MessageBusQueue{
		client: client,
		codec:  json.Codec,
	}, nil
}

func (q *MessageBusQueue) Publish(job *Job) error {
	topic := strings.Trim(job.Type, "./")
	message, err := q.codec.Marshal(job)
	if err != nil {
		log.Errorf("error marshalling job %d: %s", job.ID, err)
		return err
	}

	log.Debugf("publishing to %s", topic)
	return q.client.Publish(topic, string(message))
}

func (q *MessageBusQueue) Subscribe(job *Job) error {
	return nil
}

func (q *MessageBusQueue) Close() error {
	return q.Close()
}
