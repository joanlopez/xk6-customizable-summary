package report

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.k6.io/k6/lib"

	"github.com/joanlopez/xk6-custosummary/sink"
	"github.com/joanlopez/xk6-custosummary/timeseries"
)

// Report holds the source data to build a human-readable summary (see the `summary` package).
// It is mainly composed by a map of metrics, where the key is the metric name.
type Report struct {
	Metrics map[string]Metric
}

// From creates a Report from a timeseries.Collection.
//
// For now, it only adds a report.Metric for each pair of metric name and
// timeseries.TimeSeries in the collection, without really using tags
// (e.g. group, scenario, etc.).
//
// In the future, we might want to define a default behavior that also uses
// certain tags, and make it configure, so the user can choose.
func From(
	c timeseries.Collection,
	testDuration time.Duration, opts lib.Options,
) Report {
	r := Report{Metrics: make(map[string]Metric)}
	getMetricValues := metricValueGetter(opts.SummaryTrendStats)

	// We only want to add a report.Metric for each unique metric name.
	seen := make(map[string]struct{})
	for _, ts := range c {
		metricName := ts.Key.MetricName()
		if _, ok := seen[metricName]; ok {
			continue
		}

		// The call to `c.Get(ts.Key.MetricNameKey()).Sink` should return
		// a Sink that has been filled with all the samples for the metric,
		// despite the tags.
		seen[ts.Key.MetricName()] = struct{}{}
		r.Metrics[metricName] = Metric{
			Meta:   ts.Meta,
			Values: getMetricValues(c.Get(ts.Key.MetricNameKey()).Sink, testDuration),
		}
	}

	return r
}

// Metric is a metric that belongs to a report.Report.
// So, it doesn't exactly correlate with a k6 metric, but it's a representation.
type Metric struct {
	timeseries.Meta
	Values map[string]float64
}

// metricValueGetter returns a function that can extract the values from a sink.Sink
// that are going to be used in the report, depending on the sink type.
// For instance, for Counter sinks it will return the count and the rate.
func metricValueGetter(summaryTrendStats []string) func(sink.Sink, time.Duration) map[string]float64 {
	trendResolvers, err := getResolversForTrendColumns(summaryTrendStats)
	if err != nil {
		panic(err.Error()) // this should have been validated already
	}

	return func(s sink.Sink, t time.Duration) (result map[string]float64) {
		switch typed := s.(type) {
		case *sink.CounterSink:
			result = typed.Format(t)
			result["rate"] = calculateCounterRate(typed.Value, t)
		case *sink.GaugeSink:
			result = typed.Format(t)
			result["min"] = typed.Min
			result["max"] = typed.Max
		case *sink.RateSink:
			result = typed.Format(t)
			result["passes"] = float64(typed.Trues)
			result["fails"] = float64(typed.Total - typed.Trues)
		case *sink.TrendSink:
			result = make(map[string]float64, len(summaryTrendStats))
			for _, col := range summaryTrendStats {
				result[col] = trendResolvers[col](typed)
			}
		}

		return result
	}
}

// getResolversForTrendColumns checks if passed trend columns are valid for use in
// the summary output and then returns a map of the corresponding resolvers.
func getResolversForTrendColumns(trendColumns []string) (map[string]func(s *sink.TrendSink) float64, error) {
	staticResolvers := map[string]func(s *sink.TrendSink) float64{
		"avg":   func(s *sink.TrendSink) float64 { return s.Avg() },
		"min":   func(s *sink.TrendSink) float64 { return s.Min() },
		"med":   func(s *sink.TrendSink) float64 { return s.P(0.5) },
		"max":   func(s *sink.TrendSink) float64 { return s.Max() },
		"count": func(s *sink.TrendSink) float64 { return float64(s.Count()) },
	}
	dynamicResolver := func(percentile float64) func(s *sink.TrendSink) float64 {
		return func(s *sink.TrendSink) float64 {
			return s.P(percentile / 100)
		}
	}

	result := make(map[string]func(s *sink.TrendSink) float64, len(trendColumns))

	for _, stat := range trendColumns {
		if staticStat, ok := staticResolvers[stat]; ok {
			result[stat] = staticStat
			continue
		}

		percentile, err := parsePercentile(stat)
		if err != nil {
			return nil, err
		}
		result[stat] = dynamicResolver(percentile)
	}

	return result, nil
}

// parsePercentile is a helper function to parse and validate percentile notations
func parsePercentile(stat string) (float64, error) {
	if !strings.HasPrefix(stat, "p(") || !strings.HasSuffix(stat, ")") {
		return 0, fmt.Errorf("invalid trend stat '%s', unknown format", stat)
	}

	percentile, err := strconv.ParseFloat(stat[2:len(stat)-1], 64)

	if err != nil || (percentile < 0) || (percentile > 100) {
		return 0, fmt.Errorf("invalid percentile trend stat value '%s', provide a number between 0 and 100", stat)
	}

	return percentile, nil
}

// calculateCounterRate calculates the rate of a counter metric,
// given the count and the duration of the test.
func calculateCounterRate(count float64, duration time.Duration) float64 {
	if duration == 0 {
		return 0
	}
	return count / (float64(duration) / float64(time.Second))
}
