package trend

import (
	"go.k6.io/k6/metrics"

	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/HdrHistogram/hdrhistogram-go"
)

type Sink interface {
	// P calculates the given percentile from sink values.
	// The given percentile must be a floating point within [0, 1] range.
	P(pct float64) float64

	// Min returns the minimum value.
	Min() float64

	// Max returns the maximum value.
	Max() float64

	// Count returns the number of recorded values.
	Count() uint64

	// Avg returns the average (i.e. mean) value.
	Avg() float64

	// Add a single sample into the trend
	Add(s metrics.Sample)
}

// We want to make sure that the official
// metrics.TrendSink implements our own Sink interface.
var _ Sink = &metrics.TrendSink{}

// HdrHistogramSink is a Sink implementation that relies
// on a HdrHistogram under that hood. So, it is less accurate
// but also has less impact in terms of memory allocations.
type HdrHistogramSink struct {
	hdr *hdrhistogram.Histogram
}

// NewHdrHistogramSink instantiates a new HdrHistogramSink
// with values between [1, 100000000000] with the number
// of significant value digits up to 5.
func NewHdrHistogramSink() HdrHistogramSink {
	return HdrHistogramSink{
		hdr: hdrhistogram.New(
			1,
			100000000000,
			5,
		),
	}
}

// P implements the Sink interface, returning the value at percentile.
func (h HdrHistogramSink) P(pct float64) float64 {
	return float64(h.hdr.ValueAtPercentile(pct * 100.0))
}

// Min implements the Sink interface, returning the minimum value recorded.
func (h HdrHistogramSink) Min() float64 {
	return float64(h.hdr.Min())
}

// Max implements the Sink interface, returning the maximum value recorded.
func (h HdrHistogramSink) Max() float64 {
	return float64(h.hdr.Max())
}

// Count implements the Sink interface, returning the total amount of values recorded.
func (h HdrHistogramSink) Count() uint64 {
	return uint64(h.hdr.TotalCount())
}

// Avg implements the Sink interface, returning the average (mean) of values recorded.
func (h HdrHistogramSink) Avg() float64 {
	return h.hdr.Mean()
}

// Add implements the Sink interface, recording the value of the given metrics.Sample.
func (h HdrHistogramSink) Add(s metrics.Sample) {
	// FIXME: Handle error
	_ = h.hdr.RecordValue(int64(s.Value))
}

// We want to make sure that the HdrHistogramSink
// implements the TrendSink interface.
var _ Sink = HdrHistogramSink{}

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

// Add implements the Sink interface, recording the value of the given metrics.Sample.
func (d DDSketchHistogramSink) Add(s metrics.Sample) {
	// FIXME: Handle error
	_ = d.dds.Add(s.Value)
}

// We want to make sure that the DDSketchHistogramSink
// implements the TrendSink interface.
var _ Sink = DDSketchHistogramSink{}
