package fides

import (
	"errors"
	"math"
)

// capacitorType is a row of the capacitor tables of FIDES 2022 (pp. 151-157).
type capacitorType struct {
	Name          string
	l0, ea, sref  float64 // basic failure rate, activation energy (eV), reference electrical stress
	th, tcy, mech float64 // weights γ of the physical stresses
	eos, mos, tos float64 // relative sensitivities to overstress
}

// CapacitorFIT returns the FIDES 2022 FIT of a ceramic, aluminium or
// tantalum capacitor.
func CapacitorFIT(comp *Component, mission *Mission) (float64, error) {

	// Vmax and V are needed for capacitors

	if comp.Vmax == 0 || math.IsNaN(comp.Vmax) {
		return math.NaN(), errors.New("Vmax not set")
	}

	// V = 0 is valid: the thermo-electrical term of the FIDES formula is
	// then zero and the FIT comes from the remaining terms.
	if math.IsNaN(comp.V) {
		return math.NaN(), errors.New("working V not set")
	}

	if comp.V > comp.Vmax {
		return math.NaN(), errors.New("working V higher than limit Vmax")
	}

	ct, err := CapacitorType(comp)
	if err != nil {
		return math.NaN(), err
	}
	cs := csens(ct.eos, ct.mos, ct.tos)
	placement := piPlacement(comp.Tags)

	var factor float64
	for _, ph := range mission.Phases {

		// General rule
		if ph.Tamb > comp.Tmax {
			return math.NaN(), errors.New("Using component above its Tmax")
		}

		pi := ct.th*PiThermal_cap(ct.ea, ph.Tamb, ct.sref, comp.V/comp.Vmax, ph.On) +
			ct.tcy*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			ct.mech*PiMech(ph.Grms)

		pi *= ph.Duration / mission.Ttotal
		pi *= piInducedCs(cs, placement, ph)
		factor += pi
	}

	return ct.l0 * factor * PiPM() * PiProcess(), nil
}

// https://en.wikipedia.org/wiki/Ceramic_capacitor
//
// X = -55, Y = -30
// 5 = 85, 6 = 105, 7 = 125, 8 = 150
// R = 15%, S= 22%, U=+22/-56, V=+22/-82

var cerCapTags = []string{"x5r", "x5s", "x6r", "x6s", "x7r", "x7s", "x8r", "x8s", "np0", "c0g", "y5v"}

// Can be improved to return tolerance and temperature limits
func capType(tags []string) string {

	if contains(tags, "tant") || contains(tags, "tantalium") || contains(tags, "tantalum") {
		return "tant"
	}

	if contains(tags, "alu") || contains(tags, "elco") {
		return "alu"
	}

	for _, tag := range cerCapTags {
		if contains(tags, tag) || contains(tags, "type1") || contains(tags, "type2") {
			return "cer"
		}
	}

	return ""
}

// CapacitorType returns the FIDES 2022 capacitor type of a component, from
// its tags, value and rated voltage.
func CapacitorType(c *Component) (capacitorType, error) {

	switch capType(c.Tags) {
	case "cer":
		return ceramicType(c), nil
	case "alu":
		// Aluminium electrolytic (p. 154)
		if contains(c.Tags, "solid") || contains(c.Tags, "dry") || contains(c.Tags, "polymer") {
			return capacitorType{"solid aluminium electrolytic", 0.4, 0.40, 0.55, 0.85, 0.14, 0.01, 7, 7, 1}, nil
		}
		return capacitorType{"wet aluminium electrolytic", 0.21, 0.40, 0.5, 0.85, 0.14, 0.01, 7, 7, 1}, nil
	case "tant":
		return tantalumType(c), nil
	}
	return capacitorType{}, errors.New("unknown capacitor type: tag it with its dielectric (x7r, c0g, …), alu or tant")
}

// ceramicType is the ceramic capacitor type (pp. 151-152). Type I has a
// defined temperature coefficient (np0, c0g, type1); flex marks polymer
// terminations. The category follows from the CV product, and topend marks
// the largest capacitance of a range.
func ceramicType(c *Component) capacitorType {

	type1 := contains(c.Tags, "np0") || contains(c.Tags, "c0g") || contains(c.Tags, "type1")
	flex := contains(c.Tags, "flex")
	topend := contains(c.Tags, "topend")

	cvp := c.Value * c.Vmax
	cat := 3
	if type1 {
		if cvp <= 5e-8 {
			cat = 1
		} else if cvp <= 1e-6 || !topend {
			cat = 2
		}
	} else {
		if cvp <= 5e-6 {
			cat = 1
		} else if cvp <= 1e-4 || !topend {
			cat = 2
		}
	}

	switch {
	case type1:
		switch cat {
		case 1:
			return capacitorType{"ceramic type I, category 1", 0.03, 0.1, 0.3, 0.70, 0.28, 0.02, 7, 5, 2}
		case 2:
			return capacitorType{"ceramic type I, category 2", 0.05, 0.1, 0.3, 0.70, 0.28, 0.02, 7, 5, 2}
		}
		return capacitorType{"ceramic type I, category 3", 0.40, 0.1, 0.3, 0.69, 0.26, 0.05, 7, 5, 2}

	case flex:
		// Polymer terminations: X7R and X5R differ in their TOS sensitivity
		tos := 2.0
		if contains(c.Tags, "x7r") {
			tos = 1
		}
		if cat == 1 {
			return capacitorType{"ceramic type II, polymer terminations, category 1", 0.08, 0.1, 0.3, 0.70, 0.28, 0.02, 7, 4, tos}
		}
		return capacitorType{"ceramic type II, polymer terminations, category 2 or 3", 0.15, 0.1, 0.3, 0.70, 0.28, 0.02, 7, 4, tos}
	}

	switch cat {
	case 1:
		return capacitorType{"ceramic type II, category 1", 0.08, 0.1, 0.3, 0.70, 0.28, 0.02, 7, 6, 1}
	case 2:
		return capacitorType{"ceramic type II, category 2", 0.15, 0.1, 0.3, 0.70, 0.28, 0.02, 7, 6, 1}
	}
	return capacitorType{"ceramic type II, category 3", 1.20, 0.1, 0.3, 0.44, 0.51, 0.05, 7, 6, 1}
}

// tantalumType is the tantalum capacitor type (pp. 156-157). Wet (gel
// electrolyte) capacitors are silver case, glass-sealed unless tagged
// elastomer or tantalum_case; solid capacitors are SMD unless tagged axial,
// bead or tht.
func tantalumType(c *Component) capacitorType {

	if contains(c.Tags, "wet") {
		switch {
		case contains(c.Tags, "elastomer"):
			return capacitorType{"wet tantalum, silver case, elastomer-sealed", 0.77, 0.15, 0.6, 0.87, 0.01, 0.12, 8, 7, 1}
		case contains(c.Tags, "tantalum_case") || contains(c.Tags, "glass_sealed") && !contains(c.Tags, "silver_case"):
			return capacitorType{"wet tantalum, tantalum case, glass-sealed", 0.05, 0.15, 0.6, 0.88, 0.04, 0.08, 8, 7, 1}
		}
		return capacitorType{"wet tantalum, silver case, glass-sealed", 0.33, 0.15, 0.6, 0.81, 0.01, 0.18, 8, 7, 1}
	}

	switch {
	case contains(c.Tags, "axial"):
		return capacitorType{"solid tantalum, axial metal packaging", 0.25, 0.15, 0.4, 0.94, 0.04, 0.02, 8, 7, 1}
	case contains(c.Tags, "bead") || contains(c.Tags, "tht"):
		return capacitorType{"solid tantalum, bead packaging", 1.09, 0.15, 0.4, 0.86, 0.12, 0.02, 8, 7, 1}
	}
	return capacitorType{"solid tantalum, SMD packaging", 0.54, 0.15, 0.4, 0.84, 0.14, 0.02, 8, 7, 1}
}
