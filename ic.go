package fides

import (
	"math"
)

func IcFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	for _, ph := range mission.Phases {

		nfit = ph.Time / 8760.0

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 6.05)

		fit += nfit
	}

	return fit * PiPM() * PiProcess()

}

// PiThermal for ICs is 0 in non-operating mode
func PiThermal_ic(tj float64, on bool) float64 {
	if !on {
		return 0
	}
	return math.Exp(11604.0 * 0.7 * (1.0/293.0 - 1.0/(tj-273)))
}

func PiRH_ic(rh, temp float64, on bool) float64 {
	if on {
		return 0
	}
	return math.Pow(rh/70, 4.4) * math.Exp(11604.0*0.9*(1.0/293.0-1.0/(temp-273)))
}

// give package and pins, return: l0rh, l0tc_case, l0tc_solder, l0mech
func Lbase_ic(pkg string, n int) (float64, float64, float64, float64) {

	var arh, brh, atc, btc, ats, bts, am, bm float64

	switch pkg {

	case "pdip":
		arh = -5.88
		brh = 0.94
		atc = 9.85
		btc = 1.35
		ats = 8.24
		bts = 1.35
		am = 12.85
		bm = 1.35

	case "cdip":
		atc = 6.77
		btc = 1.35
		if n < 21 {
			ats = 5.16
			am = 8.38
		} else {
			ats = 4.47
			am = 7.69
		}
		bts = 1.35
		bm = 1.35

	case "pqfp":
		arh = 11.16
		brh = 1.76
		atc = 12.41
		btc = 1.46
		if n < 241 {
			ats = 10.8
			am = 14.71
		} else {
			ats = 10.11
			am = 14.02
		}
		bts = 1.46
		bm = 1.46

	case "sqfp":
		fallthrough
	case "tqfp":
		fallthrough
	case "vqfp":
		fallthrough
	case "lqfp":
		arh = 7.75
		brh = 1.13
		atc = 8.57
		btc = 0.73
		if n < 121 {
			ats = 6.96
			am = 11.57
		} else {
			ats = 5.57
			am = 10.18
		}
		bts = 0.73
		bm = 0.73

	}

	if arh != 0 {
		arh = math.Exp(arh) * math.Pow(float64(n), brh)
	}
	atc = math.Exp(atc) * math.Pow(float64(n), btc)
	arh = math.Exp(ats) * math.Pow(float64(n), bts)
	am = math.Exp(am) * math.Pow(float64(n), bm)

	return arh, atc, ats, am
}

func Lbase_th(typ string) float64 {

	switch typ {
	case "fpga":
		fallthrough
	case "cpld":
		return 0.166
	case "analog":
		fallthrough
	case "mixed":
		return 0.123
	case "digital":
		return 0.021
	case "cpu":
		fallthrough
	case "dsp":
		return 0.075
	case "eprom":
		fallthrough
	case "eeprom":
		fallthrough
	case "flash":
		return 0.06
	case "dram":
		return 0.047
	case "sram":
		return 0.055

	default:
		return -1
	}
}
