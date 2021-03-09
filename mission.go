package fides

type Phase struct {
	Name          string
	NCycles       int
	CycleDuration float64

	// hours in a year in this phase
	Time   float64
	On     bool
	Tamb   float64
	Tdelta float64
	Tmax   float64
	RH     float64
	Grms   float64

	// 1 = weak, 3 = strong
	SalinePollution float64

	// 1 = weak, 2 = moderate, 3 = strong
	AmbientPollution float64

	// 1 = weak, 2 = moderate, 3 = strong
	ApplicationPollution float64

	// Ingress protection (false = not hermetic, true = hermetic)
	IP bool
}

type Mission struct {
	Phases []Phase
}

func NewMission() *Mission {
	return &Mission{}
}

func (mission *Mission) AddPhase(ph Phase) {
	mission.Phases = append(mission.Phases, ph)
}
