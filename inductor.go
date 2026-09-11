package fides

// InductorFIT
func InductorFIT(comp *Component, mission *Mission) (float64, error) {

	fit, ea, lth, ltc, lm, tdelta, cs := lbase_inductor(comp.Tags)
	placement := piPlacement(comp.Tags)
	var factor float64

	for _, ph := range mission.Phases {

		pi := 0.0
		if ph.On {
			pi = lth * Arrhenius25(ea, ph.Tamb+tdelta)
		}

		pi += ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lm*PiMech(ph.Grms)

		// Proportion of time in this phase
		pi *= ph.Duration / mission.Ttotal

		// Stress factors and sensibility
		pi *= piInducedCs(cs, placement, ph)

		factor += pi
	}

	return fit * factor * PiPM() * PiProcess(), nil
}

// Returns l0, ea, lth, ltc, lmech, tdelta, Cs (FIDES 2022, pp. 161-162)
func lbase_inductor(tags []string) (float64, float64, float64, float64, float64, float64, float64) {

	if contains(tags, "trafo") {
		if contains(tags, "power") {
			return 0.25, 0.15, 0.15, 0.69, 0.16, 30, csens(6, 7, 4)
		} else {
			return 0.125, 0.15, 0.01, 0.73, 0.26, 10, csens(6, 5, 3)
		}
	}

	if contains(tags, "multilayer") || contains(tags, "ferrite_bead") {
		return 0.05, 0.15, 0.71, 0.28, 0.01, 10, csens(4, 6, 1)
	}

	if contains(tags, "power") {
		return 0.05, 0.15, 0.09, 0.79, 0.12, 30, csens(7, 6, 3)
	}

	// Default ww low power
	return 0.025, 0.15, 0.01, 0.73, 0.26, 10, csens(5, 4, 4)
}
