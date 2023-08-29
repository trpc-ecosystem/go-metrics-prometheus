package prometheus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsCache_Loader(t *testing.T) {
	testCache := NewMetricsCache()

	testValue := 0
	testCache.Loader("test_key", func() interface{} {
		testValue++
		return testValue
	})

	assert.Equal(t, 1, testValue)
}
func BenchmarkMetricsCache_Loader(b *testing.B) {
	testCache := NewMetricsCache()

	testValue := 0
	for n := 0; n < b.N; n++ {
		testCache.Loader("test_key", func() interface{} {
			testValue++
			return testValue
		})
	}

	assert.Equal(b, 1, testValue)
}
