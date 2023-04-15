//

package memory

import (
	"math"
	"sort"
)

// Distance calculates the euclidean distance between two vectors. If the
// vectors are not the same length, the distance is meaningless and will be
// `math.MaxFloat32`.
func Distance(base []float32, other []float32) float32 {
	if len(base) != len(other) {
		return math.MaxFloat32
	}

	var sum float32

	for i := range base {
		sum += (base[i] - other[i]) * (base[i] - other[i])
	}

	return sum
}

// Closest is a struct that keeps track of the closest values.
type Closest struct {
	Base   []float32
	Count  int
	Values []struct {
		Distance float32
		Key      []float32
		Value    []byte
	}
}

// NewClosest creates a new Closest instance.
func NewClosest(base []float32, count int) *Closest {
	return &Closest{
		Base:  base,
		Count: count,

		Values: make([]struct {
			Distance float32
			Key      []float32
			Value    []byte
		}, 0, count),
	}
}

// Add adds a new value to the Closest instance.
func (closest *Closest) Add(key []float32, value []byte) {
	entry := struct {
		Distance float32
		Key      []float32
		Value    []byte
	}{
		Distance: Distance(closest.Base, key),
		Key:      key,
		Value:    value,
	}

	if len(closest.Values) < closest.Count {
		closest.Values = append(closest.Values, entry)

		return
	}

	worst := -1

	for i := range closest.Values {
		if closest.Values[i].Distance > entry.Distance {
			worst = i
		}
	}

	if worst > -1 {
		closest.Values[worst] = entry
	}
}

// Strings returns the closest values as strings.
func (closest *Closest) Strings() []string {
	sort.Slice(
		closest.Values,
		func(i, j int) bool {
			return closest.Values[i].Distance < closest.Values[j].Distance
		},
	)

	values := make([]string, 0, len(closest.Values))
	for _, value := range closest.Values {
		values = append(values, string(value.Value))
	}

	return values
}
