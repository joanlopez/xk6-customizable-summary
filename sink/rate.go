package sink

import "go.k6.io/k6/metrics"

// We want to make sure *RateSink implements the Sink interface.
var _ Sink = &RateSink{}

// RateSink is a wrapper of metrics.RateSink that implements the Sink interface.
type RateSink struct {
	*metrics.RateSink
}

// Merge merges the given sink into the current one.
// If the given sink is not a *RateSink, it panics.
func (r *RateSink) Merge(s Sink) {
	toMerge, ok := s.(*RateSink)
	if !ok {
		panic("trying to merge incompatible sinks")
	}

	r.Total += toMerge.Total
	r.Trues += toMerge.Trues
}
