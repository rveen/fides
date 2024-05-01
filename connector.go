package fides

import (
	"errors"
	"fmt"
	"math"
)

// PCB connectors, less that one insertion/year
func ConnectorFIT(comp *Component, mission *Mission) (float64, error) {

	if comp.Np < 1 {
		return math.NaN(), errors.New("Connector with 0 contacts")
	}

	piMounting := 10.0
	if contains(comp.Tags, "pressfit") {
		piMounting = 1
	} else if contains(comp.Tags, "tht") {
		piMounting = 6
	}

	// Base FIT
	fit := 0.1 * piMounting * math.Pow(float64(comp.Np), 0.5) * 0.2
	var factor float64

	for _, ph := range mission.Phases {

		// General rule
		if ph.Tamb+comp.T > comp.Tmax {
			s := fmt.Sprintf("Using component above its Tmax %f (Tamb=%f, Td=%f)\n", comp.Tmax, ph.Tamb, comp.T)
			return math.NaN(), errors.New(s)
		}

		// Thermal (0 if off)
		pi := 0.58 * PiThermal(0.1, ph.Tamb+comp.T, ph.On)

		// Thermal cycling
		pi += 0.04 * PiTCSolder(ph.NCycles, ph.Duration, ph.CycleDuration, ph.Tdelta, ph.Tmax)

		// Mechanical
		pi += 0.05 * PiMech(ph.Grms)

		// Humidity
		pi += 0.13 * PiRH(0.8, ph.RH, ph.Tamb)

		// Chemical
		pi += PiChemical(0.2, ph.SalinePollution, ph.AmbientPollution, ph.ZonePollution, ph.IP)

		// Proportion of time in this phase
		pi *= ph.Duration / mission.Ttotal

		// Stress factors and sensibility
		ifactor, err := PiInduced(comp, ph)
		if err != nil {
			return math.NaN(), err
		}
		pi *= ifactor

		factor += pi

	}

	return fit * factor * PiPM() * PiProcess(), nil
}
