package sink

import (
	"fmt"

	"go.k6.io/k6/metrics"

	"github.com/joanlopez/xk6-custosummary/sink/trend"
)

// Sink defined the behavior expected for a sink implementation.
// It is heavily inspired by the metrics.Sink interface, but it also
// adds a method to merge two sinks, used to simplify the process of
// accumulating partial results at time of building the report/summary.
type Sink interface {
	metrics.Sink
	Merge(s Sink)
}

// New creates a new Sink based on the given metrics.MetricType.
// It is a wrapper around metrics.NewSink, but to initialize a
// Sink instead of a metrics.Sink.
func New(mt metrics.MetricType) Sink {
	var sink Sink
	switch mt {
	case metrics.Counter:
		sink = &CounterSink{CounterSink: &metrics.CounterSink{}}
	case metrics.Gauge:
		sink = &GaugeSink{GaugeSink: &metrics.GaugeSink{}}
	case metrics.Trend:
		sink = &TrendSink{Sink: trend.NewSink()}
	case metrics.Rate:
		sink = &RateSink{RateSink: &metrics.RateSink{}}
	default:
		panic(fmt.Sprintf("MetricType %q is not supported", mt))
	}
	return sink
}
