package custosummary

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"

	"github.com/joanlopez/xk6-custosummary/report"
	"github.com/joanlopez/xk6-custosummary/summary"
	"github.com/joanlopez/xk6-custosummary/timeseries"
)

// TODO: Parameterize
const (
	flushInterval = 1 * time.Second
)

func init() {
	// Initialize the global RootModule instance accessor.
	root := &RootModule{
		Collection: timeseries.NewCollection(),
	}

	New = func() *RootModule { return root }

	// Register the extension as a k6 module.
	modules.Register("k6/x/custosummary", New())

	// Register the extension as a k6 output.
	output.RegisterExtension("xk6-custosummary", NewOutput)
}

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct {
		params output.Params

		start time.Time
		timeseries.Collection

		output.SampleBuffer
		periodicFlusher *output.PeriodicFlusher
		logger          logrus.FieldLogger
	}
)

// Ensure the interfaces are implemented correctly.
var _ interface {
	modules.Module
	output.WithStopWithTestError
} = &RootModule{}

// New returns a pointer to a new RootModule instance.
var New func() *RootModule

// NewOutput is a wrapper on top of New, that uses the given output.Params
// and returns (the same) output.Output instance.
func NewOutput(params output.Params) (output.Output, error) {
	root := New()
	root.params = params
	root.logger = params.Logger
	return root, nil
}

// NewModuleInstance implements the modules.Module interface returning a new instance for each VU.
func (rm *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu:   vu,
		root: rm,
	}
}

// Description implements the output.Output interface, by returning the module's description.
func (*RootModule) Description() string {
	return "xk6-custosummary"
}

// Start implements the output.Output interface, exposing a method to initialize the output.
func (rm *RootModule) Start() error {
	rm.logger.Debug("Starting output...")

	pf, err := output.NewPeriodicFlusher(flushInterval, rm.flushMetrics)
	if err != nil {
		return err
	}

	rm.logger.Debug("Started!")
	rm.start = time.Now()
	rm.periodicFlusher = pf

	return nil
}

// StopWithTestError flushes all remaining metrics and finalizes the test run
func (rm *RootModule) StopWithTestError(err error) error {
	logger := rm.loggerWithError(err)
	logger.Debug("Stopping...")
	defer rm.logger.Debug("Stopped!")

	rm.periodicFlusher.Stop()

	r := report.From(rm.Collection, time.Since(rm.start), rm.params.ScriptOptions)
	s := summary.From(r, rm.params.ScriptOptions)
	_, _ = fmt.Fprintln(os.Stdout) // FIXME: Handle error.
	_, _ = s.WriteTo(os.Stdout)    // FIXME: Handle error.

	return nil
}

// Stop implements the output.Output interface, exposing a method to stop the output.
func (rm *RootModule) Stop() error {
	return rm.StopWithTestError(nil)
}

func (rm *RootModule) loggerWithError(err error) logrus.FieldLogger {
	logger := rm.logger
	if err != nil {
		logger = logger.WithError(err)
	}
	return logger
}

func (rm *RootModule) flushMetrics() {
	samples := rm.GetBufferedSamples()
	for _, sc := range samples {
		samples := sc.GetSamples()
		for _, sample := range samples {
			rm.flushSample(sample)
		}
	}
}

func (rm *RootModule) flushSample(s metrics.Sample) {
	// We register the metric and its sub-metrics,
	// and we add the sample value to their sinks.
	rm.AddSample(s)
	for _, sub := range s.Metric.Submetrics {
		rm.AddMetricSample(sub.Metric, s)
	}
}
