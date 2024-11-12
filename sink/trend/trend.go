package trend

import (
	"os"
	"time"

	"go.k6.io/k6/metrics"
)

// Sink defines the behavior expected for any of the
// trend.Sink implementations living on this package.
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

	// Merge merges two Sink instances.
	Merge(s Sink)

	// Format returns data for thresholds.
	Format(t time.Duration) map[string]float64

	// IsEmpty check if the Sink is empty.
	IsEmpty() bool
}

// NewSink instantiates a new Sink.
//
// By default, it initializes a *K6Sink.
// This can be changed by using the XK6_CUSTOSUMMARY_TRENDSINK_TYPE environment variable.
// Possible values are:
//   - "k6" (default) => *K6Sink
//   - "hdr"		  => HdrHistogramSink
//   - "dds" 		  => DDSketchHistogramSink
func NewSink() Sink {
	sinkType, ok := os.LookupEnv(sinkTypeEnvVar)
	if !ok || len(sinkType) == 0 {
		return NewK6Sink() // Default
	}

	switch sinkType {
	case "hdr":
		return NewHdrHistogramSink()
	case "dds":
		return NewDDSketchHistogramSink()
	case "k6":
		return NewK6Sink()
	default:
		panic("unknown trend sink type: " + sinkType)
	}
}

// Possible values are: "k6" (default), "hdr", and "dds".
const sinkTypeEnvVar = "XK6_CUSTOSUMMARY_TRENDSINK_TYPE"
