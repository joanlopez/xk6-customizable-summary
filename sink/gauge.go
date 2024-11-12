package sink

import (
	"time"

	"go.k6.io/k6/metrics"
)

// We want to make sure *GaugeSink implements the Sink interface.
var _ Sink = &GaugeSink{}

// GaugeSink is a wrapper of metrics.GaugeSink that implements the Sink interface.
type GaugeSink struct {
	*metrics.GaugeSink
	last time.Time
}

// Add is a wrapper of metrics.GaugeSink.Add that keeps track of the most recent
// value, so at Merge time, we can determine what's the most recent value.
func (g *GaugeSink) Add(s metrics.Sample) {
	g.GaugeSink.Add(s)
	if s.Time.After(g.last) {
		g.last = s.Time
	}
}

// Merge merges the given sink into the current one.
// If the given sink is not a *GaugeSink, it panics.
func (g *GaugeSink) Merge(s Sink) {
	toMerge, ok := s.(*GaugeSink)
	if !ok {
		panic("trying to merge incompatible sinks")
	}

	if toMerge.Max > g.Max {
		g.Max = toMerge.Max
	}

	if toMerge.Min < g.Min {
		g.Min = toMerge.Min
	}

	if toMerge.last.After(g.last) {
		g.last = toMerge.last
		g.Value = toMerge.Value
	}
}
