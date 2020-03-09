package fides

type Component struct {
	Class            string
	Type             string
	Package          string
	N                int
	Rth              float64
	IsAnalog         bool
	IsPower          bool
	IsInterface      bool
	Value            float64
	V, P, I          float64
	Vmax, Pmax, Imax float64
	T, Tmax          float64
}
