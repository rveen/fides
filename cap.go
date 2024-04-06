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

	fields := strings.Fields(comp.Tags)

	if contains(fields, "tant") || contains(fields, "tantalium") {
		return CapacitorTantFIT(comp, mission, "dry, smd")
	}

	if contains(fields, "alu") || contains(fields, "elco") {
		return CapacitorAluFIT(comp, mission, true)
	}

	// TODO Better detection of ceramic types
	if contains(fields, "x7r") || contains(fields, "x7s") || contains(fields, "x5r") || contains(fields, "np0") || contains(fields, "type1") || contains(fields, "type2") {
		return CapacitorCeramicFIT(comp, mission)
	}

	return math.NaN()
}

func CapacitorCeramicFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	log.Printf("ceramic capacitor %e F %f V", comp.Value, comp.Vmax)

	flex := containsTag(comp.Tags, "flex")
	type1 := containsTag(comp.Tags, "np0") || containsTag(comp.Tags, "type1")
	topend := containsTag(comp.Tags, "topend")

	l0, ea, sref, lth, ltc, lmech := Lbase_capCer(flex, type1, topend, comp.Value, comp.Vmax)
	cs := Cs(comp.Class, comp.Tags)

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

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, cs)

		fit += nfit
	}

	return fit
}

func Lbase_capCer(flex, type1, topend bool, value, vmax float64) (float64, float64, float64, float64, float64, float64) {

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

func CapacitorAluFIT(comp *Component, mission *Mission) float64 {

	var fit, nfit float64

	dry := containsTag(comp.Tags, "dry") || containsTag(comp.Tags, "solid")
	l0, ea, sref, lth, ltc, lmech := Lbase_capAlu(dry)
	cs := Cs(comp.Class, comp.Tags)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax) +
					ltc*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = l0 * ph.Time / 8760.0 * (lmech * PiMech(ph.Grms))
		}

		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, cs)

		fit += nfit
	}

	return fit
}

func Lbase_capAlu(solid bool) (float64, float64, float64, float64, float64, float64) {

	if !solid {
		return 0.21, 0.4, 0.5, 0.85, 0.14, 0.01
	}
	return 0.4, 0.4, 0.55, 0.85, 0.14, 0.01
}

func CapacitorTantFIT(comp *Component, mission *Mission, typ string) float64 {

	var fit, nfit float64

	smd := SMD(comp.Package)

	l0, ea, sref, lth, ltc, lmech := Lbase_capTant(comp.Tags, smd)
	cs := Cs(comp.Class, comp.Tags)

	for _, ph := range mission.Phases {

		if ph.On {
			nfit = l0 * ph.Time / 8760.0 *
				(lth*PiThermal_cap(ea, ph.Tamb, sref, comp.V/comp.Vmax) +
					ltc*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
					lmech*PiMech(ph.Grms))
		} else {
			nfit = l0 * ph.Time / 8760.0 * (lmech * PiMech(ph.Grms))
		}
		nfit *= PiInduced(ph.On, comp.IsAnalog, comp.IsInterface, comp.IsPower, cs)

		fit += nfit
	}

	return fit
}

func Lbase_capTant(tags string, smd bool) (float64, float64, float64, float64, float64, float64) {

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
