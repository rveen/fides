package fides

import (
	"fmt"
	"strconv"
)

type Phase struct {
	Name          string
	Duration      float64
	NCycles       int
	CycleDuration float64
	On            bool
	Tamb          float64
	Tdelta        float64
	Tmax          float64
	RH            float64
	Grms          float64

	// 1 = weak,low; 2 = high,strong
	SalinePollution float64

	// 1 = weak,low; 1.5 = moderate; 2 = strong,high
	AmbientPollution float64

	// 1 = weak,low; 2 = moderate, 4 = strong,high
	ZonePollution float64

	// Ingress protection (true = hermetic, sealed)
	IP   bool
	Tags string

	// Application factor
	AppFactor float64
}

type Mission struct {
	Phases []*Phase
}

func NewMission() *Mission {
	return &Mission{}
}

func (mission *Mission) AddPhase(ph *Phase) {
	mission.Phases = append(mission.Phases, ph)
}

func (mission *Mission) FromCsv(file string) error {

	m, err := csvRead(file)
	if err != nil {
		return err
	}

	for i := 0; i < len(m); i++ {
		ph := &Phase{}
		p := m[i]

		ph.Name = p["phase"]
		ph.Duration, _ = strconv.ParseFloat(p["duration"], 64)
		ph.On = (p["on"] == "on" || p["on"] == "true")
		ph.Tamb, _ = strconv.ParseFloat(p["tamb"], 64)
		ph.Tdelta, _ = strconv.ParseFloat("tdelta", 64)
		ph.NCycles, _ = strconv.Atoi(p["ncycles"])
		ph.CycleDuration, _ = strconv.ParseFloat(p["tcycle"], 64)
		ph.Tmax, _ = strconv.ParseFloat(p["tmax"], 64)
		ph.RH, _ = strconv.ParseFloat(p["rh"], 64)
		ph.Grms, _ = strconv.ParseFloat(p["grms"], 64)
		ph.SalinePollution = level(2, p["saline_pollution"])
		ph.AmbientPollution = level(2, p["env_pollution"])
		ph.ZonePollution = level(4, p["app_pollution"])
		ph.IP = (p["ip"] == "sealed" || p["ip"] == "hermetic")
		ph.AppFactor, _ = strconv.ParseFloat(p["pi_app"], 64)

		mission.AddPhase(ph)
	}
	return nil
}

func level(max float64, s string) float64 {

	switch s {
	case "weak", "low":
		return 1
	case "moderate":
		return max / 2
	default: // strong, high
		return max
	}
}

func (m *Mission) ToCsv() string {

	s := "phase, duration, on, tamb, tdelta, ncycles, tcycle, rh, grm, tmax, saline, env, app, ip, factor\n"

	for _, ph := range m.Phases {
		s += fmt.Sprintf("%s, ", ph.Name)
		s += fmt.Sprintf("%.1f, ", ph.Duration)
		s += fmt.Sprintf("%t, %.1f, ", ph.On, ph.Tamb)
		s += fmt.Sprintf("%.0f, ", ph.RH)
		s += fmt.Sprintf("%.1f, ", ph.Grms)
		s += fmt.Sprintf("%.1f, ", ph.Tdelta)
		s += fmt.Sprintf("%d, ", ph.NCycles)
		s += fmt.Sprintf("%.2f, ", ph.CycleDuration)
		s += fmt.Sprintf("%.1f, ", ph.Tmax)
		s += fmt.Sprintf("%.0f, ", ph.SalinePollution)
		s += fmt.Sprintf("%.0f, ", ph.AmbientPollution)
		s += fmt.Sprintf("%.0f, ", ph.ZonePollution)
		s += fmt.Sprintf("%t, ", ph.IP)
		s += fmt.Sprintf("%.1f\n", ph.AppFactor)
	}
	return s
}
