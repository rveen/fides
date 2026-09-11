package fides

import (
	"errors"
	"math"
)

// PcbFIT returns the FIDES 2022 FIT of a printed circuit board (class PCB;
// pp. 174-176), from comp.Layers, comp.Mounts (mounting points: SMD pads and
// through-holes), comp.PiClass (1 to 6, from the minimum conductor width and
// spacing) and comp.PiTechno (0.25 through-holes, 0.5 blind holes, 1
// micro-vias, 2.5 pad on via).
func PcbFIT(comp *Component, mission *Mission) (float64, error) {

	if comp.Layers < 1 || comp.Mounts < 1 {
		return math.NaN(), errors.New("PCB: set the number of layers and of mounting points")
	}
	if comp.PiClass <= 0 || comp.PiTechno <= 0 {
		return math.NaN(), errors.New("PCB: set ΠClass (1 to 6) and ΠTechno (0.25, 0.5, 1 or 2.5)")
	}

	l0 := Lbase_Pcb(comp.Layers, comp.Mounts, comp.PiClass, comp.PiTechno)
	cs := Cs("PCB", nil)

	var fit float64
	for _, ph := range mission.Phases {

		prot := 0.0
		if !ph.IP {
			prot = 1
		}
		tv := PiTV(ph.Tamb)

		pi := 0.6*tv*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			0.2*tv*PiMech(ph.Grms) +
			0.18*tv*PiRH(0.8, ph.RH, ph.Tamb) +
			0.02*tv*ph.SalinePollution*ph.AmbientPollution*ph.ZonePollution*prot

		// The placement factor of PCBs is 1 (p. 174)
		fit += l0 * ph.Duration / mission.Ttotal * pi * piInducedCs(cs, 1, ph)
	}

	return fit * PiPM() * PiProcess(), nil
}

// Lbase_Pcb is λ0 of a PCB (p. 175).
func Lbase_Pcb(nLayers, nMounts int, class, techno float64) float64 {
	return 5e-4 * math.Sqrt(float64(nLayers)) * float64(nMounts) / 2 * class * techno
}

// PiTV is the high temperature factor of a PCB (p. 176).
func PiTV(tamb float64) float64 {
	if tamb < 110 {
		return 1
	}
	return math.Exp(0.2 * (tamb - 110))
}
