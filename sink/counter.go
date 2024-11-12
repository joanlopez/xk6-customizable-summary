package sink

import "go.k6.io/k6/metrics"

// We want to make sure *CounterSink implements the Sink interface.
var _ Sink = &CounterSink{}

// CounterSink is a wrapper of metrics.CounterSink that implements the Sink interface.
type CounterSink struct {
	*metrics.CounterSink
}

// Merge merges the given sink into the current one.
// If the given sink is not a *CounterSink, it panics.
func (c *CounterSink) Merge(s Sink) {
	toMerge, ok := s.(*CounterSink)
	if !ok {
		panic("trying to merge incompatible sinks")
	}

	c.Value += toMerge.Value
	if c.First.IsZero() ||
		(!toMerge.First.IsZero() && toMerge.First.Before(c.First)) {
		c.First = toMerge.First
	}
}
