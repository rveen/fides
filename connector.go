package fides

import (
	// "log"
	"math"
	"strings"
)

func ConnectorFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	// PCB connectors, SMD, less that one insertion/year

	piReport := 6.0

	// TODO derive smd from package
	if containsTag(comp.Tags, "smd") {
		piReport = 10
	}
	l0connector := 0.1 * piReport * 0.2 * math.Pow(float64(comp.N), 0.5) // PCB connectors

	// log.Printf("Connector %s: N=%d, l0=%f\n", comp.Name, comp.N, l0connector)

	for _, ph := range mission.Phases {

		// Thermal
		nfit = 0.58 * PiThermal_connector(ph.On, ph.Tamb)

		// Thermal cycling case
		nfit += 0.04 * PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax)

		// Mechanical
		nfit += 0.05 * PiMech(ph.Grms)

		// RH
		nfit += 0.13 * PiRHea(ph.RH, ph.Tamb, 0.8)

		// Chemical
		if !ph.IP {
			nfit += 0.2 * ph.SalinePollution * ph.AmbientPollution * ph.ApplicationPollution
		}

		// Force PiPlacement to 1
		nfit *= PiInduced(ph.On, false, false, comp.IsPower, 4.4)

		// Time
		nfit *= ph.Time / 8760.0 * l0connector

		fit += nfit
	}
	return fit
}

// TODO Change comp.Tags to []string
func containsTag(tags, tag string) bool {

	fields := strings.Fields(tags)
	for _, field := range fields {
		if field == tag {
			return true
		}
	}
	return false
}
