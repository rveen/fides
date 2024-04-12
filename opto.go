package fides

import (
	"errors"
	"math"
)

func OptoFIT(comp *Component, mission *Mission) (float64, error) {

	var fit, nfit float64

	lth := Lchip_th(comp)
	ltc_chip := 0.021
	lm_chip := 0.011

	if containsTag(comp.Tags, "photodiode") {
		ltc_chip = 0.01
		lm_chip = 0.005
	}

	lrh, ltc, lts, lm := Lbase_pkg(comp.Package)
	cs := Cs(comp.Class, comp.Tags)
	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	for _, ph := range mission.Phases {

		// TODO Add disipated power
		tj := ph.Tamb

		nfit = ph.Duration / 8760.0 * (lth*PiThermal(0.4, tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Duration, ph.Tdelta, ph.Tmax) +
			(lts+ltc_chip)*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			(lm+lm_chip)*PiMech(ph.Grms))

		nfit *= PiInduced(ph.On, comp.Tags, cs)

		fit += nfit
	}

	return fit, nil

}
