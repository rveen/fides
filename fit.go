package fides

import (
	"math"
	"strings"
)

func FIT(comp *Component, mission *Mission) float64 {

	switch strings.ToUpper(comp.Class) {

	case "U":

		typ := strings.ToLower(comp.Type)
		if strings.HasPrefix(typ, "opto") {
			return OptoFIT(comp, mission)
		}
		return IcFIT(comp, mission)
	case "ASIC":
		return math.NaN()
	case "Q":
		return SemiconductorFIT(comp, mission)
	case "D":
		return SemiconductorFIT(comp, mission)
	case "LED":
		return math.NaN()
	case "OPTOCOUPLER", "OPTO":
		return OptoFIT(comp, mission)
	case "R":
		return ResistorFIT(comp, mission)
	case "FUSE":
		return math.NaN()
	case "C":
		return CapacitorFIT(comp, mission)
	case "L":
		return InductorFIT(comp, mission)
	case "PIEZO":
		return math.NaN()
	case "RELAY":
		return math.NaN()
	case "SW":
		return math.NaN()
	case "J":
		return ConnectorFIT(comp, mission)
	case "PCB":
		return math.NaN()

	}
	return math.NaN()
}
