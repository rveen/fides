package fides

type Component struct {
	Type        string
	Package     string
	N           int
	Rth         float64
	IsAnalog    bool
	IsPower     bool
	IsInterface bool
	Value       float64
	Power       float64
	Voltage     float64
	Current     float64
}
