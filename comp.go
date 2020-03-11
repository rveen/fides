package fides

type Component struct {
	Name             string
	Code             string
	Class            string
	Type             string
	Description      string
	Package          string
	N                int
	Rth              float64
	IsAnalog         bool
	IsPower          bool
	IsInterface      bool
	Value            float64
	Vp, V, P, I      float64
	Vmax, Pmax, Imax float64
	T, Tmax          float64
	// Temperature coefficient. Set to NaN for undefined!
	TC   float64
	FIT  float64
	Cost float64
}
