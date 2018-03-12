package worker

import (
	"testing"
)

func BenchmarkRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Run("./hello.sh")
	}
}
