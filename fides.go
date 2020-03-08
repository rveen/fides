// Fides 2009A library
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

func SemiconductorFIT(comp *Component, mission *Mission, ratio float64) float64 {

	var fit, nfit float64

	if comp.Type != "diode_signal" {
		ratio = 1
	}

	lrh, lcase, lsolder, lmech := Lcase(comp.Package)

	for _, ph := range mission.Phases {

		nfit = ph.Time/8760.0*
			Lchip(comp.Type, comp.N)*PiThermal(ratio, ph.Tmax*comp.Rth*comp.Power) +
			lcase*PiCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
			lsolder*PiSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH(ph.RH, ph.Tamb) +
			lmech*PiMech(ph.Grms)

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 5.2)

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

// [1] ratio is 1 for component other than signal diodes
func PiThermal(ratio, tj float64) float64 {
	return math.Max(0.056, math.Pow(ratio, 2.4)) * math.Exp(11604*0.7*(1/293-1/(tj+273)))
}

func PiMech(grms float64) float64 {
	return math.Pow(grms*2, 1.5)
}

func PiRH(rh, temp float64) float64 {
	return math.Pow(rh/70, 4.4) * math.Exp(11604*0.9*(1/293-1/(temp+273)))
}

func PiCase(nc int, time, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(tdelta/20, 4) * math.Exp(1414*(1/313-1/(tmax+273)))
}

func PiSolder(nc int, time, phi, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(math.Min(phi, 2)/2, 1.3) * math.Pow(tdelta/20, 1.9) * math.Exp(1414*(1/313-1/(tmax+273)))
}

func PiPM() float64 {
	return 0.5
}

func PiProcess() float64 {
	return 2
}

func CSensibility(typ string) float64 {
	switch typ {
	case "discrete_semiconductor":
		return 5.2
	}
	return 1
}

func PiInduced(on, analog, itf, power bool, csensibility float64) float64 {
	return math.Pow(PiPlacement(analog, itf, power)*PiApplication(on)*PiRuggedising(), 0.511*math.Log(csensibility))
}

func PiPlacement(analog bool, itf bool, power bool) float64 {

	if !analog {
		if itf {
			return 1.6
		}
		return 1
	}

	if !power {
		if itf {
			return 2
		}
		return 1.3
	}

	if !itf {
		return 1.6
	}
	return 2.5
}

// Mess
func PiApplication(on bool) float64 {
	if on {
		return 5.1
	}
	return 3.1
}

// Return max value (very controlled process)
func PiRuggedising() float64 {
	return 2
}
