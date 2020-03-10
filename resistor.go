package fides

import (
	"math"
)

func ResistorFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	l0, A, lth, ltc, lmech, lrh := Lbase_resistor(comp.Type, comp.N, comp.Value)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_resistor(0.15, comp.T+A*comp.P/comp.Pmax) +
					ltc*PiThermalCycling(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms) +
					lrh*PiRH(ph.RH, ph.Tamb))
		} else {
			nfit = l0 * ph.Time / 8760.0 * (lmech * PiMech(ph.Grms))
		}
		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, CSensibility(comp.Class, comp.Type))

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
}

// returning lbase, A, lth, ltc, lmech, lrh
func Lbase_resistor(typ string, n int, v float64) (float64, float64, float64, float64, float64, float64) {

	switch typ {

	case "melf":
		return 0.1, 85, 0.04, 0.89, 0.01, 0.06
	case "power_film":
		return 0.4, 130, 0.04, 0.89, 0.01, 0.06
	case "ww_precision":
		return 0.03, 30, 0.02, 0.96, 0.01, 0.01
	case "ww_power":
		return 0.4, 130, 0.01, 0.97, 0.01, 0.01
	case "pot_cermet":
		return 0.3, 65, 0.42, 0.35, 0.22, 0.01
	case "chip":
		return 0.01, 70, 0.01, 0.97, 0.01, 0.01
	case "network_smd":
		return 0.01 * math.Sqrt(float64(n)), 70, 0.01, 0.97, 0.01, 0.01
	case "metal_foil_precision_smd":
		if v < 10000 {
			return 0.18, 85, 0.14, 0.53, 0.07, 0.26
		} else if v < 100000 {
			return 0.21, 85, 0.10, 0.54, 0.06, 0.30
		} else {
			return 0.25, 85, 0.07, 0.55, 0.05, 0.33
		}
	case "metal_foil_precision_tht":
		if v < 10000 {
			return 0.14, 85, 0.18, 0.43, 0.08, 0.31
		} else if v < 100000 {
			return 0.18, 85, 0.12, 0.44, 0.07, 0.37
		} else {
			return 0.21, 85, 0.08, 0.45, 0.06, 0.41
		}

	}
	return 0, 0, 0, 0, 0, 0
}
