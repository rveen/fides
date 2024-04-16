package fides

import "math"

func OptoFIT(comp *Component, mission *Mission) (float64, error) {

	var fit, nfit float64

	lth := Lchip_th(comp)
	ltc_chip := 0.021
	lm_chip := 0.011

	if contains(comp.Tags, "photodiode") {
		ltc_chip = 0.01
		lm_chip = 0.005
	}

	p := NewPackage(comp.Package)
	lrh, ltc, lts, lm := p.FitBase()

	for _, ph := range mission.Phases {

		// TODO Add disipated power
		tj := ph.Tamb

		// Physical
		nfit = ph.Duration / 8760.0 * (lth*PiThermal(0.4, tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Duration, ph.Tdelta, ph.Tmax) +
			(lts+ltc_chip)*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			(lm+lm_chip)*PiMech(ph.Grms))

		// Induced
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
