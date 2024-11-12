package custosummary

import (
	"fmt"
	"regexp"
	"strings"

	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

type (
	// ModuleInstance represents an instance of the JS module.
	ModuleInstance struct {
		vu   modules.VU
		root *RootModule
	}
)

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Instance = &ModuleInstance{}
)

// Exports implements the output.Output interface, by returning the module's (ESM) exports.
func (m ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"includeAllMetrics":    m.includeAllMetrics,
			"excludeAllMetrics":    m.excludeAllMetrics,
			"filterMetric":         m.filterMetric,
			"filterMetricByRegexp": m.filterMetricByRegexp,
		},
	}
}

func (m ModuleInstance) includeAllMetrics() {
	m.debug("All metrics will be included")
	// TODO: Implement this.
}

func (m ModuleInstance) excludeAllMetrics() {
	m.debug("All metrics will be excluded")
	// TODO: Implement this.
}

func (m ModuleInstance) filterMetric(name string) {
	m.debug("Metric '" + name + "' will be filtered")
	// TODO: Implement this.
}

func (m ModuleInstance) filterMetricByRegexp(re string) {
	_, err := regexp.Compile(re)
	if err != nil {
		m.error("Metrics regexp '" + re + "' is invalid: " + err.Error())
		// FIXME: Can we avoid the 'GoError' and stack trace here?
		common.Throw(m.vu.Runtime(), err)
		return
	}

	m.debug("Metrics will be filtered by regexp '" + re + "'")
	// TODO: Implement this.
}

// log is a helper method to log (i.e. `console.log`) messages that serves
// as a workaround for the lack of a proper logger in the module instance
// during the init context.
func (m ModuleInstance) log(str string) {
	str = strings.ReplaceAll(str, "`", "'") // For safety.
	_, err := m.vu.Runtime().RunString(fmt.Sprintf("console.log(`%s`)", str))
	if err != nil {
		panic(err)
	}
}

// error is a helper method to error (i.e. `console.error`) messages that serves
// as a workaround for the lack of a proper logger in the module instance
// during the init context.
func (m ModuleInstance) error(str string) {
	str = strings.ReplaceAll(str, "`", "'") // For safety.
	_, err := m.vu.Runtime().RunString(fmt.Sprintf("console.error(`%s`)", str))
	if err != nil {
		panic(err)
	}
}

// debug is a helper method to debug (i.e. `console.debug`) messages that serves
// as a workaround for the lack of a proper logger in the module instance
// during the init context.
func (m ModuleInstance) debug(str string) {
	str = strings.ReplaceAll(str, "`", "'") // For safety.
	_, err := m.vu.Runtime().RunString(fmt.Sprintf("console.debug(`%s`)", str))
	if err != nil {
		panic(err)
	}
}
