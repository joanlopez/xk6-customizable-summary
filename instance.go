package custosummary

import (
	"regexp"

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
	if m.vu.State() != nil {
		m.vu.State().Logger.Errorln("'includeAllMetrics' must be called in the init context to take effect")
		return
	}

	m.vu.InitEnv().Logger.Debugln("All metrics will be included")
	// TODO: Implement this.
}

func (m ModuleInstance) excludeAllMetrics() {
	if m.vu.State() != nil {
		m.vu.State().Logger.Errorln("'excludeAllMetrics' must be called in the init context to take effect")
		return
	}

	m.vu.InitEnv().Logger.Debugln("All metrics will be excluded")
	// TODO: Implement this.
}

func (m ModuleInstance) filterMetric(name string) {
	if m.vu.State() != nil {
		m.vu.State().Logger.Errorln("'filterMetric' must be called in the init context to take effect")
		return
	}

	m.vu.InitEnv().Logger.Debugln("Metric '" + name + "' will be filtered")
	// TODO: Implement this.
}

func (m ModuleInstance) filterMetricByRegexp(re string) {
	if m.vu.State() != nil {
		m.vu.State().Logger.Errorln("'filterMetricByRegexp' must be called in the init context to take effect")
		return
	}

	_, err := regexp.Compile(re)
	if err != nil {
		m.vu.InitEnv().Logger.Errorln("Metrics regexp '" + re + "' is invalid: " + err.Error())
		// FIXME: Can we avoid the 'GoError' and stack trace here?
		common.Throw(m.vu.Runtime(), err)
		return
	}

	m.vu.InitEnv().Logger.Debugln("Metrics will be filtered by regexp '" + re + "'")
	// TODO: Implement this.
}
