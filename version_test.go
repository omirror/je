package je

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullVersion(t *testing.T) {
	actual := FullVersion()
	expected := fmt.Sprintf("%s-%s@%s", Version, Build, GitCommit)
	assert.Equal(t, expected, actual)
}
