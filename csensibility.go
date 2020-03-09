package fides

import (
	"log"
)

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
			log.Fatalln("unknown resistor type", typ)
		}
	}

	if class == "Q" || class == "D" {
		switch typ {
		case "discrete":
			return 5.2
		}
	}

	log.Fatalln("unknown class and/or type", class, typ)
	return 0
}
