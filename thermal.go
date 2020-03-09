package fides

import "math"

// [1] ratio is 1 for component other than signal diodes
func PiThermal_resistor(tamb, tratio float64) float64 {
	return math.Exp(11604 * 0.15 * (1/293 - 1/(tamb+273+tratio)))
}

// [1] ratio is 1 for component other than signal diodes
func PiThermal_semiconductor(ratio, tj float64) float64 {
	return math.Max(0.056, math.Pow(ratio, 2.4)) * math.Exp(11604*0.7*(1/293-1/(tj+273)))
}

func PiMech(grms float64) float64 {
	return math.Pow(grms*2, 1.5)
}

func PiRH(rh, temp float64) float64 {
	return math.Pow(rh/70, 4.4) * math.Exp(11604*0.9*(1/293-1/(temp+273)))
}

func PiCase(nc int, time, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(tdelta/20, 4) * math.Exp(1414*(1/313-1/(tmax+273)))
}

func PiSolder(nc int, time, tdelta, tmax, phi float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(math.Min(phi, 2)/2, 1.3) * math.Pow(tdelta/20, 1.9) * math.Exp(1414*(1/313-1/(tmax+273)))
}
