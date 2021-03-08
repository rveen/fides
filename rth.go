package fides

import (
	"math"
	"strings"
)

// Rthermal returns the Rja (thermal resistance from junction to ambient)
// for known semiconductor packages, according to tables in FIDES 2009.
// The package name implies/contains the number of pins.
//
// K is a constant that depends on the substrate's thermal conductivity:
// k==false: K = 1.15 (low conductivity (FR4?) )
// k==true: K = 0.94 (high conductivity)
//
func Rthja(pkg string, k bool) float64 {

	pkg = strings.ToUpper(pkg)

	n := 0
	K := 1.15
	if k {
		K = 0.94
	}

	return rthBase(pkg) * math.Pow(float64(n), -0.58) * K
}

func rthBase(pkg string) float64 {

	switch pkg {

	case "QFN":
		return 223
	case "CERDIP":
		fallthrough
	case "CDIP":
		return 320
	case "HQFP":
		fallthrough
	case "RQFP":
		return 340
	case "PDIP":
		return 360
	case "PPGA":
		return 380
	case "PLCC":
		return 390
	case "SOIC":
		fallthrough
	case "SOJ":
		return 400
	case "CPGA":
		return 410
	case "SBGA":
		fallthrough
	case "TBGA":
		return 450
	case "JCLCC":
		return 470
	case "CBGA":
		fallthrough
	case "CERPACK":
		fallthrough
	case "TQFP":
		fallthrough
	case "VQFP":
		fallthrough
	case "LQFP":
		return 480
	case "PBGA1.27":
		return 530
	case "SBGA-error":
		fallthrough
	case "TBGA-error":
		return 550
	case "SSOP":
		return 560
	case "CQFP":
		return 560
	case "PQFP":
		return 570
	case "TSSOP":
		return 650
	case "PBGA1.0":
		return 670
	case "PBGA0.8":
		return 700
	case "TSOP":
		return 750
	}

	return -1
}

type Package struct {
	name    string
	pins    int
	class   string
	rjaLow  float64
	rjaHigh float64
	rjc     float64
}

// Convention: use hyphen coherently
var packages = []Package{
	{"DO-15", 2, "", 60, 42, 5},
	{"DO-204AC", 2, "", 60, 42, 5},

	{"DO-27", 2, "", 41, 30, 1},
	{"DO-201AA", 2, "", 41, 30, 1},

	{"DO-35", 2, "", 378, 241, 134},
	{"DO-204AH", 2, "", 378, 241, 134},

	{"DO-41", 2, "", 73, 50, 45},
	{"DO-204AL", 2, "", 73, 50, 45},

	{"DO-92", 3, "", 195, 126, 150},
	{"DO-220", 3, "", 65, 45, 4},

	{"DPAK", 4, "", 97, 71, 4},
	{"TO-252", 4, "", 97, 71, 4},
	{"TO-252AA", 4, "", 97, 71, 4},
	{"SC-63", 4, "", 97, 71, 4},
	{"SOT-428", 4, "", 97, 71, 4},

	{"D2PAK", 4, "", 58, 40, 1},
	{"TO-263", 4, "", 58, 40, 1},
	{"SC-83A", 4, "", 58, 40, 1},
	{"SMD-220", 4, "", 58, 40, 1},

	{"IPACK", 3, "", 96, 50, 3},
	{"TO-251AA", 3, "", 96, 50, 3},

	{"I2PAK", 3, "", 63, 44, 1},

	{"SMA-J", 2, "", 110, 73, 41},
	{"SMB-J", 2, "", 88, 59, 27},
	{"SMC-J", 2, "", 67, 46, 2},

	{"SOD-80", 2, "", 568, 361, 172},
	{"Mini-MELF", 2, "", 568, 361, 172},
	{"DO-213AA", 2, "", 568, 361, 172},

	{"SOD-100", 2, "", 315, 202, 119},
}

func rthja(pkg string, k bool) float64 {
	switch pkg {

	}
	return -1
}
