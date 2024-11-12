package timeseries

import (
	"sort"
	"strings"

	"github.com/mstoykov/atlas"

	"go.k6.io/k6/metrics"

	"github.com/joanlopez/xk6-custosummary/sink"
)

// Collection is a collection of time series.
type Collection map[Key]TimeSeries

// NewCollection initializes a new empty Collection.
func NewCollection() Collection {
	return make(map[Key]TimeSeries)
}

// AddSample is the equivalent of AddMetricSample,
// but it takes the *Metric from the given Sample.
func (c Collection) AddSample(s metrics.Sample) {
	c.AddMetricSample(s.Metric, s)
}

// AddMetricSample adds the sample to the Sink of the corresponding time series,
// which is identified by the given metric and sample's tags.
// If there's no Sink for that time series yet stored in the collection,
// it is also responsible for its initialization.
func (c Collection) AddMetricSample(m *metrics.Metric, s metrics.Sample) {
	k := NewKey(metrics.TimeSeries{
		Metric: m,
		Tags:   s.TimeSeries.Tags,
	})

	if _, exists := c[k]; !exists {
		c[k] = TimeSeries{
			Key: k,
			Meta: Meta{
				Type:     m.Type,
				Contains: m.Contains,
			},
			Sink: sink.New(m.Type),
		}
	}

	c[k].Sink.Add(s)
}

// Get returns a TimeSeries that matches the given key.
//
// Use NewKey to create a key from a TimeSeries.
//
// This method merges all the time series that match the key prefix, so it behaves like a Prometheus query:
//   - http_reqs{} => will return a time series with all the values from the `http_reqs` metric.
//   - http_reqs{group='auth'} => will return a time series with all the values from the `http_reqs` metric, tagged with `group=auth`.
func (c Collection) Get(get Key) *TimeSeries {
	// We merge all the stored time series that matches
	// the key prefix.
	var result *TimeSeries
	for key, ts := range c {
		// If the time series key is a prefix of the given key,
		// we merge the sink. If not, we skip it.
		if !strings.HasPrefix(string(key), string(get)) {
			continue
		}

		// If we don't have a totalSink yet, we initialize it
		// with the same type as the time series sink.
		if result == nil {
			result = &TimeSeries{
				Key:  get,
				Meta: ts.Meta,
				Sink: sink.New(ts.Meta.Type),
			}
		}
		result.Sink.Merge(ts.Sink)
	}

	return result
}

// Meta defines the shape (metric and values type) of a time series.
type Meta struct {
	Type     metrics.MetricType
	Contains metrics.ValueType
}

// TimeSeries holds all the values of a given time series,
// identified by Key, shaped by Meta, in a Sink.
type TimeSeries struct {
	Key
	Meta Meta
	Sink sink.Sink
}

// Key is a unique identification for time series.
type Key string

// NewKey returns a key that uniquely identifies the given time series.
// It sorts the labels to ensure that the key is always the same.
//
// It follows a style similar to Prometheus queries:
//   - http_reqs{} => NewKey(metrics.TimeSeries{Metric: &metrics.Metric{Name: "http_reqs"}}).
//   - http_reqs{group='auth'} => NewKey(metrics.TimeSeries{Metric: &metrics.Metric{Name: "http_reqs"}, Tags: metrics.TagSet.With("group", "auth")}).
func NewKey(ts metrics.TimeSeries) Key {
	labelPairs := []string{"__name__=" + ts.Metric.Name}
	for k, v := range normalizeTagSet(ts.Tags).Map() {
		// FIXME: Find a more efficient way to do this, like hashing.
		labelPairs = append(labelPairs, k+"="+v)
	}
	sort.Strings(labelPairs)
	return Key(strings.Join(labelPairs, "|"))
}

// MetricName returns the metric name from the key.
//
// It is internally stored as `__name__`.
func (k Key) MetricName() string {
	// All the accesses by index here should be fine.
	return strings.Split(strings.Split(string(k), "|")[0], "=")[1]
}

// MetricNameKey returns a Key with only the metric name.
// It can be used in combination with Collection.Get,
// to get a time series with all the values from a given metric, despite the tags.
func (k Key) MetricNameKey() Key {
	// Access by index here should be fine.
	return Key(strings.Split(string(k), "|")[0])
}

// FIXME: This should be configurable
// normalizeTagSet is a helper function to "normalize" a given *TagSet.
// Normalization here means just keeping the labels we're interested in.
func normalizeTagSet(ts *metrics.TagSet) *metrics.TagSet {
	// For now, we only care about 'group' and 'scenario'.
	labelSet := []string{"group", "scenario"}

	// Create a new *TagSet, set the label values we're interested in, if present, and return it.
	result := newTagSet()
	for _, label := range labelSet {
		if value, hasLabel := ts.Get(label); hasLabel {
			result = result.With(label, value)
		}
	}
	return ts
}

// newTagSet is a helper function to initialize an empty *TagSet.
func newTagSet() *metrics.TagSet {
	return (*metrics.TagSet)(atlas.New())
}
