package fides

import "math"

// Arrhenius law
func PiThermal_resistor(ea, temp float64) float64 {
	return math.Exp(11604 * ea * (1/293 - 1/(temp+273)))
}

func PiThermal_semiconductor(ratio, temp float64) float64 {
	ea := 0.7
	return math.Max(0.056, math.Pow(ratio, 2.4)*math.Exp(11604*ea*(1/293-1/(temp+273))))
}

func PiThermal_cap(ea, tamb, sref, ratio float64) float64 {
	return math.Pow(1/sref*ratio, 3) * math.Exp(11604*ea*(1/293-1/(tamb+273)))
}

// -----------------------------------------------------------------------------

// Basquin's law
func PiMech(grms float64) float64 {
	return math.Pow(grms*2, 1.5)
}

// Peck’s model
func PiRH(rh, temp float64) float64 {
	return math.Pow(rh/70, 4.4) * math.Exp(11604*0.9*(1/293-1/(temp+273)))
}

// Temperature cycling, Norris-Landzberg model (semiconductor cases)
func PiTCCase(nc int, time, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(tdelta/20, 4) * math.Exp(1414*(1/313-1/(tmax+273)))
}

// Temperature cycling, Norris-Landzberg model
func PiThermalCycling(nc int, time, phi, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(math.Min(phi, 2)/2, 1.3) * math.Pow(tdelta/20, 1.9) * math.Exp(1414*(1/313-1/(tmax+273)))
}
