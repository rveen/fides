package fides

import "math"

// Acceleration factors

// InvBoltzman is 1/k_B in K/eV, rounded as in the formulas of the FIDES 2022
// guide (p. 41).
const InvBoltzman float64 = 11604

// PiThermal_voltageFactor is ΠEl of a signal diode up to 1 A (not zener or
// tvs): (V/Vrated)^2.4 above a ratio of 0.3, and 0.056 up to it (FIDES 2022,
// p. 137).
func PiThermal_voltageFactor(v, vmax float64) float64 {
	ratio := v / vmax
	if ratio <= 0.3 {
		return 0.056
	}
	return math.Pow(ratio, 2.4)
}

func PiThermal_cap(ea, tamb, sref, ratio float64, on bool) float64 {
	if !on {
		return 0
	}
	return math.Pow(1/sref*ratio, 3) * Arrhenius25(ea, tamb)
}

// PiThermal for ICs is 0 in non-operating mode
func PiThermal(ea, temp float64, on bool) float64 {
	if !on {
		return 0
	}
	return Arrhenius25(ea, temp)
}

// -----------------------------------------------------------------------------

func PiChemical(ea, sal, env, zone float64, sealed bool) float64 {
	if sealed {
		return 0
	}

	return ea * sal * env * zone
}

// Basquin's law
func PiMech(grms float64) float64 {
	return math.Pow(grms*2, 1.5)
}

// Peck’s model
func PiRH(ea, rh, temp float64) float64 {
	return math.Pow(rh/70, 4.4) * Arrhenius25(ea, temp)
}

// Same as PiRH, but return 0 in on mode
func PiRH2(ea, rh, temp float64, on bool) float64 {
	if on {
		return 0
	}
	return math.Pow(rh/70, 4.4) * Arrhenius25(ea, temp)
}

// Arrhenius25 returns the Arrhenius factor at temp [°C] relative to 293 K (20 °C).
// FIDES 2022 uses 293 K as the reference (Kelvin offset: 273, not 273.15).
// This differs intentionally from model.Arrhenius25 (298.15 K / 25 °C).
func Arrhenius25(ea, temp float64) float64 {
	// 1.0/293: the untyped constant 1/293 is an integer division (0)
	return math.Exp(InvBoltzman * ea * (1.0/293 - 1/(temp+273)))
}

// Arrhenius (in K)
func ArrheniusK(ea, t0, t1 float64) float64 {
	return math.Exp(InvBoltzman * ea * (1/t0 - 1/t1))
}

// Arrhenius law
func Arrhenius(ea, t1, t0 float64) float64 {
	return math.Exp(InvBoltzman * ea * (1/(t0+273) - 1/(t1+273)))
}

// Norris-Landberg, general form
//
// For SAC305 lead-free solder: a=2.3, b=0.3, c=4562
// See "Norris–Landzberg Acceleration Factors and Goldmann Constants for SAC305 Lead-Free Electronics"
// (Journal of Electronic Packaging · September 2012)
// See also: https://www.lamar.edu/engineering/_files/documents/mechanical/dr.-fan-publications/2008/Fan%202008_13%20ECTC_3.pdf
func NorrisLandzberg(tdeltaRef, tdeltaUse, tmaxRef, tmaxUse, fRef, fUse float64, a, b, c float64) float64 {
	return math.Pow(tdeltaRef/tdeltaUse, a) * math.Pow(fUse/fRef, b) * math.Exp(c*(1/(tmaxUse+273)-1/(tmaxRef+273)))
}

// Temperature cycling, case, Norris-Landzberg model (semiconductor cases)
func PiTCCase(nc int, time, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(tdelta/20, 4) * math.Exp(1414*(1.0/313-1/(tmax+273)))
}

// Temperature cycling, solder joints, Norris-Landzberg model (FIDES 2022,
// pp. 43-44): the cycle duration factor is (min(θcy, 2)/2)^(1/3).
// See https://www.lamar.edu/engineering/_files/documents/mechanical/dr.-fan-publications/2008/Fan%202008_13%20ECTC_3.pdf
// (The 1.9 factor is OK for lead-free also, according to this paper)
func PiTCSolder(nc int, time, duration, tdelta, tmax float64) float64 {
	return 12 * float64(nc) / float64(time) * math.Pow(math.Min(duration, 2)/2, 1.0/3) * math.Pow(tdelta/20, 1.9) * math.Exp(1414*(1.0/313-1/(tmax+273)))
}
