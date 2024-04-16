package fides

//	"errors"
//	"math"

func PiezoFIT(comp *Component, mission *Mission) (float64, error) {

	return 0, nil

}

func lbase_piezo(c *Component) (float64, float64, float64, float64, float64) {

	osc := contains(c.Tags, "osc") || contains(c.Tags, "oscillator")

	if contains(c.Tags, "smd") {
		if osc {
			return 1.63, 0.31, 0.53, 0.07, 0.09
		} else {
			return 0.79, 0.16, 0.59, 0.15, 0.1
		}
	}

	// THT
	if osc {
		return 1.6, 0.32, 0.42, 0.14, 0.12
	}

	return 0.82, 0.16, 0.46, 0.27, 0.11
}
