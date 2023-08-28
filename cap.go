package fides

import (
	"log"
	"math"
	"strings"
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
func CapacitorFIT(comp *Component, mission *Mission) float64 {

	fields := strings.Fields(strings.ToUpper(comp.Description))

	if contains(fields, "TANT") || contains(fields, "TAN") {
		return CapacitorTantFIT(comp, mission, "dry, smd")
	}

	if contains(fields, "ELEC") || contains(fields, "ELCO") || contains(fields, "ELE") {
		return CapacitorAluFIT(comp, mission, true)
	}

	flex := contains(fields, "FLEX")
	class := capClass(fields[1])

	return CapacitorCeramicFIT(comp, mission, flex, class)
}

// https://es.slideshare.net/RandallGhany/class-1-and-class-2-mlccs
func capClass(s string) int {

	if s == "" {
		return 3
	}

	switch s[0] {
	case 'U':
		fallthrough
	case 'C':
		fallthrough
	case 'N':
		return 1

	case 'X':
		return 2
	}
	return 3
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CapacitorCeramicFIT(comp *Component, mission *Mission, flex bool, class int) float64 {

	var fit, nfit float64

	l0, ea, sref, lth, ltc, lmech := Lbase_capCer(flex, class, comp.Value, comp.Vmax)

	// log.Println(comp.Name, comp.Vmax, comp.V, l0, ea, sref, lth, lmech)

	if comp.Vmax == 0 || math.IsNaN(comp.Vmax) {
		log.Println("Vmax not set in capacitor", comp.Name)
		return math.NaN()
	}

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax) +
					ltc*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = l0 * ph.Time / 8760.0 * (lmech * PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 6.05)

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
}

func Lbase_capCer(polymer_terminations bool, class int, value, vmax float64) (float64, float64, float64, float64, float64, float64) {

	// type I or type II
	type1 := class == 1

	// CV product class
	cvp := value * vmax

	if type1 {
		if cvp < 1e-9 {
			return 0.03, 0.1, 0.3, 0.7, 0.28, 0.02
		} else if cvp > 1e-7 {
			return 0.4, 0.1, 0.3, 0.69, 0.26, 0.05
		}
		return 0.05, 0.1, 0.3, 0.7, 0.28, 0.02
	} else {
		if cvp < 1e-7 {
			return 0.08, 0.1, 0.3, 0.7, 0.28, 0.02
		} else if polymer_terminations || cvp <= 1e-5 {
			return 0.15, 0.1, 0.3, 0.7, 0.28, 0.02
		}
		return 1.2, 0.1, 0.3, 0.44, 0.51, 0.02
	}
}

func CapacitorAluFIT(comp *Component, mission *Mission, dry bool) float64 {

	var fit, nfit float64

	l0, ea, sref, lth, ltc, lmech := Lbase_capAlu(dry)

	log.Printf("CapAluFIT %s: %f/%f\n", comp.Name, comp.V, comp.Vmax)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax) +
					ltc*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = l0 * ph.Time / 8760.0 * (lmech * PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 6.4)

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
}

func Lbase_capAlu(solid bool) (float64, float64, float64, float64, float64, float64) {

	if !solid {
		return 0.21, 0.4, 0.5, 0.85, 0.14, 0.01
	}
	return 0.4, 0.4, 0.55, 0.85, 0.14, 0.01
}

func CapacitorTantFIT(comp *Component, mission *Mission, typ string) float64 {

	var fit, nfit float64

	l0, ea, sref, lth, ltc, lmech := Lbase_capTant(typ)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax) +
					ltc*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = l0 * ph.Time / 8760.0 * (lmech * PiMech(ph.Grms))
		}
		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, 6.95)

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
}

func Lbase_capTant(typ string) (float64, float64, float64, float64, float64, float64) {

	switch typ {
	case "wet, silver case, elastomer seal":
		return 0.77, 0.15, 0.6, 0.87, 0.01, 0.12
	case "wet, silver case, glass seal":
		return 0.33, 0.15, 0.6, 0.81, 0.01, 0.18
	case "wet, tantalum case, glass seal":
		return 0.05, 0.15, 0.6, 0.88, 0.04, 0.08
	case "dry, radial":
		return 1.09, 0.15, 0.4, 0.86, 0.12, 0.02
	case "dry, smd":
		return 0.54, 0.15, 0.4, 0.84, 0.14, 0.02
	case "dry, axial":
		return 0.25, 0.15, 0.4, 0.94, 0.04, 0.02
	}

	return 0, 0, 0, 0, 0, 0

}
