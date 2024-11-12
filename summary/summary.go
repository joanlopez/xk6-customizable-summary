package summary

import (
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"

	"github.com/joanlopez/xk6-custosummary/report"
)

// Code heavily inspired by: https://github.com/grafana/k6/blob/master/js/summary.js.

// Summary is a sequence of lines (strings) that represents a
// human-readable summary of a report.Report, that can be written to an io.Writer.
type Summary []string

// WriteTo writes the summary to the given io.Writer.
func (ss Summary) WriteTo(w io.Writer) (n int64, err error) {
	clearLine := func() {
		_, _ = fmt.Fprintf(w, "\r")
		_, _ = fmt.Fprintf(w, strings.Repeat(" ", 100))
		_, _ = fmt.Fprintf(w, "\r")
	}
	for i, s := range ss {
		if i < 5 {
			// We clear the first few lines to avoid the
			// summary to be printed on top of the progress bar.
			clearLine()
		}
		nn, ee := fmt.Fprintln(w, s)
		n += int64(nn)
		err = errors.Join(err, ee)

	}
	return
}

// From creates a Summary from a report.Report.
// It is heavily inspired by the JavaScript implementation in k6.
func From(r report.Report, opts lib.Options) Summary {
	var s Summary

	const indent = "   "

	var names []string
	nameLenMax := 0

	nonTrendValues := map[string]string{}
	nonTrendValueMaxLen := 0
	nonTrendExtras := map[string][]string{}
	nonTrendExtraMaxLens := []int{0, 0}

	trendCols := map[string][]string{}
	numTrendColumns := len(opts.SummaryTrendStats)
	trendColMaxLens := make([]int, numTrendColumns)

	for name, metric := range r.Metrics {
		names = append(names, name)
		displayName := indentForMetric(name) + displayNameForMetric(name)
		displayNameWidth := strWidth(displayName)
		if displayNameWidth > nameLenMax {
			nameLenMax = displayNameWidth
		}

		if metric.Type == metrics.Trend {
			cols := make([]string, numTrendColumns)
			for i, tc := range opts.SummaryTrendStats {
				value := fmt.Sprintf("%v", metric.Values[tc])
				if tc != "count" {
					value = humanizeValue(metric.Values[tc], metric, opts.SummaryTimeUnit.String)
				}
				valLen := strWidth(value)
				if valLen > trendColMaxLens[i] {
					trendColMaxLens[i] = valLen
				}
				cols[i] = value
			}
			trendCols[name] = cols
			continue
		}

		values := nonTrendMetricValueForSum(metric, opts.SummaryTimeUnit.String)
		nonTrendValues[name] = values[0]
		valueLen := strWidth(values[0])
		if valueLen > nonTrendValueMaxLen {
			nonTrendValueMaxLen = valueLen
		}
		nonTrendExtras[name] = values[1:]
		for i := 1; i < len(values); i++ {
			extraLen := strWidth(values[i])
			if extraLen > nonTrendExtraMaxLens[i-1] {
				nonTrendExtraMaxLens[i-1] = extraLen
			}
		}
	}

	sort.Slice(names, func(i, j int) bool {
		return strings.Compare(names[i], names[j]) < 0
	})

	getData := func(name string) string {
		if cols, found := trendCols[name]; found {
			tmpCols := make([]string, numTrendColumns)
			for i, col := range cols {
				tmpCols[i] = fmt.Sprintf("%s=%s%s",
					opts.SummaryTrendStats[i],
					decorate(col, palette["cyan"]),
					strings.Repeat(" ", trendColMaxLens[i]-strWidth(col)),
				)
			}
			return strings.Join(tmpCols, " ")
		}

		value := nonTrendValues[name]
		fmtData := decorate(value, palette["cyan"]) + strings.Repeat(" ", nonTrendValueMaxLen-strWidth(value))

		extras := nonTrendExtras[name]
		if len(extras) == 1 {
			fmtData += " " + decorate(extras[0], palette["cyan"], palette["faint"])
		} else if len(extras) > 1 {
			parts := make([]string, len(extras))
			for i, extra := range extras {
				parts[i] = decorate(extra, palette["cyan"], palette["faint"]) +
					strings.Repeat(" ", nonTrendExtraMaxLens[i]-strWidth(extra))
			}
			fmtData += " " + strings.Join(parts, " ")
		}

		return fmtData
	}

	for _, name := range names {
		mark := " "
		markColor := func(text string) string { return text }

		fmtIndent := indentForMetric(name)
		fmtName := displayNameForMetric(name)
		fmtName += decorate(strings.Repeat(".", nameLenMax-strWidth(fmtName)-strWidth(fmtIndent)+3)+":", palette["faint"])

		s = append(s, indent+fmtIndent+markColor(mark)+" "+fmtName+" "+getData(name))
	}

	return s
}

func indentForMetric(name string) string {
	if strings.Contains(name, "{") {
		return "  "
	}
	return ""
}

func displayNameForMetric(name string) string {
	subMetricPos := strings.Index(name, "{")
	if subMetricPos >= 0 {
		return "{ " + name[subMetricPos+1:len(name)-1] + " }"
	}
	return name
}

func strWidth(s string) int {
	// Normalize the string to NFKC form
	data := norm.NFKC.String(s)

	inEscSeq := false
	inLongEscSeq := false
	width := 0

	for _, char := range data {
		// Skip over ANSI escape codes
		if char == '\x1b' { // Start of ANSI escape sequence
			inEscSeq = true
			continue
		}
		if inEscSeq && char == '[' {
			inLongEscSeq = true
			continue
		}
		if inEscSeq && inLongEscSeq && (char >= 0x40 && char <= 0x7e) {
			inEscSeq = false
			inLongEscSeq = false
			continue
		}
		if inEscSeq && !inLongEscSeq && (char >= 0x40 && char <= 0x5f) {
			inEscSeq = false
			continue
		}

		// If not in escape sequence, increase width
		if !inEscSeq && !inLongEscSeq {
			// Use utf8.RuneLen to account for multi-byte characters
			width += utf8.RuneLen(char)
		}
	}
	return width
}

func humanizeValue(val float64, metric report.Metric, timeUnit string) string {
	if metric.Type == metrics.Rate {
		// Truncate instead of round to 2 decimal places
		truncated := math.Trunc(val*100*100) / 100
		return fmt.Sprintf("%.2f%%", truncated)
	}

	switch metric.Contains {
	case metrics.Data:
		return humanizeBytes(val)
	case metrics.Time:
		return humanizeDuration(val, timeUnit)
	default:
		return toFixedNoTrailingZeros(val, 6)
	}
}

func humanizeBytes(bytes float64) string {
	units := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	base := 1000.0

	if bytes < 10 {
		return fmt.Sprintf("%.0f B", bytes)
	}

	e := math.Floor(math.Log(bytes) / math.Log(base))
	suffix := units[int(e)]
	val := math.Floor((bytes/math.Pow(base, e))*10+0.5) / 10 // Truncate to one decimal place if necessary

	// Format with 1 decimal if val is less than 10, otherwise no decimals
	if val < 10 {
		return fmt.Sprintf("%.1f %s", val, suffix)
	}
	return fmt.Sprintf("%.0f %s", val, suffix)
}

type unit struct {
	coef float64
	unit string
}

var unitMap = map[string]unit{
	"s":  {unit: "s", coef: 0.001},
	"ms": {unit: "ms", coef: 1},
	"us": {unit: "µs", coef: 1000},
}

func humanizeDuration(dur float64, timeUnit string) string {
	if unit, exists := unitMap[timeUnit]; exists {
		return fmt.Sprintf("%.2f %s", dur*unit.coef, unit.unit)
	}

	return humanizeGenericDuration(dur)
}

func humanizeGenericDuration(dur float64) string {
	if dur == 0 {
		return "0s"
	}

	if dur < 0.001 {
		return fmt.Sprintf("%dns", int64(dur*1000000))
	}
	if dur < 1 {
		return toFixedNoTrailingZerosTrunc(dur*1000, 2) + "µs"
	}
	if dur < 1000 {
		return toFixedNoTrailingZerosTrunc(dur, 2) + "ms"
	}

	result := toFixedNoTrailingZerosTrunc(math.Mod(dur, 60000)/1000, 2) + "s"
	rem := math.Trunc(dur / 60000)

	if rem < 1 {
		return result
	}

	result = fmt.Sprintf("%dm%s", int(rem)%60, result)
	rem = math.Trunc(rem / 60)

	if rem < 1 {
		return result
	}

	return fmt.Sprintf("%dh%s", int(rem), result)
}

func toFixedNoTrailingZeros(val float64, prec int) string {
	format := "%." + strconv.Itoa(prec) + "f"
	str := fmt.Sprintf(format, val)

	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")
	return str
}

func toFixedNoTrailingZerosTrunc(val float64, prec int) string {
	mult := math.Pow(10, float64(prec))
	truncatedVal := math.Trunc(val*mult) / mult
	return toFixedNoTrailingZeros(truncatedVal, prec)
}

func nonTrendMetricValueForSum(metric report.Metric, timeUnit string) []string {
	switch metric.Type {
	case metrics.Counter:
		return []string{
			humanizeValue(metric.Values["count"], metric, timeUnit),
			humanizeValue(metric.Values["rate"], metric, timeUnit) + "/s",
		}
	case metrics.Gauge:
		return []string{
			humanizeValue(metric.Values["value"], metric, timeUnit),
			"min=" + humanizeValue(metric.Values["min"], metric, timeUnit),
			"max=" + humanizeValue(metric.Values["max"], metric, timeUnit),
		}
	case metrics.Rate:
		passes := metric.Values["passes"]
		fails := metric.Values["fails"]
		total := passes + fails
		return []string{
			humanizeValue(metric.Values["rate"], metric, timeUnit),
			fmt.Sprintf("%.0f out of %.0f", passes, total),
		}
	default:
		return []string{"[no data]"}
	}
}

func decorate(text string, colorCode string, additionalCodes ...string) string {
	result := "\x1b[" + colorCode
	for _, code := range additionalCodes {
		result += ";" + code
	}
	return fmt.Sprintf("%sm%s\x1b[0m", result, text)
}

var palette = map[string]string{
	"faint": "2",
	"cyan":  "36",
}
