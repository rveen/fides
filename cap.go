package fides

import (
	"errors"
	"math"
)

func CapacitorFIT(comp *Component, mission *Mission) (float64, error) {

	// Vmax and V are needed for capacitors

	if comp.Vmax == 0 || math.IsNaN(comp.Vmax) {
		return math.NaN(), errors.New("Vmax not set")
	}

	if comp.V == 0 || math.IsNaN(comp.V) {
		return math.NaN(), errors.New("working V not set")
	}

	if comp.V > comp.Vmax {
		return math.NaN(), errors.New("working V higher than limit Vmax")
	}

	// Determine basic type (alu, tant, cer)
	ctype := capType(comp.Tags)
	if ctype == "" {
		return math.NaN(), errors.New("unknown capacitor type " + ctype)
	}

	var fit, factor, ea, sref, lth, ltc, lm float64

	if ctype == "cer" {

		flex := contains(comp.Tags, "flex")
		type1 := contains(comp.Tags, "np0") || contains(comp.Tags, "c0g") || contains(comp.Tags, "type1")
		topend := contains(comp.Tags, "topend")
		fit, ea, sref, lth, ltc, lm = lbase_capCer(flex, type1, topend, comp.Value, comp.Vmax)

	} else if ctype == "alu" {

		dry := contains(comp.Tags, "dry") || contains(comp.Tags, "solid")
		fit, ea, sref, lth, ltc, lm = lbase_capAlu(dry)

	} else { // tant

		smd := contains(comp.Tags, "smd") || IsSmd(comp)
		fit, ea, sref, lth, ltc, lm = lbase_capTant(comp.Tags, smd)

	}

	for _, ph := range mission.Phases {

		// General rule
		if ph.Tamb > comp.Tmax {
			return math.NaN(), errors.New("Using component above its Tmax")
		}

		pi := lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax, ph.On) +
			ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
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

	return fit * factor * PiPM() * PiProcess(), nil
}

// https://en.wikipedia.org/wiki/Ceramic_capacitor
//
// X = -55, Y = -30
// 5 = 85, 6 = 105, 7 = 125, 8 = 150
// R = 15%, S= 22%, U=+22/-56, V=+22/-82

var cerCapTags = []string{"x5r", "x5s", "x6r", "x6s", "x7r", "x7s", "x8r", "x8s", "np0", "c0g", "y5v"}

// Can be improved to return tolerance and temperature limits
func capType(tags []string) string {

	if contains(tags, "tant") || contains(tags, "tantalium") {
		return "tant"
	}

	if contains(tags, "alu") || contains(tags, "elco") {
		return "alu"
	}

	for _, tag := range cerCapTags {
		if contains(tags, tag) {
			return "cer"
		}
	}

	return ""
}

func lbase_capCer(flex, type1, topend bool, value, vmax float64) (float64, float64, float64, float64, float64, float64) {

	// CV product class
	cvp := value * vmax
	cat := 3

	if type1 {
		if cvp <= 5e-8 {
			cat = 1
		} else if cvp < 1e-6 || !topend {
			cat = 2
		}
	} else {
		if cvp <= 5e-6 {
			cat = 1
		} else if cvp < 1e-4 || !topend {
			cat = 2
		}
	}

	if type1 {
		switch cat {
		case 1:
			return 0.03, 0.1, 0.3, 0.7, 0.28, 0.02
		case 2:
			return 0.05, 0.1, 0.3, 0.7, 0.28, 0.02
		default:
			return 0.4, 0.1, 0.3, 0.69, 0.26, 0.05
		}
	} else {
		if flex && cat > 1 {
			return 0.15, 0.1, 0.3, 0.7, 0.28, 0.02
		}

		switch cat {
		case 1:
			return 0.08, 0.1, 0.3, 0.7, 0.28, 0.02
		case 2:
			return 0.15, 0.1, 0.3, 0.7, 0.28, 0.02
		default:
			return 1.2, 0.1, 0.3, 0.44, 0.51, 0.02
		}
	}
}

func lbase_capAlu(solid bool) (float64, float64, float64, float64, float64, float64) {

	if !solid {
		return 0.21, 0.4, 0.5, 0.85, 0.14, 0.01
	}
	return 0.4, 0.4, 0.55, 0.85, 0.14, 0.01
}

func lbase_capTant(tags []string, smd bool) (float64, float64, float64, float64, float64, float64) {

	wet := contains(tags, "wet")            // solid is default
	glass := contains(tags, "glass_sealed") // default is anything else
	silver := contains(tags, "silver_case")
	axial := contains(tags, "axial")

	if wet {
		if glass {
			if silver {
				return 0.33, 0.15, 0.6, 0.81, 0.01, 0.18
			} else {
				return 0.05, 0.15, 0.6, 0.88, 0.04, 0.08
			}
		}
		return 0.77, 0.15, 0.6, 0.87, 0.01, 0.12
	}

	// Solid

	if smd {
		return 0.54, 0.15, 0.4, 0.84, 0.14, 0.02
	}
	if axial {
		return 0.25, 0.15, 0.4, 0.94, 0.04, 0.02
	}
	return 1.09, 0.15, 0.4, 0.86, 0.12, 0.02

}
