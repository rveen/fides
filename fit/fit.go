package main

import (
	"flag"
	"fmt"
	"goapp/fides"
)

func main() {

	flag.Parse()

	mission := fides.NewMission()
	mission.AddPhase(fides.Phase{"parking", 365, 22.5, 8212.5, false, 23, 20, 43, 70, 0})
	mission.AddPhase(fides.Phase{"city", 730, 0.25, 182.5, true, 50, 30, 80, 50, 1.6})
	mission.AddPhase(fides.Phase{"road", 730, 0.5, 365, true, 30, 30, 60, 50, 1.4})

	comp := &fides.Component{"diode_signal", "SMD, signal, llead, plastic", 1, 400, true, false, true, 0, 0.001, 0, 0}

	// Ciclos
	// carretera: 365 h, 35º, nc = 730, phi = 0.5, PiApp = 5.1
	// ciudad congestionada: 182.5 h, 50º, nc = 730, phi = 0.25, PiApp = 5.1
	// Apagado: 8212.5 h, 15, tdelta=10, nc = 365, phi = 22.5, grms=0, PiApp = 3.1

	fmt.Printf("%f\n", fides.SemiconductorFIT(comp, mission, 0.5))
}
