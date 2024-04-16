package fides

import (
	"errors"
	"math"
)

// PCB connectors, SMD, less that one insertion/year
func ConnectorFIT(comp *Component, mission *Mission) (float64, error) {

	var fit, nfit float64

	piReport := 6.0

	if contains(comp.Tags, "smd") || IsSmd(comp) {
		piReport = 10
	}

	if comp.N < 1 {
		return math.NaN(), errors.New("Connector with 0 contacts")
	}
	l0connector := 0.1 * piReport * 0.2 * math.Pow(float64(comp.N), 0.5)

	cs := Cs(comp.Class, comp.Tags)
	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	for _, ph := range mission.Phases {

		// Thermal
		nfit = 0.58 * PiThermal(0.1, ph.Tamb, ph.On)

		// Thermal cycling case
		nfit += 0.04 * PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax)

		// Mechanical
		nfit += 0.05 * PiMech(ph.Grms)

		// RH
		nfit += 0.13 * PiRH(0.8, ph.RH, ph.Tamb)

		// Chemical
		if !ph.IP {
			nfit += 0.2 * ph.SalinePollution * ph.AmbientPollution * ph.ApplicationPollution
		}

		// Time
		nfit *= ph.Duration / 8760.0 * l0connector

		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		nfit *= ifactor

		fit += nfit
	}

	fit *= PiPM() * PiProcess()

	return fit, nil
}
