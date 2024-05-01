package fides

//	"errors"
import (
	"errors"
	"math"
)

// Current not considered currently
func PiezoFIT(comp *Component, mission *Mission) (float64, error) {

	fit, lth, ltc, lm, lrh := lbase_piezo(comp)
	var factor float64

	for _, ph := range mission.Phases {

		// General rule
		if ph.Tamb > comp.Tmax {
			return math.NaN(), errors.New("Using component above its Tmax")
		}

		tfactor := 1.0
		if ph.On == false {
			tfactor = 0
		} else if ph.Tamb+40 > comp.Tmax {
			tfactor = 5
		}

		pi := lth*tfactor +
			ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lm*PiMech(ph.Grms) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On)

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

	return fit * factor * PiPM() * PiProcess(), nil

}

func lbase_piezo(c *Component) (float64, float64, float64, float64, float64) {

	osc := contains(c.Tags, "osc") || contains(c.Tags, "oscillator")

	if contains(c.Tags, "smd") {
		if osc {
			return 1.63, 0.31, 0.53, 0.07, 0.09
		} else {
			return 0.79, 0.16, 0.59, 0.15, 0.1
		}
	}

	// THT
	if osc {
		return 1.6, 0.32, 0.42, 0.14, 0.12
	}

	return 0.82, 0.16, 0.46, 0.27, 0.11
}
