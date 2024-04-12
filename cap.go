package fides

import (
	"errors"
	"math"
)

// CapacitorFIT
//
// Attributes that need to be known for capacitors:
// basic type: ceramic, aluminium, tantalum
// ceramic: if flex or not
// ceramic: working voltage, temperature coefficient
// aluminium: dry/wet
// tantalum: dry/wet, glass/elastomer seal, tantalum/silver case, radial/smd/axial
//
// Codification of attributes in Component
// (1) Detection of dielectric (EIA codes): LnL
// (2) TANT / TAN, ELE / ELEC
// (3) FLEX
// (4) Default for tantalum: smd dry.
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
		return math.NaN(), errors.New("unknown capacitor type")
	}

	var l0, ea, sref, lth, ltc, lm, fit, nfit float64

	if ctype == "cer" {

		flex := containsTag(comp.Tags, "flex")
		type1 := containsTag(comp.Tags, "np0") || containsTag(comp.Tags, "c0g") || containsTag(comp.Tags, "type1")
		topend := containsTag(comp.Tags, "topend")
		l0, ea, sref, lth, ltc, lm = lbase_capCer(flex, type1, topend, comp.Value, comp.Vmax)

	} else if ctype == "alu" {

		dry := containsTag(comp.Tags, "dry") || containsTag(comp.Tags, "solid")
		l0, ea, sref, lth, ltc, lm = lbase_capAlu(dry)

	} else { // tant

		smd := containsTag(comp.Tags, "smd") || IsSmd(comp)
		l0, ea, sref, lth, ltc, lm = lbase_capTant(comp.Tags, smd)

	}

	cs := Cs(comp.Class, comp.Tags)
	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	for _, ph := range mission.Phases {

		// General rule
		if ph.Tamb > comp.Tmax {
			return math.NaN(), errors.New("Using component above its Tmax")
		}

		if ph.On {
			nfit = l0 * ph.Duration / 8760.0 *
				(lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax) +
					ltc*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lm*PiMech(ph.Grms))
		} else {
			nfit = l0 * ph.Duration / 8760.0 * (lm * PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.Tags, cs)

		fit += nfit
	}

	return fit, nil

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

	wet := containsTag(tags, "wet")            // solid is default
	glass := containsTag(tags, "glass_sealed") // default is anything else
	silver := containsTag(tags, "silver_case")
	axial := containsTag(tags, "axial")

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
