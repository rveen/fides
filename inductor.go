package fides

// InductorFIT
func InductorFIT(comp *Component, mission *Mission) (float64, error) {

	var fit, nfit float64

	l0, ea, lth, ltc, lmech, tdelta, cs := lbase_inductor(comp.Tags)

	/*
		cs := Cs(comp.Class, comp.Tags)
		if math.IsNaN(cs) {
			return math.NaN(), errors.New("Missing data for stress sensibility calculation")
		}*/

	for _, ph := range mission.Phases {

		lth2 := lth
		if !ph.On {
			lth2 = 0
		}

		nfit = l0 * ph.Duration / 8760.0 * (lmech*PiMech(ph.Grms) +
			ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lth2*Arrhenius25(ea, ph.Tamb+tdelta))

		nfit *= PiInduced(ph.On, comp.Tags, cs)

		fit += nfit
	}

	return fit, nil
}

// Returns l0, ea, lth, ltc, lmech, tdelta, Cs
func lbase_inductor(tags []string) (float64, float64, float64, float64, float64, float64, float64) {

	if containsTag(tags, "trafo") {
		if containsTag(tags, "power") {
			return 0.25, 0.15, 0.15, 0.69, 0.16, 30, 6.13
		} else {
			return 0.125, 0.15, 0.01, 0.73, 0.26, 10, 5.63
		}
	}

	if containsTag(tags, "multilayer") || containsTag(tags, "ferrite") {
		return 0.05, 0.15, 0.71, 0.28, 0.01, 10, 4.3
	}

	if containsTag(tags, "power") {
		return 0.05, 0.15, 0.09, 0.79, 0.12, 30, 6.58
	}

	// Default ww low power
	return 0.025, 0.15, 0.01, 0.73, 0.26, 10, 4.73
}
