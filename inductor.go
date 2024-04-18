package fides

import "math"

// InductorFIT
func InductorFIT(comp *Component, mission *Mission) (float64, error) {

	fit, ea, lth, ltc, lm, tdelta, _ := lbase_inductor(comp.Tags)
	var factor float64

	for _, ph := range mission.Phases {

		pi := 0.0
		if ph.On {
			pi = lth * Arrhenius25(ea, ph.Tamb+tdelta)
		}

		pi += ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lm*PiMech(ph.Grms)

		// Proportion of time in this phase
		pi *= ph.Duration / 8760.0

		// Stress factors and sensibility
		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		pi *= ifactor

		factor += pi
	}

	return fit * factor * PiPM() * PiProcess(), nil
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

	if contains(tags, "multilayer") || contains(tags, "ferrite_bead") {
		return 0.05, 0.15, 0.71, 0.28, 0.01, 10, 4.3
	}

	if contains(tags, "power") {
		return 0.05, 0.15, 0.09, 0.79, 0.12, 30, 6.58
	}

	// Default ww low power
	return 0.025, 0.15, 0.01, 0.73, 0.26, 10, 4.73
}
