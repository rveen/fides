package fides

import (
	"log"
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

	{"C", "type2", 7, 6, 1},
	{"C", "x5r", 7, 6, 1},
	{"C", "x7r", 7, 6, 1},
	{"C", "x7s", 7, 6, 1},

	{"C", "x5r flex", 7, 4, 2},
	{"C", "x7r flex", 7, 4, 1},

	{"C", "alu", 7, 7, 1},
	{"C", "tant", 8, 7, 1},
	{"C", "film", 7, 6, 1},

	{"R", "melf", 4, 2, 4},
	{"R", "fuse", 6, 6, 4},
	{"R", "thick power", 2, 4, 1},
	{"R", "thick", 4, 3, 5},
	{"R", "ww power", 2, 4, 1},
	{"R", "ww", 2, 1, 3},
	{"R", "thin", 5, 5, 4},
	{"R", "network", 3, 5, 3},

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
func Cs(class, tags string) float64 {

	ctags := strings.Fields(tags)

	for _, c := range css {

		if c.class == class {
			if tags == "" && c.tags == "" {
				return 0.725*c.eos + 0.225*c.mos + 0.05*c.tos
			}

			// All tags present in c.tags must be present in ctags argument

			ttags := strings.Fields(c.tags)

			for _, tag := range ttags {
				if !contains(ctags, tag) {
					continue
				}
			}
			return 0.725*c.eos + 0.225*c.mos + 0.05*c.tos
		}
	}

	log.Printf("[!] Cs(%s,%s) no result\n", class, tags)

	return -1
}

// FIDES 2009
func CSensibility(class, typ string) float64 {

	if class == "R" {
		switch typ {
		case "melf":
			return 3.85
		case "power_film":
			return 2.25
		case "ww_precision":
			return 1.75
		case "ww_power":
			return 2.25
		case "pot_cermet":
			return 2.5
		case "chip":
			return 4.75
		case "smd_network":
			return 4.25
		case "metal_foil_precision":
			return 5.8
		default:
			log.Println("unknown resistor type", typ, ". Returning default Csens")
			return 4.75
		}
	}

	log.Fatalln("unknown class and/or type", class, typ)
	return 0
}
