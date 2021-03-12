package fides

import (
	"log"
	"math"
)

func SemiconductorFIT(comp *Component, mission *Mission) float64 {

	// Type: D / Q
	// Subtype: TVS, ZENER, MOS, JFET, IGBT, TRIAC, THYRISTOR
	//
	// Imax (D)
	// Pmax (ZENER, TVS)
	// Vmax (D)

	var fit, nfit float64

	if comp.Vmax == 0 || math.IsNaN(comp.Vmax) {
		log.Println("Vmax not set in semiconductor", comp.Name)
		return math.NaN()
	}

	ratio := comp.V / comp.Vmax
	if comp.Class == "D" && comp.Type != "" {
		ratio = 1
	}

	lrh, lcase, lsolder, lmech := Lcase_semi(comp.Package)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = ph.Time / 8760.0 *
				(Lchip(comp)*PiThermal_semiconductor(ratio, ph.Tmax /* *comp.Rth*comp.P */) +
					lcase*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
					lsolder*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lrh*PiRH(ph.RH, ph.Tamb) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = ph.Time / 8760.0 *
				(lcase*PiTCCase(ph.NCycles, ph.Time, ph.Tdelta, ph.Tmax) +
					lsolder*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 5.2)

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
}

func Lchip(c *Component) float64 {

	var base float64

	// Transistors

	if c.Class == "Q" {

		switch c.Type {

		case "IGBT":
			base = 0.3021
		case "TRIAC":
			base = 0.1976
		case "JFET":
			base = 0.0143

		case "MOS":

			if c.Pmax > 5 {
				base = 0.0202
			} else {
				base = 0.0145
			}

		default:

			// bipolar silicon transistor
			if c.Pmax > 5 {
				base = 0.0478
			} else {
				base = 0.0138
			}
		}

		if c.N == 1 {
			return base
		}

		return base * math.Sqrt(float64(c.N))
	}

	// Diodes

	switch c.Type {

	case "ZENER":
		if c.Pmax < 1.5 {
			base = 0.08
		} else {
			base = 0.0954
		}

	case "TVS":
		if c.Pmax < 3000 {
			base = 0.021
		} else {
			base = 1.498
		}

	default:

		// diode, signal or rectifier
		if c.Imax < 1 {
			base = 0.01
		} else if c.Imax < 3 {
			base = 0.0044
		} else {
			base = 0.1574
		}

	}

	if c.N == 1 {
		return base
	}

	return base * math.Sqrt(float64(c.N))
}
