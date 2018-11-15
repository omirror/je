package je

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitStore_Valid(t *testing.T) {
	assert := assert.New(t)

	_, err := InitStore("memory://")
	assert.NoError(err)
	//assert.Implements((*Store)(nil), store)
}
