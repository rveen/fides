package fides

import (
	"errors"
	"math"
)

func SemiconductorFIT(comp *Component, mission *Mission) (float64, error) {

	vfactor := 1.0
	if comp.Class == "D" && comp.Imax < 1 && !(contains(comp.Tags, "tvs") || contains(comp.Tags, "zener")) {

		if comp.Vmax == 0 || math.IsNaN(comp.Vmax) {
			return math.NaN(), errors.New("Vmax not set")
		}

		if comp.V == 0 || math.IsNaN(comp.V) {
			return math.NaN(), errors.New("working V not set")
		}

		vfactor = PiThermal_voltageFactor(comp.V, comp.Vmax)
	}

	lth := Lchip_th(comp)
	if lth < 0 {
		return math.NaN(), errors.New("Missing data for lchip(th) calculation")
	}

	p := NewPackage(comp.Package)
	if p == nil {
		return math.NaN(), errors.New("Package not found: [" + comp.Package + "]")
	}
	comp.Np = p.Npins
	lrh, ltc, lts, lm := p.FitBase()
	if lrh < 0 || math.IsNaN(lrh) {
		return math.NaN(), errors.New("Missing data for lpkg(rh,tc...) calculation for package: [" + p.Name + "]")
	}

	var factor float64

	// fmt.Printf("semi: lth %f, vfactor %f, lrh %f, ltc %f, lts %f, lm %f\n", lth, vfactor, lrh, ltc, lts, lm)

	for _, ph := range mission.Phases {

		// TODO Add disipated power
		tj := ph.Tamb

		pi := lth*PiThermal(0.7, tj, ph.On)*vfactor +
			ltc*PiTCCase(ph.NCycles, ph.Duration, ph.Tdelta, ph.Tmax) +
			lts*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			lm*PiMech(ph.Grms)

		// Proportion of time in this phase
		pi *= ph.Duration / mission.Ttotal

		// Stress factors and sensibility
		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		pi *= ifactor

		factor += pi
	}

	return factor * PiPM() * PiProcess(), nil
}

func Lchip_th(c *Component) float64 {

	var base float64

	nfactor := 1.0
	if c.N > 1 {
		nfactor = math.Sqrt(float64(c.N))
	}

	// ICs

	if c.Class == "U" {

		base = 0.086

		for _, tag := range c.Tags {

			switch tag {
			case "opto", "optocoupler":
				if contains(c.Tags, "photodiode") {
					base = 0.05
				} else {
					base = 0.11
				}
				break
			case "mixed", "analog": // mixed or analog asic
				base = 0.123
				break
			case "fpga", "cpld", "pal":
				base = 0.076
				break
			case "microprocessor", "microcontroller", "dsp", "complex": // complex asic
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

		for _, tag := range c.Tags {

			switch tag {

			case "gan":
				base = 0.3033

			case "gaas":
				base = 0.3756

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

		if contains(c.Tags, "zener") {

			if c.Pmax < 1.5 {
				base = 0.008
			} else {
				base = 0.0954
			}
		} else if contains(c.Tags, "tvs") {
			if c.Pmax < 3000 {
				base = 0.021
			} else {
				base = 1.498
			}
		} else {
			// diode, signal or rectifier

			if c.Imax < 1 {
				base = 0.0044
			} else if c.Imax < 3 {
				base = 0.01
			} else {
				base = 0.1574
			}
		}

		//log.Printf("D, base=%f, nfactor=%f\n", base, nfactor)

		return base * nfactor

	}

	return -1
}
