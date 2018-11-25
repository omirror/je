package je

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type WriteCloseBuffer struct {
	*bufio.Writer
}

func (b WriteCloseBuffer) Close() error {
	b.Flush()
	return nil
}

type RemoteData struct {
	client *Client
}

func NewRemoteData(uri string) (Data, error) {
	return &RemoteData{
		client: NewClient(uri, nil),
	}, nil
}

func (d *RemoteData) Read(id ID, dtype DataType) (io.ReadCloser, error) {
	switch dtype {
	case DATA_INPUT:
		return d.client.Read(id.String())
	case DATA_OUTPUT:
		return d.client.Output(id.String(), false)
	case DATA_LOGS:
		return d.client.Logs(id.String(), false)
	default:
		return nil, fmt.Errorf("unsupported dtype %d", dtype)
	}
}

func (d *RemoteData) Write(id ID, dtype DataType) (io.WriteCloser, error) {
	r := bytes.NewBuffer([]byte{})
	w := WriteCloseBuffer{bufio.NewWriter(r)}

	go func() {
		err := d.client.Write(id.String(), dtype.String(), r)
		if err != nil {
			log.Errorf("error writing data to #%d: %s", id, err)
		}
	}()

	return w, nil
}

func (d *RemoteData) Tail(id ID, dtype DataType, ctx context.Context) (lines chan string, errors chan error) {
	var (
		r   io.ReadCloser
		err error
	)

	switch dtype {
	case DATA_OUTPUT:
		r, err = d.client.Output(id.String(), true)
	case DATA_LOGS:
		r, err = d.client.Logs(id.String(), true)
	default:
		err = fmt.Errorf("unsupported dtype %d", dtype)
		log.Error(err)
		errors <- err
		close(errors)
		close(lines)
		return

	}

	scanner := bufio.NewScanner(r)
	for {
		if !scanner.Scan() {
			lines <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			err = fmt.Errorf("error scanning remote data from job #%d: %s", id, err)
			log.Error(err)
			errors <- err
			close(errors)
			close(lines)
			return
		}
	}
}
