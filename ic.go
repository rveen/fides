package fides

import (
	"log"
	"strconv"
	"strings"
)

func IcFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	comp.Package, comp.N = splitPkg(comp.Package)

	if comp.Package == "" {
		log.Printf("IcFIT: ERROR: no package defined for %s\n", comp.Name)
	}
	if comp.N == 0 {
		log.Printf("IcFIT: ERROR: Number of pins is 0 for %s (package %s)\n", comp.Name, comp.Package)
	}

	lth := Lbase_ic_th(comp.Type)

	lrh, ltc, lts, lm := Lbase_case(comp.Package, comp.N)

	// log.Println("Lbase_case(ic)", comp.Name, comp.Type, comp.Package, comp.N, lth, lrh, ltc, lts, lm)

	for _, ph := range mission.Phases {

		// What is the junction temperature?
		// TODO
		tj := ph.Tamb

		nfit = ph.Time / 8760.0 * (lth*PiThermal_ic(tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
			lts*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(ph.RH, ph.Tamb, ph.On) +
			lm*PiMech(ph.Grms))

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 6.3)
		// log.Println("nfit", ph.Name, nfit, lth, lth*PiThermal_ic(tj, ph.On))
		fit += nfit
	}

	return fit * PiPM() * PiProcess()

}

func Lbase_ic_th(typ string) float64 {

	switch typ {
	case "fpga", "cpld":
		return 0.166
	case "analog", "interface", "mixed":
		return 0.123
	case "digital":
		return 0.021
	case "cpu", "dsp":
		return 0.075
	case "eprom", "eeprom", "flash":
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

	if s == "sot223" {
		return "sot", 3
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
