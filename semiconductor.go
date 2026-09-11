package fides

import (
	"errors"
	"math"
)

// IsPowerTransistor reports whether a component falls under the FIDES 2022
// model for discrete power semiconductors (pp. 138-140): silicon MOS rated
// 5 W or more, and IGBTs.
func IsPowerTransistor(c *Component) bool {
	if c.Class != "Q" {
		return false
	}
	mos := contains(c.Tags, "mos") || contains(c.Tags, "mosfet")
	return contains(c.Tags, "igbt") || mos && c.Pmax >= 5
}

// piThermalPower is ΠThermal of a power transistor: Arrhenius relative to
// TRef = 60 °C, with the junction temperature at most 175 °C (p. 140).
func piThermalPower(tj float64, on bool) float64 {
	if !on {
		return 0
	}
	tj = math.Min(tj, 175)
	return math.Exp(InvBoltzman * 0.7 * (1.0/(60+273) - 1/(tj+273)))
}

// SemiconductorFIT returns the FIDES 2022 FIT of a diode, transistor or
// integrated circuit (pp. 123-129, 133-140).
func SemiconductorFIT(comp *Component, mission *Mission) (float64, error) {

	if contains(comp.Tags, "gan") || contains(comp.Tags, "gaas") {
		return math.NaN(), errors.New("GaN and GaAs components have RF/microwave FIDES models, which are not supported")
	}
	power := IsPowerTransistor(comp)

	vfactor := 1.0
	if comp.Class == "D" && comp.Imax < 1 && !(contains(comp.Tags, "tvs") || contains(comp.Tags, "zener")) {

		if comp.Vmax == 0 || math.IsNaN(comp.Vmax) {
			return math.NaN(), errors.New("Vmax not set")
		}

		// V = 0 (forward-biased or unbiased diode) is valid: the voltage
		// factor has a floor.
		if math.IsNaN(comp.V) {
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

	// Self-heating: comp.T is the junction temperature rise over ambient,
	// P·Rth, supplied by the caller.
	dt := comp.T
	if math.IsNaN(dt) {
		dt = 0
	}

	var factor float64

	// fmt.Printf("semi: lth %f, vfactor %f, lrh %f, ltc %f, lts %f, lm %f\n", lth, vfactor, lrh, ltc, lts, lm)

	for _, ph := range mission.Phases {

		tj := ph.Tamb + dt

		th := PiThermal(0.7, tj, ph.On) * vfactor
		if power {
			th = piThermalPower(tj, ph.On)
		}

		pi := lth*th +
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

	pw := 1.0
	if power {
		pw = PiPW()
	}
	return factor * pw * PiPMActive() * PiProcess(), nil
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
			case "mixed", "analog":
				// Analog and mixed circuits (FIDES 2022, p. 128)
				base = 0.086
			case "fpga", "cpld", "pal":
				base = 0.076
			case "microprocessor", "microcontroller", "dsp", "complex": // complex asic
				base = 0.075
			case "flash", "eprom", "eeprom":
				base = 0.06
			case "sram":
				base = 0.053
			case "dram":
				base = 0.047
			case "digital": // also simple digital asic"
				base = 0.021
			}
		}

		// Default value is for analog, mixed, interface
		return base * nfactor
	}

	// Transistors

	if c.Class == "Q" {

		for _, tag := range c.Tags {

			switch tag {

			case "igbt":
				// IGBTs: power semiconductor model (p. 139)
				base = 0.56
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

		// bipolar silicon transistor, if no other type matched
		if base == 0 {
			if c.Pmax >= 5 {
				base = 0.0478
			} else {
				base = 0.0138
			}
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
