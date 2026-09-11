package fides

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// LEDFIT returns the FIDES 2022 FIT of a light-emitting diode (class D, tag
// led; pp. 141-143).
//
// The chip is a colour LED unless tagged white. The package basic failure
// rates depend on the maximum direct current Imax and the package (see
// ledPackage). comp.T is the junction temperature rise over ambient, P·Rth.
func LEDFIT(comp *Component, mission *Mission) (float64, error) {
	if comp.Tmax == 0 || math.IsNaN(comp.Tmax) {
		return math.NaN(), errors.New("LED: Tmax (junction temperature rating) not set")
	}

	lrh, lcase, lsolder, lmech := ledPackage(comp)

	l0th := 0.01
	if contains(comp.Tags, "white") {
		l0th = 0.05
	}
	if comp.N > 1 {
		l0th *= math.Sqrt(float64(comp.N))
	}

	dt := comp.T
	if math.IsNaN(dt) {
		dt = 0
	}

	var factor float64
	for _, ph := range mission.Phases {
		tj := ph.Tamb + dt
		if tj > comp.Tmax && ph.On {
			return math.NaN(), fmt.Errorf("LED %s: junction temperature %.0f°C exceeds Tmax %.0f°C",
				comp.Name, tj, comp.Tmax)
		}

		pi := l0th*PiThermal(0.4, tj, ph.On) +
			lcase*PiTCCase(ph.NCycles, ph.Duration, ph.Tdelta, ph.Tmax) +
			lsolder*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			lmech*PiMech(ph.Grms)

		pi *= ph.Duration / mission.Ttotal

		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		pi *= ifactor
		factor += pi
	}
	return factor * PiPMActive() * PiProcess(), nil
}

// ledPackage returns λ0RH, λ0TCyCase, λ0TCySolderJoints and λ0Mech of a LED
// package (p. 142). LEDs with a maximum direct current Imax of 150 mA or more
// are SMD, plastic or ceramic (tag ceramic). Below 150 mA the package is
// T1 (T1-x), HIGHFLUX, CHIP, PLCC or MINI (all the same rates), ROUND, LGA
// (plastic or ceramic), or any other (plastic or ceramic).
func ledPackage(c *Component) (float64, float64, float64, float64) {

	ceramic := contains(c.Tags, "ceramic")

	if c.Imax >= 0.15 {
		if ceramic {
			return 0.0031, 0.0042, 0.1470, 0.0735
		}
		return 0.0031, 0.0042, 0.0420, 0.0064
	}

	pkg := strings.ToUpper(c.Package)
	has := func(prefixes ...string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(pkg, p) {
				return true
			}
		}
		return false
	}

	switch {
	case has("T1", "HIGHFLUX", "CHIP", "PLCC", "MINI"):
		return 0.0034, 0.0104, 0.0520, 0.0052
	case has("ROUND"):
		return 0.0034, 0.0104, 0.1560, 0.0624
	case has("LGA") && ceramic:
		return 0.0034, 0.0104, 0.3640, 0.1820
	case has("LGA"):
		return 0.0034, 0.0104, 0.2080, 0.0832
	case ceramic:
		return 0.0034, 0.0104, 0.3640, 0.1820
	}
	return 0.0034, 0.0104, 0.1560, 0.0624
}
