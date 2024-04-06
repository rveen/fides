package fides

import (
	"math"
	"strings"
)

func SemiconductorFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	vfactor := 1.0
	if comp.Class == "D" && comp.Imax < 1 && (containsTag(comp.Tags, "tvs") || containsTag(comp.Tags, "zener")) {
		vfactor = PiThermal_voltageFactor(comp.V, comp.Vmax)
	}

	lth := Lchip_th(comp)
	lrh, ltc, lts, lm := Lbase_pkg(comp.Package)
	cs := Cs(comp.Class, comp.Tags)

	for _, ph := range mission.Phases {

		// TODO Add disipated power
		tj := ph.Tamb

		nfit = ph.Time / 8760.0 * (lth*PiThermal_ic(tj, ph.On)*vfactor +
			ltc*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
			lts*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(ph.RH, ph.Tamb, ph.On) +
			lm*PiMech(ph.Grms))

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, cs)

		fit += nfit
	}

	return fit
}

func Lchip_th(c *Component) float64 {

	var base float64

	nfactor := 1.0
	if c.N > 1 {
		nfactor = math.Sqrt(float64(c.N))
	}

	tags := strings.Fields(strings.ToLower(c.Tags))

	// ICs

	if c.Class == "U" {

		base = 0.086

		for _, tag := range tags {

			switch tag {
			case "opto", "optocoupler":
				if contains(tags, "photodiode") {
					base = 0.05
				} else {
					base = 0.11
				}
				break
			case "mixed": // mixed or analog asic
				base = 0.123
				break
			case "fpga", "cpld", "pal":
				base = 0.076
				break
			case "microprocessor", "microcontroller", "dsp", "asic": // complex asic
				base = 0.075
				break
			case "flash", "eprom", "eeprom":
				base = 0.06
				break
			case "sram":
				base = 0.053
				break
			case "dram":
				base = 0.047
				break
			case "digital": // also simple digital asic"
				base = 0.021
				break
			}
		}

		// Default value is for analog, mixed, interface
		return base * nfactor
	}

	// Transistors

	if c.Class == "Q" {

		for _, tag := range tags {

			switch tag {

			case "igbt":
				base = 0.3021
				if c.Pmax >= 5 {
					base = 0.56
				}
				break
			case "triac", "thyristor":
				base = 0.1976
				break
			case "jfet":
				base = 0.0143
				break
			case "mos", "mosfet":

				if c.Pmax >= 5 {
					base = 0.56
				} else {
					base = 0.0145
				}
				break
			}

		}

		// bipolar silicon transistor
		if c.Pmax >= 5 {
			base = 0.0478
		} else {
			base = 0.0138
		}

		return base * nfactor
	}

	// Diodes

	if c.Class == "D" {

		for _, tag := range tags {

			switch tag {

			case "zener":
				if c.Pmax < 1.5 {
					base = 0.008
				} else {
					base = 0.0954
				}
				break

			case "tvs":
				if c.Pmax < 3000 {
					base = 0.021
				} else {
					base = 1.498
				}
				break
			}
		}

		// diode, signal or rectifier
		if c.Imax < 1 {
			base = 0.0044
		} else if c.Imax < 3 {
			base = 0.01
		} else {
			base = 0.1574
		}

		return base * nfactor

	}

	return -1
}
