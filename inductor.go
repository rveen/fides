package fides

// InductorFIT
func InductorFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	l0, ea, lth, ltc, lmech, tdelta, _ := Lbase_inductor(comp.Tags)

	cs := Cs(comp.Class, comp.Tags)

	for _, ph := range mission.Phases {

		lth2 := lth
		if !ph.On {
			lth2 = 0
		}

		nfit = l0 * ph.Time / 8760.0 * (lmech*PiMech(ph.Grms) +
			ltc*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lth2*PiThermal_resistor(ea, ph.Tamb+tdelta))

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, cs)

		fit += nfit
	}

	return fit
}

// Returns l0, ea, lth, ltc, lmech, tdelta, Cs
func Lbase_inductor(typ string) (float64, float64, float64, float64, float64, float64, float64) {

	switch typ {
	case "trafo":
		return 0.125, 0.15, 0.01, 0.73, 0.26, 10, 6.9
	case "trafo_power":
		return 0.25, 0.15, 0.15, 0.69, 0.16, 30, 6.8
	case "multilayer", "ferrite":
		return 0.05, 0.15, 0.71, 0.28, 0.01, 10, 4.4
	case "wirewound", "choke":
		return 0.025, 0.15, 0.01, 0.73, 0.26, 10, 4.05
	case "pot", "wirewound_power":
		return 0.05, 0.15, 0.09, 0.79, 0.12, 30, 8.05
	}
	return -1, -1, -1, -1, -1, -1, -1
}
