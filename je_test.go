package je

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	datadir, err := ioutil.TempDir("", "jetest")
	if err != nil {
		log.Errorf("error creating test datadir: %s", err)
		os.Exit(1)
	}
	defer os.RemoveAll(datadir)

	_, err = InitData(datadir)
	if err != nil {
		log.Errorf("error initializing data: %s", err)
		os.Exit(1)
	}

	store, err := InitStore("memory://")
	if err != nil {
		log.Errorf("error initializing database: %s", err)
		os.Exit(1)
	}
	defer store.Close()

	InitMetrics("jetest")

	go NewServer(":8000", nil).ListenAndServe()

	os.Exit(m.Run())
}

func TestRun(t *testing.T) {
	assert := assert.New(t)

	_, err := http.Post("http://127.0.0.1:8000/create/samples/hello.sh", "text/plain", nil)
	assert.NoError(err)
}
