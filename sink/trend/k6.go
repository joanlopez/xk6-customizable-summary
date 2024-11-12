package trend

import (
	"math"
	"sort"
	"time"

	"go.k6.io/k6/metrics"
)

// Code copied (and slightly modified) from: https://github.com/grafana/k6/blob/master/metrics/sink.go.

// NewK6Sink makes a Trend sink with the OpenHistogram circllhist histogram.
func NewK6Sink() *K6Sink {
	return &K6Sink{}
}

// K6Sink is a sink for a Trend.
type K6Sink struct {
	values []float64
	sorted bool

	count    uint64
	min, max float64
	sum      float64
}

// IsEmpty indicates whether the TrendSink is empty.
func (t *K6Sink) IsEmpty() bool { return t.count == 0 }

// Add a single sample into the trend
func (t *K6Sink) Add(s metrics.Sample) {
	if t.count == 0 {
		t.max, t.min = s.Value, s.Value
	} else {
		if s.Value > t.max {
			t.max = s.Value
		}
		if s.Value < t.min {
			t.min = s.Value
		}
	}

	t.values = append(t.values, s.Value)
	t.sorted = false
	t.count++
	t.sum += s.Value
}

// P calculates the given percentile from sink values.
func (t *K6Sink) P(pct float64) float64 {
	switch t.count {
	case 0:
		return 0
	case 1:
		return t.values[0]
	default:
		if !t.sorted {
			sort.Float64s(t.values)
			t.sorted = true
		}

		// If percentile falls on a value in Values slice, we return that value.
		// If percentile does not fall on a value in Values slice, we calculate (linear interpolation)
		// the value that would fall at percentile, given the values above and below that percentile.
		i := pct * (float64(t.count) - 1.0)
		j := t.values[int(math.Floor(i))]
		k := t.values[int(math.Ceil(i))]
		f := i - math.Floor(i)
		return j + (k-j)*f
	}
}

// Min returns the minimum value.
func (t *K6Sink) Min() float64 {
	return t.min
}

// Max returns the maximum value.
func (t *K6Sink) Max() float64 {
	return t.max
}

// Count returns the number of recorded values.
func (t *K6Sink) Count() uint64 {
	return t.count
}

// Avg returns the average (i.e. mean) value.
func (t *K6Sink) Avg() float64 {
	if t.count > 0 {
		return t.sum / float64(t.count)
	}
	return 0
}

// Format trend and return a map
func (t *K6Sink) Format(_ time.Duration) map[string]float64 {
	return map[string]float64{
		"min":   t.Min(),
		"max":   t.Max(),
		"avg":   t.Avg(),
		"med":   t.P(0.5),
		"p(90)": t.P(0.90),
		"p(95)": t.P(0.95),
	}
}

// Merge merges two Sink instances.
func (t *K6Sink) Merge(s Sink) {
	toMerge, ok := s.(*K6Sink)
	if !ok {
		panic("trying to merge incompatible trend sinks")
	}

	for _, v := range toMerge.values {
		t.Add(metrics.Sample{Value: v})
	}
}

// We want to make sure that the K6Sink
// implements the Sink interface.
var _ Sink = &K6Sink{}
