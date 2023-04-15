//

package memory_test

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/lazygpt/lazygpt/plugin/local/pkg/memory"
)

func TestDistance(t *testing.T) {
	t.Parallel()

	base := []float32{1, 2, 3, 4, 5}

	assert.Equal(t, memory.Distance(base, []float32{1, 2, 3, 4, 5}), float32(0))
	assert.Equal(t, memory.Distance(base, []float32{2, 3, 4, 5, 6}), float32(5))
	assert.Equal(t, memory.Distance(base, []float32{0, 1, 2, 3, 4}), float32(5))
	assert.Equal(t, memory.Distance(base, []float32{5, 4, 3, 2, 1}), float32(40))
	assert.Equal(t, memory.Distance(base, []float32{6, 5, 4, 3, 2}), float32(45))
	assert.Equal(t, memory.Distance(base, []float32{}), float32(math.MaxFloat32))
}

func TestClosest(t *testing.T) {
	t.Parallel()

	closest := memory.NewClosest([]float32{1, 2, 3, 4, 5}, 2)

	closest.Add([]float32{5, 4, 3, 2, 1}, []byte("a"))
	closest.Add([]float32{2, 3, 4, 5, 6}, []byte("b"))
	closest.Add([]float32{6, 5, 4, 3, 2}, []byte("c"))
	closest.Add([]float32{}, []byte("d"))

	values := closest.Strings()
	assert.DeepEqual(t, values, []string{"b", "a"})
}
