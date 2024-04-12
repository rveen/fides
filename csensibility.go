package fides

import (
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

	{"Q", "", 8, 2, 1},

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
				if containsTag(tags, tag) {
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
