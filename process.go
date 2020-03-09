package fides

import (
	"math"
)

// Mess
func PiApplication(on bool) float64 {
	if on {
		return 5.1
	}
	return 3.1
}

// Return max value (very controlled process)
func PiRuggedising() float64 {
	return 2
}

func PiPM() float64 {
	return 0.5
}

func PiProcess() float64 {
	return 2
}

func PiInduced(on, analog, itf, power bool, csensibility float64) float64 {
	return math.Pow(PiPlacement(analog, itf, power)*PiApplication(on)*PiRuggedising(), 0.511*math.Log(csensibility))
}

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
