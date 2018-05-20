package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	tmpfile, err := ioutil.TempFile("", "je")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	db, err = storm.Open(tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	os.Exit(m.Run())
}

func TestZeroJob(t *testing.T) {
	assert := assert.New(t)

	j := Job{}
	assert.Equal(j.ID, 0)
	assert.Equal(j.Name, "")
	assert.Equal(j.CreatedAt, time.Time{})
	assert.Equal(j.StartedAt, time.Time{})
	assert.Equal(j.EndedAt, time.Time{})
}

func TestNewJob(t *testing.T) {
	assert := assert.New(t)

	j, err := NewJob("foo")
	assert.Nil(err, nil)

	assert.Equal(j.ID, 1)
	assert.Equal(j.Name, "foo")
}
