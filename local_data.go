package je

import (
	"context"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/hpcloud/tail"
)

type LocalData struct {
	path string
}

func NewLocalData(path string) (data Data, err error) {
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Errorf("error creating local data path: %s", path)
		return
	}

	data = &LocalData{
		path: path,
	}
	return
}

func (d *LocalData) makepath(id ID, dtype DataType) string {
	return fmt.Sprintf("%s/%d.%s", d.path, id, dtype)
}

func (d *LocalData) Read(id ID, dtype DataType) (io.ReadCloser, error) {
	return os.Open(d.makepath(id, dtype))
}

func (d *LocalData) Write(id ID, dtype DataType) (io.WriteCloser, error) {
	return os.OpenFile(d.makepath(id, dtype), os.O_RDWR|os.O_CREATE, 0644)
}

func (d *LocalData) Tail(id ID, dtype DataType, ctx context.Context) (lines chan string, errors chan error) {
	lines = make(chan string)
	errors = make(chan error)

	t, err := tail.TailFile(
		d.makepath(id, dtype),
		tail.Config{Follow: true},
	)
	if err != nil {
		log.Errorf("error tailing output for job #%d: %s", id, err)
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
