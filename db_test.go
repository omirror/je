package je

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDB_Valid(t *testing.T) {
	assert := assert.New(t)

	_, err := InitDB("memory://")
	assert.NoError(err)
	//assert.Implements((*Store)(nil), store)
}
