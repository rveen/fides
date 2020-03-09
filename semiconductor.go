package fides

import (
	"math"
)

func FIT(comp *Component, mission *Mission) {

	switch comp.Type {

	case "ic":
	case "asic":
	case "dicrete":
	case "led":
	case "optocoupler":
	case "resistor":
	case "fuse":
	case "cap_ceramic":
	case "cap_alu":
	case "cap_tant":
	case "inductor":
	case "piezo":
	case "relay":
	case "switch":
	case "connector":

	}

}

func SemiconductorFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	ratio := comp.V / comp.Vmax
	if comp.Type != "diode_signal" {
		ratio = 1
	}

	lrh, lcase, lsolder, lmech := Lcase(comp.Package)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = ph.Time / 8760.0 *
				(Lchip(comp.Type, comp.N)*PiThermal_semiconductor(ratio, ph.Tmax*comp.Rth*comp.P) +
					lcase*PiCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
					lsolder*PiSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lrh*PiRH(ph.RH, ph.Tamb) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = ph.Time / 8760.0 *
				(lcase*PiCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
					lsolder*PiSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, CSensibility(comp.Class, comp.Type))

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
}

func Lcase(typ string) (float64, float64, float64, float64) {

	switch typ {

	case "THT, metal":
		return 0, 0.0101, 0.0505, 0.00101
	case "THT, signal, plastic":
		return 0.0310, 0.00110, 0.0055, 0.00011
	case "THT, power, plastic":
		return 0.0589, 0.00303, 0.01515, 0.0003
	case "SMD, signal, llead, plastic":
		return 0.0055, 0.00057, 0.00285, 0.000057
	case "SMD, signal, clead, plastic":
		return 0.0124, 0.00091, 0.00455, 0.00009
	case "SMD, medium, llead, plastic":
		return 0.0126, 0.00091, 0.00455, 0.000091
	case "SMD, power, llead, plastic":
		return 0.0335, 0.00413, 0.02065, 0.00041
	case "SMD, glass":
		return 0, 0.00781, 0.03905, 0.00078
	case "ISOTOP":
		return 0.99, 0.03333, 0.16665, 0.0033
	}
	return 0, 0, 0, 0
}

func Lchip(typ string, n int) float64 {

	var base float64

	switch typ {
	case "diode_signal":
		base = 0.0044
	case "rectifier":
		base = 0.01
	case "zener":
		return 0.008
	case "zener_pow":
		base = 0.0954
	case "tvs":
		base = 0.021
	case "tvs_pow":
		base = 1.498
	case "bipolar":
		base = 0.0138
	case "mos":
		base = 0.0145
	case "fet":
		base = 0.0143
	case "bipolar_pow":
		base = 0.0478
	case "mos_pow":
		base = 0.0202
	case "igbt":
		base = 0.3021
	case "triac":
		base = 0.1976
	case "rectifier_pow":
		base = 0.1574

	default:
		return 0.01
	}

	if n < 2 {
		return base
	}

	return base * math.Sqrt(float64(n))
}
