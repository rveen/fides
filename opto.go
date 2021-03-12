package fides

func OptoFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	// TODO: different for optos with photodiode
	lth := 0.11
	ltc_chip := 0.021
	lm_chip := 0.011

	comp.Package, comp.N = splitPkg(comp.Package)

	lrh, ltc, lts, lm := Lbase_case(comp.Package, comp.N)

	for _, ph := range mission.Phases {

		// What is the junction temperature?
		// TODO
		tj := ph.Tamb

		nfit = ph.Time / 8760.0 * (lth*PiThermal_ic(tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
			(lts+ltc_chip)*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(ph.RH, ph.Tamb, ph.On) +
			(lm+lm_chip)*PiMech(ph.Grms))

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 5.2)
		// log.Println("nfit", ph.Name, nfit, lth, lth*PiThermal_ic(tj, ph.On))
		fit += nfit
	}

	return fit * PiPM() * PiProcess()

}
