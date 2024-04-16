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

// FIDES 2022
var css []cs = []cs{
	{"U", "opto", 7, 2, 2},
	{"U", "", 10, 2, 1},

	{"Q", "gaas", 9, 3, 5},
	{"Q", "gan", 8, 3, 4},
	{"Q", "", 8, 2, 1},

	{"D", "gaas", 9, 3, 5},
	{"D", "gan", 8, 3, 4},
	{"D", "led", 7, 2, 3},
	{"D", "", 8, 2, 1},

	{"C", "type1", 7, 5, 2},
	{"C", "np0", 7, 5, 2},
	{"C", "c0g", 7, 5, 2},

	{"C", "type2", 7, 6, 1},
	{"C", "x5r", 7, 6, 1},
	{"C", "x7r", 7, 6, 1},
	{"C", "x7s", 7, 6, 1},

	{"C", "x5r flex", 7, 4, 2},
	{"C", "x7r flex", 7, 4, 1},

	{"C", "alu", 7, 7, 1},
	{"C", "tant", 8, 7, 1},
	{"C", "film", 7, 6, 1},
	{"C", "elco", 7, 7, 1},

	{"R", "melf", 4, 2, 4},
	{"R", "fuse", 6, 6, 4},
	{"R", "thick power", 2, 4, 1},
	{"R", "thick", 4, 3, 5},
	{"R", "ww power", 2, 4, 1},
	{"R", "ww", 2, 1, 3},
	{"R", "thin", 5, 5, 4},
	{"R", "network", 3, 5, 3},
	{"R", "", 5, 5, 4}, // Assume thin

	{"R", "potmeter", 1, 5, 2},
	{"R", "variable", 1, 5, 2},

	{"L", "trafo power", 6, 7, 4},
	{"L", "power", 7, 6, 3},
	{"L", "trafo", 6, 5, 3},
	{"L", "", 5, 4, 4},

	{"X", "oscillator", 7, 9, 3},
	{"X", "", 2, 10, 5},
	{"RL", "", 7, 10, 2},
	{"SW", "", 7, 10, 1},
	{"PCB", "", 4, 10, 8},
	{"J", "", 1, 10, 3},
}

// FIDES 2022
func Cs(class string, tags []string) float64 {

	class = strings.ToUpper(class)

	for _, cref := range css {

		if cref.class == class {
			if len(tags) == 0 && len(cref.tags) == 0 {
				return 0.725*cref.eos + 0.225*cref.mos + 0.05*cref.tos
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
				return 0.725*cref.eos + 0.225*cref.mos + 0.05*cref.tos
			}
		}
	}

	return math.NaN()
}

// contribution of induced factors (overstresses):
// Electrical overstress, mechanical overstress, thermal overstress

func PiInduced(comp *Component, phase *Phase) (float64, error) {

	cs := Cs(comp.Class, comp.Tags)

	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	return math.Pow(piPlacement(comp.Tags)*phase.AppFactor*PiRuggedising(), 0.511*math.Log(cs)), nil
}

func PiInducedPcb(phase *Phase) float64 {
	return math.Pow(phase.AppFactor*PiRuggedising(), 0.511*math.Log(Cs("PCB", nil)))
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

// Quality and technical control over manufacturing of the item
// (Use default value)
func PiPM() float64 {
	return 1.7
}

// quality and technical control over the development, manufacturing and
// usage process for the product containing the item
// (Use default value)
func PiProcess() float64 {
	return 4 // Not evaluated. Give a little room for risk (rolf).
}
