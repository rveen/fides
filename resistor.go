package fides

import (
	"errors"
	"fmt"
	"math"

	"github.com/rveen/electronics"
)

func ResistorFIT(comp *Component, mission *Mission) (float64, error) {

	var fit, nfit float64

	l0, A, lth, ltc, lmech, lrh := Lbase_resistor(comp)
	A = A

	if (comp.P == 0 || math.IsNaN(comp.P)) && comp.V != 0 {
		if comp.Value == 0 {
			comp.Value = 0.001
		}
		comp.P = comp.V * comp.V / comp.Value
	}
	if comp.P == 0 || math.IsNaN(comp.P) {
		return math.NaN(), errors.New("actual power is not set")
	}
	if comp.Pmax == 0 || math.IsNaN(comp.Pmax) {
		return math.NaN(), errors.New("Pmax is not set")
	}
	if comp.P > comp.Pmax {
		return math.NaN(), errors.New("Actual power exceeds Pmax")
	}

	cs := Cs(comp.Class, comp.Tags)
	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	comp.Rtha = electronics.Rth(comp.Package)
	if comp.Rtha == 0 {
		return math.NaN(), errors.New("Rth could not be set for this package")
	}
	tdelta := comp.P * comp.Rtha

	for _, ph := range mission.Phases {

		if ph.On {

			tc := ph.Tamb + tdelta
			if tc >= comp.Tmax {
				s := fmt.Sprintf("Component temperature (%f ºC) exceeds its Tmax (%f ºC), P=%f W, Rth=%f ºC/W ", tc, comp.Tmax, comp.P, comp.Rtha)
				return math.NaN(), errors.New(s)
			}

			nfit = l0 * ph.Duration / 8760.0 *
				(lth*Arrhenius25(0.15, tc /*+A*comp.P/comp.Pmax*/) +
					ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms) +
					lrh*PiRH(0.9, ph.RH, ph.Tamb))
		} else {
			nfit = l0 * ph.Duration / 8760.0 * (lmech * PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.Tags, cs)
		fit += nfit
	}

	return fit, nil
}

// returning lbase, A, lth, ltc, lmech, lrh
func Lbase_resistor(c *Component) (float64, float64, float64, float64, float64, float64) {

	if containsTag(c.Tags, "melf") {
		return 0.1, 85, 0.04, 0.89, 0.01, 0.06
	}

	if containsTag(c.Tags, "thick") {
		if c.P >= 1 {
			return 0.4, 130, 0.04, 0.89, 0.01, 0.06
		}
		return 0.1, 85, 0.04, 0.89, 0.01, 0.06
	}

	if containsTag(c.Tags, "tht") {
		if c.Value < 10000 {
			return 0.14, 85, 0.18, 0.43, 0.08, 0.31
		} else if c.Value < 100000 {
			return 0.18, 85, 0.12, 0.44, 0.07, 0.37
		} else {
			return 0.21, 85, 0.08, 0.45, 0.06, 0.41
		}
	}

	if containsTag(c.Tags, "network") && containsTag(c.Tags, "smd") {
		return 0.01 * math.Sqrt(float64(c.N)), 70, 0.01, 0.97, 0.01, 0.01
	}

	if containsTag(c.Tags, "ww") {
		if c.P >= 1 {
			return 0.4, 130, 0.01, 0.97, 0.01, 0.01
		}
		return 0.03, 30, 0.02, 0.96, 0.01, 0.01
	}

	if containsTag(c.Tags, "potmeter") && !containsTag(c.Tags, "ww") {
		return 0.3, 65, 0.42, 0.35, 0.22, 0.01

	}

	// Default: thin film resistor
	if c.Value < 10000 {
		return 0.18, 85, 0.14, 0.53, 0.07, 0.26
	} else if c.Value < 100000 {
		return 0.21, 85, 0.10, 0.54, 0.06, 0.30
	} else {
		return 0.25, 85, 0.07, 0.55, 0.05, 0.33
	}
}
