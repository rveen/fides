package fides

import "math"

// InductorFIT
func InductorFIT(comp *Component, mission *Mission) (float64, error) {

	var fit, nfit float64

	l0, ea, lth, ltc, lmech, tdelta, _ := lbase_inductor(comp.Tags)

	for _, ph := range mission.Phases {

		lth2 := lth
		if !ph.On {
			lth2 = 0
		}

		nfit = l0 * ph.Duration / 8760.0 * (lmech*PiMech(ph.Grms) +
			ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lth2*Arrhenius25(ea, ph.Tamb+tdelta))

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

// Returns l0, ea, lth, ltc, lmech, tdelta, Cs
func lbase_inductor(tags []string) (float64, float64, float64, float64, float64, float64, float64) {

	if contains(tags, "trafo") {
		if contains(tags, "power") {
			return 0.25, 0.15, 0.15, 0.69, 0.16, 30, 6.13
		} else {
			return 0.125, 0.15, 0.01, 0.73, 0.26, 10, 5.63
		}
	}

	if contains(tags, "multilayer") || contains(tags, "ferrite") {
		return 0.05, 0.15, 0.71, 0.28, 0.01, 10, 4.3
	}

	if contains(tags, "power") {
		return 0.05, 0.15, 0.09, 0.79, 0.12, 30, 6.58
	}

	// Default ww low power
	return 0.025, 0.15, 0.01, 0.73, 0.26, 10, 4.73
}
