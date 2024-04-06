package fides

// Class = { R, C, L, J, U, Q, D, PCB, SW, RL, X, LED, OPTO, PIEZO?, FUSE }
//
// Type (space separated tags):
//
//   R: thin, thick, fuse, ptc, ntc
//   C: ceramic, film, type1, type2, x7r, x5r, np0
//   L: trafo, trafo, choke, wirewound, multilayer, ferrite, power
//   D: zener, schottky, power
//   Q: mos, gan, sic,
//   U: opto, asic
//   FUSE = R fuse
//   OPTO = U opto

type Component struct {
	Name        string
	Value       float64
	Code        string
	Description string

	Class string
	Tags  string // TODO Change to []string

	Package     string
	N           int // Number of devices
	Np          int // TODO Number of pins
	Rth         float64
	IsAnalog    bool
	IsPower     bool
	IsInterface bool

	Vp, V, P, I, T         float64 // Working conditions
	Vmax, Pmax, Imax, Tmax float64 // Device limits

	// Temperature coefficient. Set to NaN for undefined!
	TC float64

	FIT float64
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
