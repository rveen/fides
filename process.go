package fides

import (
	"math"
)

// Mess
func PiApplication(on bool) float64 {
	return 1 // Ignore
}

// PiRuggedising represents the influence of the policy for taking account of
// overstresses in the product development.
//
// Return max value (very controlled process)
func PiRuggedising() float64 {
	return 1 // Mostly ignore
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
func PiInduced(on, analog, itf, power bool, csensibility float64) float64 {
	return math.Pow(PiPlacement(analog, itf, power)*PiApplication(on)*PiRuggedising(), 0.511*math.Log(csensibility))
}

// PiPlacement represents the influence of the item placement in the system
// (particularly whether or not it is interfaced).
func PiPlacement(analog bool, itf bool, power bool) float64 {

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
