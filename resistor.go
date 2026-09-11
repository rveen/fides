package fides

import (
	"errors"
	"fmt"
	"math"
)

// resistorType is a row of the resistor tables of FIDES 2022 (pp. 146-147).
type resistorType struct {
	Name              string
	l0                float64 // basic failure rate
	a                 float64 // temperature rise at rated power, °C
	th, tcy, mech, rh float64 // weights γ of the physical stresses
	eos, mos, tos     float64 // relative sensitivities to overstress
}

// ResistorType returns the FIDES 2022 resistor type of a component, from its
// tags, rated power and value (pp. 146-147):
//
//   - melf: "MiniMELF" fixed resistor
//   - network: SMD resistor network (λ0 × √N, N resistors)
//   - thick: SMD thick film, or high-power thick film if rated 1 W or more
//   - ww: wirewound precision, or high-power wirewound if rated 1 W or more
//   - pot, potmeter, potentiometer, variable: non-wirewound potentiometer
//   - otherwise: thin film, precision, high-stability, by value; SMD unless
//     tagged tht
func ResistorType(c *Component) resistorType {

	switch {
	case contains(c.Tags, "melf"):
		return resistorType{"MiniMELF", 0.1, 85, 0.04, 0.89, 0.01, 0.06, 4, 2, 4}
	case contains(c.Tags, "network"):
		return resistorType{"network, SMD", 0.01, 70, 0.01, 0.97, 0.01, 0.01, 3, 5, 3}
	case contains(c.Tags, "thick"):
		if c.Pmax >= 1 {
			return resistorType{"high-power thick film", 0.4, 130, 0.04, 0.89, 0.01, 0.06, 2, 4, 1}
		}
		return resistorType{"thick film, SMD", 0.01, 70, 0.01, 0.97, 0.01, 0.01, 4, 3, 5}
	case contains(c.Tags, "ww"):
		if c.Pmax >= 1 {
			return resistorType{"high-power wirewound", 0.4, 130, 0.01, 0.97, 0.01, 0.01, 2, 4, 1}
		}
		return resistorType{"wirewound, precision", 0.3, 30, 0.02, 0.96, 0.01, 0.01, 2, 1, 3}
	case contains(c.Tags, "pot") || contains(c.Tags, "potmeter") ||
		contains(c.Tags, "potentiometer") || contains(c.Tags, "variable"):
		return resistorType{"potentiometer, non-wirewound", 0.3, 65, 0.42, 0.35, 0.22, 0.01, 1, 5, 2}
	}

	tht := contains(c.Tags, "tht")
	switch {
	case c.Value < 10e3 && tht:
		return resistorType{"thin film, through-hole, < 10 kΩ", 0.14, 85, 0.18, 0.43, 0.08, 0.31, 5, 5, 4}
	case c.Value < 10e3:
		return resistorType{"thin film, SMD, < 10 kΩ", 0.18, 85, 0.14, 0.53, 0.07, 0.26, 5, 5, 4}
	case c.Value < 100e3 && tht:
		return resistorType{"thin film, through-hole, 10 kΩ to 100 kΩ", 0.18, 85, 0.12, 0.44, 0.07, 0.37, 5, 5, 4}
	case c.Value < 100e3:
		return resistorType{"thin film, SMD, 10 kΩ to 100 kΩ", 0.21, 85, 0.10, 0.54, 0.06, 0.30, 5, 5, 4}
	case tht:
		return resistorType{"thin film, through-hole, > 100 kΩ", 0.21, 85, 0.08, 0.45, 0.06, 0.41, 5, 5, 4}
	}
	return resistorType{"thin film, SMD, > 100 kΩ", 0.25, 85, 0.07, 0.55, 0.05, 0.33, 5, 5, 4}
}

// ResistorFIT returns the FIDES 2022 FIT of a resistor (pp. 146-148).
func ResistorFIT(comp *Component, mission *Mission) (float64, error) {

	rt := ResistorType(comp)
	l0 := rt.l0

	// Several resistors in one package (networks)
	if comp.N > 1 {
		l0 *= math.Sqrt(float64(comp.N))
	}

	// Power calculation. Priority: P, V²/R, I²·R. NaN means not set.
	if comp.Value == 0 {
		comp.Value = 0.001
	}
	if math.IsNaN(comp.P) && !math.IsNaN(comp.V) {
		comp.P = comp.V * comp.V / comp.Value
	}
	if math.IsNaN(comp.P) && !math.IsNaN(comp.I) {
		comp.P = comp.I * comp.I * comp.Value
	}
	if math.IsNaN(comp.P) {
		return math.NaN(), errors.New("Power cannot be calculated. Either set P, V or I")
	}

	if comp.Pmax == 0 || math.IsNaN(comp.Pmax) {
		return math.NaN(), errors.New("Pmax is not set")
	}
	if comp.P > comp.Pmax {
		s := fmt.Sprintf("Actual power (%f W) exceeds its Pmax (%f W) R=%g", comp.P, comp.Pmax, comp.Value)
		return math.NaN(), errors.New(s)
	}

	// The resistor temperature is Tambient + A·P/Prated (p. 148)
	tdelta := rt.a * comp.P / comp.Pmax
	cs := csens(rt.eos, rt.mos, rt.tos)
	placement := piPlacement(comp.Tags)

	var factor float64
	for _, ph := range mission.Phases {

		tc := ph.Tamb + tdelta
		if tc >= comp.Tmax && ph.On {
			s := fmt.Sprintf("Component temperature (%f ºC) exceeds its Tmax (%f ºC), P=%f W of %f W", tc, comp.Tmax, comp.P, comp.Pmax)
			return math.NaN(), errors.New(s)
		}

		pi := 0.0
		if ph.On {
			pi = rt.th * Arrhenius25(0.15, tc)
		}
		pi += rt.tcy*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			rt.rh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			rt.mech*PiMech(ph.Grms)

		pi *= ph.Duration / mission.Ttotal
		pi *= piInducedCs(cs, placement, ph)
		factor += pi
	}

	return l0 * factor * PiPM() * PiProcess(), nil
}
