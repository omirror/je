package je

import (
	"net/http"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	db, err := InitDB("memory://")
	if err != nil {
		log.Errorf("error initializing database: %s", err)
		os.Exit(1)
	}
	defer db.Close()

	go NewServer(":8000", nil).ListenAndServe()

	os.Exit(m.Run())
}

func TestRun(t *testing.T) {
	assert := assert.New(t)

	_, err := http.Post("http://127.0.0.1:8000/create/samples/hello.sh", "text/plain", nil)
	assert.NoError(err)
}
