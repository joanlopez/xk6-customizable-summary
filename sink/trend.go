package sink

import (
	"github.com/joanlopez/xk6-custosummary/sink/trend"
)

// We want to make sure *TrendSink implements the Sink interface.
var _ Sink = &TrendSink{}

// TrendSink is a wrapper of trend.Sink that implements the Sink interface.
// It can also be considered as a "proxy type", as there are multiple
// implementations of the trend.Sink interface.
type TrendSink struct {
	trend.Sink
}

// Merge merges the given sink into the current one.
// If the given sink is not a *TrendSink, it panics.
// If inner trend.Sink implementation doesn't match, it also panics.
func (t *TrendSink) Merge(s Sink) {
	toMerge, ok := s.(*TrendSink)
	if !ok {
		panic("trying to merge incompatible sinks")
	}

	switch s := toMerge.Sink.(type) {
	case *trend.K6Sink:
		casted, ok := t.Sink.(*trend.K6Sink)
		if !ok {
			panic("trying to merge incompatible trend sinks")
		}
		casted.Merge(s)
	case trend.DDSketchHistogramSink:
		casted, ok := t.Sink.(trend.DDSketchHistogramSink)
		if !ok {
			panic("trying to merge incompatible trend sinks")
		}
		casted.Merge(s)
	case trend.HdrHistogramSink:
		casted, ok := t.Sink.(trend.HdrHistogramSink)
		if !ok {
			panic("trying to merge incompatible trend sinks")
		}
		casted.Merge(s)
	}

}
