package fides

import (
	"errors"
	"math"
)

// OptoFIT returns the FIDES 2022 FIT of an optocoupler (class U, tag opto or
// optocoupler; pp. 144-145): a phototransistor optocoupler unless tagged
// photodiode. The package basic failure rates are those of the integrated
// circuit or discrete semiconductor package; comp.T is the junction
// temperature rise over ambient, P·Rth.
func OptoFIT(comp *Component, mission *Mission) (float64, error) {

	lth, ltcChip, lmChip := 0.11, 0.021, 0.011
	if contains(comp.Tags, "photodiode") {
		lth, ltcChip, lmChip = 0.05, 0.01, 0.005
	}
	if comp.N > 1 {
		n := math.Sqrt(float64(comp.N))
		lth, ltcChip, lmChip = lth*n, ltcChip*n, lmChip*n
	}

	p := NewPackage(comp.Package)
	lrh, ltc, lts, lm := p.FitBase()
	if lrh < 0 || math.IsNaN(lrh) {
		return math.NaN(), errors.New("Missing data for lpkg(rh,tc...) calculation for package: [" + p.Name + "]")
	}

	dt := comp.T
	if math.IsNaN(dt) {
		dt = 0
	}

	var factor float64
	for _, ph := range mission.Phases {

		tj := ph.Tamb + dt

		pi := lth*PiThermal(0.4, tj, ph.On) +
			ltc*PiTCCase(ph.NCycles, ph.Duration, ph.Tdelta, ph.Tmax) +
			(lts+ltcChip)*PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax) +
			lrh*PiRH2(0.9, ph.RH, ph.Tamb, ph.On) +
			(lm+lmChip)*PiMech(ph.Grms)

		pi *= ph.Duration / mission.Ttotal

		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		pi *= ifactor

		factor += pi
	}

	return factor * PiPMActive() * PiProcess(), nil
}
