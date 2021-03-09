package fides

import (
	"math"
	"strconv"
	"strings"
)

func IcFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	if comp.N == 0 {
		comp.Package, comp.N = splitPkg(comp.Package)
	}

	lth := Lbase_ic_th(comp.Type)

	lrh, ltc, lts, lm := Lbase_ic(comp.Package, comp.N)

	for _, ph := range mission.Phases {

		// What is the junction temperature?
		// TODO
		tj := ph.Tamb

		nfit = ph.Time / 8760.0 * (lth*PiThermal_ic(tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
			lts*PiThermalCycling(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH_ic(ph.RH, ph.Tamb, ph.On) +
			lm*PiMech(ph.Grms))

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 6.3)
		// log.Println("nfit", ph.Name, nfit, lth, lth*PiThermal_ic(tj, ph.On))
		fit += nfit
	}

	return fit * PiPM() * PiProcess()

}

// PiThermal for ICs is 0 in non-operating mode
func PiThermal_ic(tj float64, on bool) float64 {
	if !on {
		return 0
	}
	return math.Exp(11604.0 * 0.7 * (1.0/293.0 - 1.0/(tj+273)))
}

func PiRH_ic(rh, temp float64, on bool) float64 {
	if on {
		return 0
	}
	return math.Pow(rh/70, 4.4) * math.Exp(11604.0*0.9*(1.0/293.0-1.0/(temp+273)))
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
		if n < 44 || n > 304 {
			return -1, -1, -1, -1
		} else if n < 241 {
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
		if n < 32 || n > 208 {
			return -1, -1, -1, -1
		} else if n < 121 {
			ats = 6.96
			am = 11.57
		} else {
			ats = 5.57
			am = 10.18
		}
		bts = 0.73
		bm = 0.73

	case "cerpack": // TODO
	case "plcc": // TODO
	case "jclcc": // TODO
	case "clcc": // TODO
	case "soj": // TODO

	case "so":
		fallthrough
	case "sow":
		fallthrough
	case "sop":
		fallthrough
	case "sol":
		fallthrough
	case "soic":

		arh = 8.23
		brh = 1.17
		atc = 13.35
		btc = 2.18
		bts = 2.18
		bm = 2.18

		if n < 8 || n > 32 {
			return -1, -1, -1, -1
		} else if n < 17 {
			ats = 11.75
			am = 16.36
		} else if n < 21 {
			ats = 11.06
			am = 15.66
		} else if n < 32 {
			ats = 10.36
			am = 14.97
		} else {
			ats = 1014
			am = 14.75
		}
	case "sot":
		fallthrough // TEMPORAL : TODO
	case "tsop":
		if n < 5 || n > 56 {
			return -1, -1, -1, -1
		} else if n < 17 {
			ats = 7.44
			am = 12.05
		} else if n < 33 {
			ats = 6.05
			am = 10.66
		} else if n < 45 {
			ats = 5.83
			am = 10.44
		} else {
			ats = 5.36
			am = 9.97
		}

		arh = 6.21
		brh = 0.97
		atc = 9.05
		btc = 0.76
		bts = 0.76
		bm = 0.76

	case "ssop":
		fallthrough
	case "vsop":
		fallthrough
	case "qsop":
		arh = 11.95
		brh = 2.23
		atc = 16.28
		btc = 2.6
		ats = 14.67
		bts = 2.6
		am = 19.28
		bm = 2.6

	case "tssop":
		fallthrough
	case "msop":
		if n >= 8 && n < 29 {
			ats = 13.95
			am = 18.56
		} else if n > 28 && n < 49 {
			ats = 13.21
			am = 17.86
		} else if n == 56 {
			ats = 12.56
			am = 17.17
		} else if n == 64 {
			ats = 12.16
			am = 16.76
		} else {
			return -1, -1, -1, -1
		}

		arh = 11.57
		brh = 2.22
		atc = 15.56
		btc = 2.66
		bts = 2.66
		bm = 2.66

	case "qfn":
		fallthrough
	case "dfn":
		fallthrough
	case "mlf":
		if n < 8 || n > 72 {
			return -1, -1, -1, -1
		} else if n < 25 {
			ats = 8.12
			am = 11.34
		} else if n < 57 {
			ats = 7.9
			am = 11.12
		} else {
			ats = 7.71
			am = 10.93
		}
		arh = 8.97
		brh = 1.14
		atc = 11.2
		btc = 1.21
		bts = 1.14
		bm = 1.21

	case "pbga_0_8": // TODO
	case "pbga_0_8_flex": // TODO
	case "pbga_1_0": // TODO
	case "pbga_1_27": // TODO
	case "powerbga": // TODO
	case "cbga": // TODO
	case "dbga": // TODO
	case "cicga": // TODO
	case "cpga": // TODO

	}

	if arh != 0 {
		arh = math.Exp(-arh) * math.Pow(float64(n), brh)
	}
	atc = math.Exp(-atc) * math.Pow(float64(n), btc)
	ats = math.Exp(-ats) * math.Pow(float64(n), bts)
	am = math.Exp(-am) * math.Pow(float64(n), bm)

	return arh, atc, ats, am
}

func Lbase_ic_th(typ string) float64 {

	switch typ {
	case "fpga":
		fallthrough
	case "cpld":
		return 0.166
	case "analog":
		fallthrough
	case "interface":
		fallthrough // this is my addition (rolf)
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

func splitPkg(s string) (string, int) {

	if s == "sot23-5" {
		return "sot", 5
	}

	var sb strings.Builder
	var nb strings.Builder

	for i, c := range s {
		if c >= '0' && c <= '9' {
			s = s[i:]
			break
		}
		sb.WriteRune(c)
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		nb.WriteRune(c)
	}

	pkg := sb.String()
	if pkg == "miniso" {
		pkg = "tssop"
	}

	n, _ := strconv.Atoi(nb.String())
	return pkg, n
}
