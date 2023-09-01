package fides

import (
	_ "embed"
	"log"
	"math"
	"strings"

	"github.com/rveen/golib/document"
	"github.com/rveen/ogdl"
)

type Package struct {
	name    string
	npins   int
	class   string
	rjaLow  float64
	rjaHigh float64
	rjc     float64

	l0rh       float64
	l0tcCase   float64
	l0tcSolder float64
	l0mech     float64
}

var packages map[string]*Package

//go:embed data.md
var datamd string

var data *ogdl.Graph

func init() {
	doc, _ := document.New(datamd)
	data = doc.Data()

	packages = make(map[string]*Package)
	pkgs := data.Get("packages")

	log.Printf(pkgs.Text())

	for _, p := range pkgs.Out {

		pkg := &Package{}

		pkg.name = p.ThisString()
		pkg.npins = int(p.Get("npins").Int64())

		pkg.l0rh = p.Get("l_0_rh").Float64()
		pkg.l0tcCase = p.Get("l_0_tc_case").Float64()
		pkg.l0tcSolder = p.Get("l_0_tc_solder").Float64()
		pkg.l0mech = p.Get("l_0_mech").Float64()

		pkg.rjaLow = p.Get("rja_l").Float64()
		pkg.rjaHigh = p.Get("rja_h").Float64()
		pkg.rjc = p.Get("rjc").Float64()

		packages[pkg.name] = pkg

		// Get equivalents

		eq := p.Get("equivalents").String()
		eqss := strings.Split(eq, ",")
		for _, equiv := range eqss {
			e := strings.TrimSpace(equiv)
			packages[e] = pkg
		}

	}
}

func Lcase_semi(pkg string) (float64, float64, float64, float64) {

	pkg = strings.ToUpper(pkg)

	p := packages[pkg]

	if p != nil {
		return p.l0rh, p.l0tcCase, p.l0tcSolder, p.l0mech
	}

	return -1, -1, -1, -1
}

// give package and pins, return: l0rh, l0tc_case, l0tc_solder, l0mech
func Lbase_case(pkg string, n int) (float64, float64, float64, float64) {

	var arh, brh, atc, btc, ats, bts, am, bm float64

	switch pkg {

	case "pdip":
		arh = -5.88
		brh = 0.94
		atc = 9.85
		btc = 1.35
		ats = 8.24
		bts = 1.35
		am = 12.85
		bm = 1.35

	case "cdip":
		atc = 6.77
		btc = 1.35
		if n < 21 {
			ats = 5.16
			am = 8.38
		} else {
			ats = 4.47
			am = 7.69
		}
		bts = 1.35
		bm = 1.35

	case "pqfp":
		arh = 11.16
		brh = 1.76
		atc = 12.41
		btc = 1.46
		if n < 44 || n > 304 {
			return -1, -1, -1, -1
		} else if n < 241 {
			ats = 10.8
			am = 14.71
		} else {
			ats = 10.11
			am = 14.02
		}
		bts = 1.46
		bm = 1.46

	case "sqfp":
		fallthrough
	case "tqfp":
		fallthrough
	case "vqfp":
		fallthrough
	case "lqfp":
		arh = 7.75
		brh = 1.13
		atc = 8.57
		btc = 0.73
		if n < 32 || n > 208 {
			return -1, -1, -1, -1
		} else if n < 121 {
			ats = 6.96
			am = 11.57
		} else {
			ats = 5.57
			am = 10.18
		}
		bts = 0.73
		bm = 0.73

	case "cerpack": // TODO
	case "plcc": // TODO
	case "jclcc": // TODO
	case "clcc": // TODO
	case "soj": // TODO

	case "so":
		fallthrough
	case "sow":
		fallthrough
	case "sop":
		fallthrough
	case "sol":
		fallthrough
	case "soic":

		arh = 8.23
		brh = 1.17
		atc = 13.35
		btc = 2.18
		bts = 2.18
		bm = 2.18

		if n < 16 { // Rolf extended to 1 pin
			ats = 11.75
			am = 16.36
		} else if n < 20 {
			ats = 11.06
			am = 15.66
		} else if n < 32 {
			ats = 10.36
			am = 14.97
		} else {
			ats = 10.14
			am = 14.75
		}
	case "sot":
		fallthrough // TEMPORAL : TODO
	case "tsop":
		if n < 17 { // Rolf: extended to 1 pin
			ats = 7.44
			am = 12.05
		} else if n < 33 {
			ats = 6.05
			am = 10.66
		} else if n < 45 {
			ats = 5.83
			am = 10.44
		} else {
			ats = 5.36
			am = 9.97
		}

		arh = 6.21
		brh = 0.97
		atc = 9.05
		btc = 0.76
		bts = 0.76
		bm = 0.76

	case "ssop":
		fallthrough
	case "vsop":
		fallthrough
	case "qsop":
		arh = 11.95
		brh = 2.23
		atc = 16.28
		btc = 2.6
		ats = 14.67
		bts = 2.6
		am = 19.28
		bm = 2.6

	case "tssop":
		fallthrough
	case "msop":
		if n >= 8 && n < 29 {
			ats = 13.95
			am = 18.56
		} else if n > 28 && n < 49 {
			ats = 13.21
			am = 17.86
		} else if n == 56 {
			ats = 12.56
			am = 17.17
		} else if n == 64 {
			ats = 12.16
			am = 16.76
		} else {
			return -1, -1, -1, -1
		}

		arh = 11.57
		brh = 2.22
		atc = 15.56
		btc = 2.66
		bts = 2.66
		bm = 2.66

	case "qfn":
		fallthrough
	case "dfn":
		fallthrough
	case "mlf":
		if n < 8 || n > 72 {
			return -1, -1, -1, -1
		} else if n < 25 {
			ats = 8.12
			am = 11.34
		} else if n < 57 {
			ats = 7.9
			am = 11.12
		} else {
			ats = 7.71
			am = 10.93
		}
		arh = 8.97
		brh = 1.14
		atc = 11.2
		btc = 1.21
		bts = 1.14
		bm = 1.21

	case "pbga_0_8": // TODO
	case "pbga_0_8_flex": // TODO
	case "pbga_1_0": // TODO
	case "pbga_1_27": // TODO
	case "powerbga": // TODO
	case "cbga": // TODO
	case "dbga": // TODO
	case "cicga": // TODO
	case "cpga": // TODO

	}

	if arh != 0 {
		arh = math.Exp(-arh) * math.Pow(float64(n), brh)
	}
	atc = math.Exp(-atc) * math.Pow(float64(n), btc)
	ats = math.Exp(-ats) * math.Pow(float64(n), bts)
	am = math.Exp(-am) * math.Pow(float64(n), bm)

	return arh, atc, ats, am
}

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

func Rthja_semi(pkg string, k bool) (int, float64, float64) {

	for _, p := range packages {
		if pkg == p.name {
			if k {
				return p.npins, p.rjaHigh, p.rjc
			} else {
				return p.npins, p.rjaLow, p.rjc
			}
		}
	}

	return -1, -1, -1
}
