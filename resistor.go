package fides

import (
	"errors"
	"fmt"
	"math"

	"github.com/rveen/electronics"
)

func ResistorFIT(comp *Component, mission *Mission) (float64, error) {

	// The A parameter from FIDES 2022 is ignored, as we are calculating
	// the actual temperature of the part based on Rtha.
	fit, _, lth, ltc, lm, lrh := Lbase_resistor(comp)
	var factor float64

	// networks
	if comp.N > 1 {
		fit *= math.Sqrt(float64(comp.N))
	}

	// Power calculation. Priority: P, V²/R, I*R
	if (comp.P == 0 || math.IsNaN(comp.P)) && comp.V != 0 {
		if comp.Value == 0 {
			comp.Value = 0.001
		}
		comp.P = comp.V * comp.V / comp.Value
	}
	if comp.P == 0 || math.IsNaN(comp.P) {
		if comp.I != 0 && math.IsNaN(comp.I) {
			comp.P = comp.Value * comp.I
		}
	}
	if comp.P == 0 || math.IsNaN(comp.P) {
		return math.NaN(), errors.New("Power cannot be calculated. Either set P, V or I")
	}

	if comp.Pmax == 0 || math.IsNaN(comp.Pmax) {
		return math.NaN(), errors.New("Pmax is not set")
	}
	if comp.P > comp.Pmax {
		s := fmt.Sprintf("Actual power (%f W) exceeds its Pmax (%f W) R=%g", comp.P, comp.Pmax, comp.Value)
		return math.NaN(), errors.New(s)
	}

	comp.Rtha = electronics.Rth(comp.Package)
	if comp.Rtha == 0 {
		return math.NaN(), errors.New("Rth could not be set for this package")
	}
	tdelta := comp.P * comp.Rtha

	for _, ph := range mission.Phases {

		tc := ph.Tamb + tdelta
		if tc >= comp.Tmax && ph.On {
			s := fmt.Sprintf("Component temperature (%f ºC) exceeds its Tmax (%f ºC), P=%f W, Rth=%f ºC/W ", tc, comp.Tmax, comp.P, comp.Rtha)
			return math.NaN(), errors.New(s)
		}

		pi := 0.0
		if ph.On {
			pi = lth * Arrhenius25(0.15, tc)
		}

		pi += ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			lm*PiMech(ph.Grms)

		// Proportion of time in this phase
		pi *= ph.Duration / 8760.0

		// Stress factors and sensibility
		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		pi *= ifactor

		factor += pi
	}

	return fit * factor * PiPM() * PiProcess(), nil
}

// Return base values: l0, A, lth, ltc, lmech, lrh
//
// Networks are not included here, they should be marked as thin or thick an then
// above the l0*sqrt(comp.N) will take care of the number of resistors in the package.
//
// Default: smd, thin
func Lbase_resistor(c *Component) (float64, float64, float64, float64, float64, float64) {

	if contains(c.Tags, "melf") {
		return 0.1, 85, 0.04, 0.89, 0.01, 0.06
	}

	if contains(c.Tags, "thick") {
		if c.P >= 1 {
			return 0.4, 130, 0.04, 0.89, 0.01, 0.06
		}
		return 0.01, 70, 0.01, 0.97, 0.01, 0.01
	}

	if contains(c.Tags, "tht") {
		if c.Value < 10000 {
			return 0.14, 85, 0.18, 0.43, 0.08, 0.31
		} else if c.Value < 100000 {
			return 0.18, 85, 0.12, 0.44, 0.07, 0.37
		} else {
			return 0.21, 85, 0.08, 0.45, 0.06, 0.41
		}
	}

	if contains(c.Tags, "ww") {
		if c.P >= 1 {
			return 0.4, 130, 0.01, 0.97, 0.01, 0.01
		}
		return 0.03, 30, 0.02, 0.96, 0.01, 0.01
	}

	if contains(c.Tags, "potmeter") && !contains(c.Tags, "ww") {
		return 0.3, 65, 0.42, 0.35, 0.22, 0.01

	}

	// Default: smd thin film resistor
	if c.Value < 10000 {
		return 0.18, 85, 0.14, 0.53, 0.07, 0.26
	} else if c.Value < 100000 {
		return 0.21, 85, 0.10, 0.54, 0.06, 0.30
	} else {
		return 0.25, 85, 0.07, 0.55, 0.05, 0.33
	}
}
