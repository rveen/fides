package fides

import (
	"errors"
	"math"
	"strings"
)

func FIT(comp *Component, mission *Mission) (float64, error) {

	// Mandatory attribues:
	// - Tmax
	if comp.Tmax == 0 || math.IsNaN(comp.Tmax) {
		return math.NaN(), errors.New("Tmax (max temperature of component) not set")
	}

	class := strings.ToUpper(comp.Class)

	switch class {

	case "U":
		if contains(comp.Tags, "opto") {
			return OptoFIT(comp, mission)
		}
		fallthrough
	case "Q", "D":
		return SemiconductorFIT(comp, mission)
	case "R":
		return ResistorFIT(comp, mission)
	case "C":
		return CapacitorFIT(comp, mission)
	case "L":
		return InductorFIT(comp, mission)
	case "J":
		return ConnectorFIT(comp, mission)
	case "X":
		return PiezoFIT(comp, mission)
	default:
		return math.NaN(), errors.New("unsupported component type " + class)

	}
}
