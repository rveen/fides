package fides

import (
	"errors"
	"math"
)

// piezoType is a row of the quartz crystal table of FIDES 2022 (p. 163).
type piezoType struct {
	Name                  string
	osc                   bool
	l0, th, tcy, mech, rh float64
	eos, mos, tos         float64
}

// PiezoType returns the quartz crystal type of a component: an oscillator
// if tagged osc or oscillator, else a resonator; through-hole if tagged tht
// or its package is through-hole, else surface-mounted.
func PiezoType(c *Component) piezoType {

	osc := contains(c.Tags, "osc") || contains(c.Tags, "oscillator")
	smd := !contains(c.Tags, "tht") && (contains(c.Tags, "smd") || IsSmd(c))

	switch {
	case osc && smd:
		return piezoType{"quartz crystal oscillator, SMD", true, 1.63, 0.31, 0.53, 0.07, 0.09, 7, 9, 3}
	case osc:
		return piezoType{"quartz crystal oscillator, through-hole", true, 1.6, 0.32, 0.42, 0.14, 0.12, 7, 9, 3}
	case smd:
		return piezoType{"quartz crystal resonator, SMD", false, 0.79, 0.16, 0.59, 0.15, 0.10, 2, 10, 5}
	}
	return piezoType{"quartz crystal resonator, through-hole", false, 0.82, 0.16, 0.46, 0.27, 0.11, 2, 10, 5}
}

// PiezoFIT returns the FIDES 2022 FIT of a quartz crystal resonator or
// oscillator (pp. 163-164). The output current I of an oscillator counts
// when it is 80 % of Imax or more.
func PiezoFIT(comp *Component, mission *Mission) (float64, error) {

	pt := PiezoType(comp)
	cs := csens(pt.eos, pt.mos, pt.tos)
	placement := piPlacement(comp.Tags)

	// Πrating_EL
	ratingEL := 1.0
	if pt.osc && comp.Imax > 0 && !math.IsNaN(comp.I) && comp.I >= 0.8*comp.Imax {
		ratingEL = 5
	}

	var factor float64
	for _, ph := range mission.Phases {

		// General rule
		if ph.Tamb > comp.Tmax {
			return math.NaN(), errors.New("Using component above its Tmax")
		}

		th := 0.0
		if ph.On {
			// Πrating_TH
			ratingTH := 1.0
			if ph.Tamb >= comp.Tmax-40 {
				ratingTH = 5
			}
			th = pt.th * ratingTH * ratingEL
		}

		pi := th +
			pt.tcy*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			pt.mech*PiMech(ph.Grms) +
			pt.rh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On)

		pi *= ph.Duration / mission.Ttotal
		pi *= piInducedCs(cs, placement, ph)
		factor += pi
	}

	return pt.l0 * factor * PiPM() * PiProcess(), nil
}
