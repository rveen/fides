package fides

import (
	"errors"
	"math"
)

func PcbFIT(mission *Mission, nLayers, nConn int) (float64, error) {

	var fit, nfit float64

	l0 := Lbase_Pcb(nLayers, nConn, 2, 0.25)
	cs := Cs("PCB", nil)
	if math.IsNaN(cs) {
		return math.NaN(), errors.New("Missing data for stress sensibility calculation")
	}

	for _, ph := range mission.Phases {

		prot := 0.0
		if !ph.IP {
			prot = 1
		}

		nfit = l0 * ph.Duration / 8760.0 *
			(0.6*PiTV(ph.Tamb)*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
				0.18*PiTV(ph.Tamb)*PiRH(0.9, ph.RH, ph.Tamb) +
				0.02*PiTV(ph.Tamb)*ph.SalinePollution*ph.AmbientPollution*ph.ApplicationPollution*prot +
				0.02*PiTV(ph.Tamb)*PiMech(ph.Grms))

		// Set PiApplication to 1 (by setting 2 falses here)
		nfit *= PiInduced(ph.On, nil, cs)

		fit += nfit
	}

	return fit, nil
}

func Lbase_Pcb(nLayers, nConn, class int, tech float64) float64 {
	return 0.0005 * math.Sqrt(float64(nLayers)) * float64(nConn) / 2.0 * float64(class) * tech
}

func PiTV(tamb float64) float64 {
	if tamb < 110 {
		return 1
	}
	return math.Exp(0.2 * (tamb - 110))
}
