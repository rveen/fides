package fides

import (
	"math"
)

// Process factors are mostly ignored, since they are subjetive.

// The lead-free process part (PiLF) is ignored as a mature process is assumed.
// (It needs to be taken into account in the thermal cycle model, nonetheless)

// Mess
func PiApplication(on bool) float64 {
	return 1 // Ignore
}

// PiRuggedising represents the influence of the policy for taking account of
// overstresses in the product development.
//
// Return default
func PiRuggedising() float64 {
	return 1.7
}

// Mess: quality and technical control over manufacturing of the item
/*
func PiPM() float64 {
	return 1 // Ignore
}
*/

// quality and technical control over the development, manufacturing and
// usage process for the product containing the item
func PiProcess() float64 {
	return 4 // Not evaluated. Give a little room for risk (rolf).
}

// contribution of induced factors (overstresses):
// Electrical overstress, mechanical overstress, thermal overstress
func PiInduced(on bool, tags []string, csensibility float64) float64 {
	return math.Pow(PiPlacement(tags)*PiApplication(on)*PiRuggedising(), 0.511*math.Log(csensibility))
}

// PiPlacement represents the influence of the item placement in the system
// (particularly whether or not it is interfaced).
func PiPlacement(tags []string) float64 {

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
