package fides

import "math"

func PcbFIT(mission *Mission, nLayers, nConn int) float64 {

	var fit, nfit float64

	l0 := Lbase_Pcb(nLayers, nConn, 2, 0.25)

	for _, ph := range mission.Phases {

		prot := 0.0
		if !ph.IP {
			prot = 1
		}

		nfit = l0 * ph.Time / 8760.0 *
			(0.6*PiTV(ph.Tamb)*PiTCSolder(ph.NCycles, ph.Time, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
				0.18*PiTV(ph.Tamb)*PiRH(ph.RH, ph.Tamb) +
				0.02*PiTV(ph.Tamb)*ph.SalinePollution*ph.AmbientPollution*ph.ApplicationPollution*prot +
				0.02*PiTV(ph.Tamb)*PiMech(ph.Grms))

		// Set PiApplication to 1 (by setting 2 falses here)
		nfit *= PiInduced(ph.On, false, false, false, 6.5)

		fit += nfit
	}

	return fit * PiPM() * PiProcess()
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
