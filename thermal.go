package fides

import "math"

// Arrhenius law, 20ºC
func PiThermal_resistor(ea, temp float64) float64 {
	return math.Exp(11604 * ea * (1/293 - 1/(temp+273)))
}

func PiThermal_connector(on bool, temp float64) float64 {
	if !on {
		return 0
	}
	ea := 0.1
	return math.Exp(11604 * ea * (1/293 - 1/(temp+273)))
}

func PiThermal_semiconductor(ratio, temp float64) float64 {
	ea := 0.7
	return math.Max(0.056, math.Pow(ratio, 2.4)*math.Exp(11604*ea*(1/293-1/(temp+273))))
}

func PiThermal_cap(ea, tamb, sref, ratio float64) float64 {
	return math.Pow(1/sref*ratio, 3) * math.Exp(11604*ea*(1/293-1/(tamb+273)))
}

// PiThermal for ICs is 0 in non-operating mode
func PiThermal_ic(temp float64, on bool) float64 {
	if !on {
		return 0
	}
	return math.Exp(11604.0 * 0.7 * (1.0/293.0 - 1.0/(temp+273)))
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

func PiRHea(rh, temp, ea float64) float64 {
	return math.Pow(rh/70, 4.4) * math.Exp(11604*ea*(1/293-1/(temp+273)))
}

// Same as PiRH, but return 0 in off mode
func PiRH2(rh, temp float64, on bool) float64 {
	if on {
		return 0
	}
	return math.Pow(rh/70, 4.4) * math.Exp(11604*0.9*(1/293-1/(temp+273)))
}

// Arrhenius25 (in ºC, reference 25ºC)
func Arrhenius25(ea, temp float64) float64 {
	return math.Exp(11604.967 * ea * (1/293 - 1/(temp+273)))
}

// Arrhenius (in K)
func ArrheniusK(ea, t0, t1 float64) float64 {
	return math.Exp(11604.967 * ea * (1/t0 - 1/t1))
}

// Arrhenius law
func Arrhenius(ea, t1, t0 float64) float64 {
	return math.Exp(11604.967 * ea * (1/(t0+273) - 1/(t1+273)))
}

// Norris-Landberg, general form
//
// For SAC305 lead-free solder: a=2.3, b=0.3, c=4562
// See "Norris–Landzberg Acceleration Factors and Goldmann Constants for SAC305 Lead-Free Electronics"
// (Journal of Electronic Packaging · September 2012)
//
func NorrisLandzberg(tdeltaRef, tdeltaUse, tmaxRef, tmaxUse, fRef, fUse float64, a, b, c float64) float64 {
	return math.Pow(tdeltaRef/tdeltaUse, a) * math.Pow(fUse/fRef, b) * math.Exp(c*(1/(tmaxUse+273)-1/(tmaxRef+273)))
}

// Temperature cycling, case, Norris-Landzberg model (semiconductor cases)
func PiTCCase(nc int, time, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(tdelta/20, 4) * math.Exp(1414*(1/313-1/(tmax+273)))
}

// Temperature cycling,solder joints, Norris-Landzberg model
func PiTCSolder(nc int, time, duration, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(math.Min(duration, 2)/2, 1.3) * math.Pow(tdelta/20, 2.65 /*1.9*/) * math.Exp(1414*(1/313-1/(tmax+273)))
}
