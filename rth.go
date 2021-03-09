package fides

import (
	"math"
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

	p, n := splitPkg(pkg)

	rth := rthBase(p)
	if rth > 0 {

		K := 1.15
		if k {
			K = 0.94
		}

		return rthBase(p) * math.Pow(float64(n), -0.58) * K
	} else {
		_, rth, _ = Rthja_semi(pkg, k)
		return rth
	}
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

// Convention: do not use hyphen
var packages = []Package{

	{"do15", 2, "", 60, 42, 5},
	{"do201aa", 2, "", 60, 42, 5},

	{"do27", 2, "", 41, 30, 1},
	{"do201aa", 2, "", 41, 30, 1},

	{"do35", 2, "", 378, 241, 134},
	{"do204ah", 2, "", 378, 241, 134},

	{"do41", 2, "", 73, 50, 45},
	{"do204al", 2, "", 73, 50, 45},

	{"do92", 3, "", 195, 126, 150},
	{"do220", 3, "", 65, 45, 4},

	{"dpack", 4, "", 97, 71, 4},
	{"to252", 4, "", 97, 71, 4},
	{"to252AA", 4, "", 97, 71, 4},
	{"sc63", 4, "", 97, 71, 4},
	{"sot428", 4, "", 97, 71, 4},

	{"d2pak", 4, "", 58, 40, 1},
	{"to263", 4, "", 58, 40, 1},
	{"sc83a", 4, "", 58, 40, 1},
	{"smd220", 4, "", 58, 40, 1},

	{"ipack", 3, "", 96, 50, 3},
	{"to251aa", 3, "", 96, 50, 3},

	{"i2pk", 3, "", 63, 44, 1},

	{"smbj", 2, "", 88, 59, 27},
	{"do214aa", 2, "", 88, 59, 27},
	{"smbj", 2, "", 88, 59, 27},

	{"smcj", 2, "", 67, 46, 2},
	{"do214ab", 2, "", 67, 46, 2},
	{"do15", 2, "", 67, 46, 2},

	{"smaj", 2, "", 110, 73, 41},
	{"do214ac", 2, "", 110, 73, 41},
	{"sod87", 2, "", 110, 73, 41},

	{"sod80", 2, "", 568, 361, 172},
	{"minimelf", 2, "", 568, 361, 172},
	{"do213aa", 2, "", 568, 361, 172},

	{"sod100", 2, "", 315, 202, 119},
	{"sod123", 2, "", 337, 216, 130},
	{"sod323", 2, "", 428, 273, 146},
	{"sod523", 2, "", 93, 62, 31},

	{"sot23", 3, "", 443, 360, 130},
	{"sot23-5", 5, "", 285, 136, 106},
	{"sot25", 5, "", 285, 136, 106},
	{"sot23-6", 5, "", 212, 133, 110},
	{"sot26", 5, "", 212, 133, 110},

	{"sot82", 3, "", 100, 67, 8},
	{"to225", 3, "", 100, 67, 8},
	{"sot89", 4, "", 142, 125, 100},
	{"sot90b", 6, "", 500, 318, 160},
	{"sot143", 4, "", 473, 250, 155},
	{"sot223", 4, "", 84, 57, 21},
	{"sot323", 3, "", 516, 328, 164},
	{"sot343", 4, "", 215, 139, 88},
	{"sot346", 3, "", 500, 318, 160},
	{"sot353", 5, "", 358, 229, 144},
	{"sot363", 6, "", 553, 351, 164},

	{"to18", 3, "", 475, 302, 150},
	{"to71", 3, "", 475, 302, 150},
	{"to72", 3, "", 475, 302, 150},
	{"sot31", 3, "", 475, 302, 150},
	{"sot18", 3, "", 475, 302, 150},

	{"to39", 3, "", 219, 142, 58},
	{"sot5", 3, "", 219, 142, 58},

	{"to92", 3, "", 180, 117, 66},
	{"sot54", 3, "", 180, 117, 66},
	{"sc43", 3, "", 180, 117, 66},
	{"to226aa", 3, "", 180, 117, 66},

	{"to126", 3, "", 95, 64, 3},
	{"sot32", 3, "", 95, 64, 3},
	{"to225aa", 3, "", 95, 64, 3},

	{"to218", 3, "", 40, 29, 1},
	{"isowatt218", 3, "", 95, 64, 3},

	{"to220", 3, "", 58, 40, 4},
	{"to247", 3, "", 47, 34, 1},
}

func Rthja_semi(pkg string, k bool) (int, float64, float64) {

	for _, p := range packages {
		if pkg == p.name {
			if k {
				return p.pins, p.rjaHigh, p.rjc
			} else {
				return p.pins, p.rjaLow, p.rjc
			}
		}
	}

	return -1, -1, -1
}
