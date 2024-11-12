package trend

import (
	"time"

	"github.com/DataDog/sketches-go/ddsketch"

	"go.k6.io/k6/metrics"
)

// DDSketchHistogramSink is a Sink implementation that relies
// on a DDSketch under that hood. So, it is less accurate
// but also has less impact in terms of memory allocations.
type DDSketchHistogramSink struct {
	dds *ddsketch.DDSketch
}

// NewDDSketchHistogramSink instantiates a new
// DDSketchHistogramSink with a relative accuracy of 0.01.
func NewDDSketchHistogramSink() DDSketchHistogramSink {
	// FIXME: Handle error
	dds, _ := ddsketch.NewDefaultDDSketch(0.01)
	return DDSketchHistogramSink{dds: dds}
}

// IsEmpty indicates whether the TrendSink is empty.
func (d DDSketchHistogramSink) IsEmpty() bool { return d.dds.GetCount() == 0 }

// Add implements the Sink interface, recording the value of the given metrics.Sample.
func (d DDSketchHistogramSink) Add(s metrics.Sample) {
	// FIXME: Handle error
	_ = d.dds.Add(s.Value)
}

// P implements the Sink interface, returning the value at percentile.
func (d DDSketchHistogramSink) P(pct float64) float64 {
	// FIXME: Handle error
	val, _ := d.dds.GetValueAtQuantile(pct)
	return val
}

// Min implements the Sink interface, returning the minimum value recorded.
func (d DDSketchHistogramSink) Min() float64 {
	// FIXME: Handle error
	val, _ := d.dds.GetMinValue()
	return val
}

// Max implements the Sink interface, returning the maximum value recorded.
func (d DDSketchHistogramSink) Max() float64 {
	// FIXME: Handle error
	val, _ := d.dds.GetMaxValue()
	return val
}

// Count implements the Sink interface, returning the total amount of values recorded.
func (d DDSketchHistogramSink) Count() uint64 {
	return uint64(d.dds.GetCount())
}

// Avg implements the Sink interface, returning the average (mean) of values recorded.
func (d DDSketchHistogramSink) Avg() float64 {
	return d.dds.GetSum() / d.dds.GetCount()
}

// Format trend and return a map
func (d DDSketchHistogramSink) Format(_ time.Duration) map[string]float64 {
	return map[string]float64{
		"min":   d.Min(),
		"max":   d.Max(),
		"avg":   d.Avg(),
		"med":   d.P(0.5),
		"p(90)": d.P(0.90),
		"p(95)": d.P(0.95),
	}
}

// Merge merges two Sink instances.
func (d DDSketchHistogramSink) Merge(s Sink) {
	toMerge, ok := s.(DDSketchHistogramSink)
	if !ok {
		panic("trying to merge incompatible trend sinks")
	}

	// FIXME: Handle error
	_ = d.dds.MergeWith(toMerge.dds)
}

// We want to make sure that the DDSketchHistogramSink
// implements the TrendSink interface.
var _ Sink = DDSketchHistogramSink{}
