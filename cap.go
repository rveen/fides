package fides

import (
	"math"
)

func CapacitorCeramicFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	l0, A, lth, ltc, lmech, lrh := Lbase_resistor(comp.Type, comp.N, comp.Value)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_resistor(comp.T, A*comp.P/comp.Pmax) +
					ltc*PiSolder(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax, ph.CycleDuration) +
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

func Lbase_CerCap(polymer_terminations bool, tc, value, vmax float64) (float64, float64, float64, float64, float64, float64) {

	// type I or type II
	type1 := true
	if math.IsNaN(tc) {
		type1 = false
	}

	// CV product class
	cvp := value * vmax

	if type1 {
		if cvp < 1e-9 {
			return 0.03, 0.1, 0.3, 0.7, 0.28, 0.02
		} else if cvp > 1e-7 {
			return 0.4, 0.1, 0.3, 0.69, 0.26, 0.05
		}
		return 0.05, 0.1, 0.3, 0.7, 0.28, 0.02
	} else {
		if cvp < 1e-7 {
			return 0.08, 0.1, 0.3, 0.7, 0.28, 0.02
		} else if polymer_terminations || cvp <= 1e-5 {
			return 0.15, 0.1, 0.3, 0.7, 0.28, 0.02
		}
		return 1.2, 0.1, 0.3, 0.44, 0.51, 0.02
	}
}
