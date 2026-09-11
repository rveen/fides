package fides

import (
	"errors"
	"math"
	"strings"
)

// FIT returns the FIDES 2022 failure rate of a component for a mission, in
// failures in 10⁹ hours.
func FIT(comp *Component, mission *Mission) (float64, error) {

	class := strings.ToUpper(comp.Class)

	if class == "PCB" {
		return PcbFIT(comp, mission)
	}

	// Tmax is mandatory for components
	if comp.Tmax == 0 || math.IsNaN(comp.Tmax) {
		return math.NaN(), errors.New("Tmax (max temperature of component) not set")
	}

	switch class {

	case "U":
		if contains(comp.Tags, "opto") || contains(comp.Tags, "optocoupler") {
			return OptoFIT(comp, mission)
		}
		fallthrough
	case "Q":
		return SemiconductorFIT(comp, mission)
	case "D":
		if contains(comp.Tags, "led") {
			return LEDFIT(comp, mission)
		}
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
