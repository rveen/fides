package fides

import (
	"errors"
	"math"
	"strings"
)

type cs struct {
	class, tags   string
	eos, mos, tos float64
}

// Relative sensitivities to overstress of the families whose models do not
// choose their own (FIDES 2022). Resistors, capacitors, inductors and
// crystals take theirs from their type tables.
var css []cs = []cs{
	{"U", "opto", 7, 2, 2},
	{"U", "optocoupler", 7, 2, 2},
	{"U", "", 10, 2, 1},

	{"Q", "", 8, 2, 1},

	{"D", "led", 7, 2, 3},
	{"D", "", 8, 2, 1},

	{"PCB", "", 4, 10, 8},
	{"J", "", 1, 10, 3},
}

// csens is Csensitivity from the relative sensitivities to electrical,
// mechanical and thermal overstress (FIDES 2022, p. 110).
func csens(eos, mos, tos float64) float64 {
	return 0.725*eos + 0.225*mos + 0.05*tos
}

// Cs returns Csensitivity of a class and tags, from the first row of css
// whose tags the component has all of.
func Cs(class string, tags []string) float64 {

	class = strings.ToUpper(class)

	for _, cref := range css {

		if cref.class == class {
			if len(tags) == 0 && len(cref.tags) == 0 {
				return csens(cref.eos, cref.mos, cref.tos)
			}

			// All tags present in cref.tags must be present in the tags argument

			ctags := strings.Fields(cref.tags)

			n := 0
			for _, tag := range ctags {
				if contains(tags, tag) {
					n++
				}
			}
			if n == len(ctags) {
				return csens(cref.eos, cref.mos, cref.tos)
			}
		}
	}

	return math.NaN()
}

// Contribution of induced factors (overstresses):
// Electrical overstress, mechanical overstress, thermal overstress
func PiInduced(comp *Component, phase *Phase) (float64, error) {

	cs := Cs(comp.Class, comp.Tags)

	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	return piInducedCs(cs, piPlacement(comp.Tags), phase), nil
}

// piInducedCs is ΠInduced for a given Csensitivity and placement factor
// (FIDES 2022, p. 110).
func piInducedCs(cs, placement float64, phase *Phase) float64 {
	return math.Pow(placement*phase.AppFactor*PiRuggedising(), 0.511*math.Log(cs))
}

func PiInducedPcb(phase *Phase) float64 {
	return piInducedCs(Cs("PCB", nil), 1, phase)
}

// PiPlacement represents the influence of the item placement in the system
// (particularly whether or not it is interfaced).
func piPlacement(tags []string) float64 {

	analog := contains(tags, "analog")
	power := contains(tags, "power")
	itf := contains(tags, "interface")

	if !analog {
		if itf {
			return 1.6
		}
		return 1
	}

	if !power {
		if itf {
			return 2
		}
		return 1.3
	}

	if !itf {
		return 1.6
	}
	return 2.5
}

// Process factors are mostly ignored

// The lead-free process part (PiLF) is ignored as a mature process is assumed.
// (It needs to be taken into account in the thermal cycle model, nonetheless)

// PiRuggedising represents the influence of the policy for taking account of
// overstresses in the product development.
// (Use default value)
func PiRuggedising() float64 {
	return 1.7
}

// PiPM is the quality and technical control over manufacturing of a passive
// item (default value, FIDES 2022, p. 36).
func PiPM() float64 {
	return 1.6
}

// PiPMActive is PiPM for active components: integrated circuits, discrete
// semiconductors, LEDs and optocouplers (default value, FIDES 2022, p. 36).
func PiPMActive() float64 {
	return 1.7
}

// PiPW is the design factor of discrete power semiconductors, silicon MOS
// over 5 W and IGBTs (default value, FIDES 2022, p. 138).
func PiPW() float64 {
	return 10
}

// quality and technical control over the development, manufacturing and
// usage process for the product containing the item
// (Use default value)
func PiProcess() float64 {
	return 4 // Not evaluated. Give a little room for risk (rolf).
}
