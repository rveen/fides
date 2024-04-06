package fides

//import "log"

func OptoFIT(comp *Component, mission *Mission) float64 {

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

	for _, ph := range mission.Phases {

		// TODO Add disipated power
		tj := ph.Tamb

		nfit = ph.Time / 8760.0 * (lth*PiThermal_opto(tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
			(lts+ltc_chip)*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(ph.RH, ph.Tamb, ph.On) +
			(lm+lm_chip)*PiMech(ph.Grms))

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, cs)

		fit += nfit
	}

	return fit

}
