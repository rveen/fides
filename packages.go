package fides

import (
	_ "embed"
    "log"
	"math"
	"strconv"
	"strings"

	"github.com/rveen/golib/csv"
	"github.com/rveen/ogdl"
)

type Package struct {
	Name    string
	Npins   int
	Tags    []string
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

//go:embed data.csv
var datacsv string

var data *ogdl.Graph

func NewPackage(name string) *Package {

	p := packages[name]

	if p != nil {
		return p
	}

	s, n := splitPkg(name)
	p = &Package{}
	p.Name = s
	p.Npins = n
	p.l0rh, p.l0tcCase, p.l0tcSolder, p.l0mech = lbase_case(s, n)

	if p.l0rh < 0 {
	    log.Printf("package not found [%s].\n", name)
    }

	return p
}

func (p *Package) FitBase() (float64, float64, float64, float64) {
	return p.l0rh, p.l0tcCase, p.l0tcSolder, p.l0mech
}

func (p *Package) Rtha(tcSusbtrate float64) float64 {
	return rthja(p.Name, tcSusbtrate)
}

func init() {

	pkgs, _ := csv.ReadString(datacsv)

	packages = make(map[string]*Package)

	for _, p := range pkgs {

		pkg := &Package{}

		pkg.Name = p["name"]

		s := p["npins"]
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			pkg.Npins = int(n)
		}

		pkg.Tags = strings.Fields(p["tags"])

		pkg.l0rh = float(p["l0rh"])
		pkg.l0tcCase = float(p["l0tc_case"])
		pkg.l0tcSolder = float(p["l0tc_solder"])
		pkg.l0mech = float(p["l0mech"])
		pkg.rjaLow = float(p["rja_l"])
		pkg.rjaHigh = float(p["rja_h"])
		pkg.rjc = float(p["rjc"])

		packages[pkg.Name] = pkg

		// Get equivalents

		eq := p["equivalents"]
		ss := strings.Fields(eq)
		for _, s := range ss {
			packages[s] = pkg
		}
	}
}

func float(s string) float64 {

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}
	return f
}

func splitPkg(s string) (string, int) {

	var sb strings.Builder
	var nb strings.Builder

	for i, c := range s {
		if c == '-' || (c >= '0' && c <= '9') {
			s = s[i:]
			break
		}
		sb.WriteRune(c)
	}

	for _, c := range s {
		if c == '-' {
			continue
		}
		if c < '0' || c > '9' {
			break
		}
		nb.WriteRune(c)
	}

	pkg := strings.ToUpper(sb.String())
	n, _ := strconv.Atoi(nb.String())
	return pkg, n
}

func lbase_pkg(pkg string) (float64, float64, float64, float64) {

	pkg = strings.ToUpper(pkg)

	p := packages[pkg]

	if p != nil {
		return p.l0rh, p.l0tcCase, p.l0tcSolder, p.l0mech
	}

	s, n := splitPkg(pkg)
	return lbase_case(s, n)
}

// give package and pins, return: l0rh, l0tc_case, l0tc_solder, l0mech
func lbase_case(pkg string, n int) (float64, float64, float64, float64) {

	var arh, brh, atc, btc, ats, bts, am, bm float64

	bts = 0.92
	bm = 0.92

	switch pkg {

	case "PDIP":
		arh = 6.27
		brh = 0.69
		atc = 10.23
		btc = 0.95
		ats = 8.29
		am = 12.9

	case "CDIP", "CERDIP":
		atc = 12.68
		btc = 2.27
		if n < 21 {
			ats = 8.29
			am = 11.51
		} else {
			ats = 7.96
			am = 11.18
		}

	case "PQFP":
		arh = 10.94
		brh = 1.57
		atc = 13.72
		btc = 1.62
		if n < 44 || n > 304 {
			return -1, -1, -1, -1
		} else if n < 241 {
			ats = 8.29
			am = 12.21
		} else {
			ats = 7.96
			am = 11.87
		}

	case "TQFP", "LQFP":
		arh = 6.62
		brh = 0.52
		atc = 13.05
		btc = 1.3
		if n < 32 || n > 208 {
			return -1, -1, -1, -1
		} else if n < 121 {
			ats = 8.29
			am = 12.9
		} else {
			ats = 7.2
			am = 11.8
		}

	case "SO", "SOIC":
		arh = 11.45
		brh = 1.95
		atc = 16.8
		btc = 2.94

		if n < 20 { // Rolf extended to 1 pin
			ats = 8.29
			am = 12.9
		} else {
			ats = 7.96
			am = 12.56
		}
	case "SOT", "TSOP":
		if n < 17 { // Rolf: extended to 1 pin
			ats = 8.29
			am = 12.9
		} else if n < 33 {
			ats = 7.2
			am = 11.8
		} else if n < 45 {
			ats = 6.68
			am = 11.29
		} else {
			ats = 6.17
			am = 10.77
		}

		arh = 6.87
		brh = 1.1
		atc = 9.6
		btc = 0.83

	case "SSOP", "QSOP":
		arh = 17.7
		brh = 3.35
		atc = 20.88
		btc = 3.38
		ats = 7.96
		am = 12.56

	case "TSSOP", "MSOP", "MINISO", "HTSSOP", "VSSOP": // rolf added HTSSOP, VSSOP
		arh = 11.25
		brh = 1.57
		atc = 14.93
		btc = 1.87

		if n < 29 {
			ats = 8.29
			am = 12.9
		} else if n > 28 && n < 49 {
			ats = 7.96
			am = 12.56
		} else if n == 56 {
			ats = 7.2
			am = 11.8
		} else {
			ats = 6.17
			am = 10.77
		}

	case "QFN", "DFN", "VQFN": // rolf added VQFN
		if n < 8 || n > 72 {
			return -1, -1, -1, -1
		} else if n < 25 {
			ats = 6.68
			am = 9.9
		} else if n < 57 {
			ats = 6.17
			am = 9.38
		} else {
			ats = 5.95
			am = 9.17
		}
		arh = 8.84
		brh = 0.77
		atc = 12.03
		btc = 0.94

	case "QFN_04":
		arh = 6.22
		brh = 0.78
		atc = 9.65
		btc = 0.91
		if n <= 40 {
			ats = 6.17
			am = 9.38
		} else {
			ats = 5.95
			am = 9.17
		}

	default:
		return -1, -1, -1, -1

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
// for known semiconductor packages, according to tables in FIDES 2022.
// The package name implies/contains the number of pins.
//
// K is a constant that depends on the substrate's thermal conductivity:
// k==false: K = 1.15 (low conductivity, <15 W/mK) )
// k==true: K = 0.94 (high conductivity >= 15 W/mK)
func rthja(pkg string, tcSusbtrate float64) float64 {

	p, n := splitPkg(pkg)

	rth := rthBase(p)
	if rth > 0 {

		K := 1.15
		if tcSusbtrate >= 15 {
			K = 0.94
		}

		return rth * math.Pow(float64(n), -0.58) * K
	} else {

		k := false
		if tcSusbtrate >= 15 {
			k = true
		}

		_, rth, _ = Rthja_semi(pkg, k)
		return rth
	}
}

func rthBase(pkg string) float64 {

	switch pkg {

	case "QFN":
		return 223
	case "CDIP", "CERDIP":
		return 320
	case "RQFP", "HQFP":
		return 340
	case "PDIP":
		return 360
	case "PPGA":
		return 380
	case "PLCC":
		return 390
	case "SOIC", "SOJ":
		return 400
	case "CPGA":
		return 410
	case "SBGA":
		fallthrough
	case "TBGA":
		return 450
	case "JCLCC":
		return 470
	case "LQFP", "VQFP", "TQFP", "CERPACK", "CBGA":
		return 480
	case "PBGA1.27":
		return 530
	case "SBGA-error":
		fallthrough
	case "TBGA-error":
		return 550
	case "SSOP", "CQFP":
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

func IsSmd(c *Component) bool {

	pkg, _ := splitPkg(c.Package)

	switch pkg {

	case "CDIP", "CERDIP", "PDIP", "TO", "DO", "DIL", "SIL", "SIP", "DIP":
		return false
	}

	return true
}

func Rthja_semi(pkg string, k bool) (int, float64, float64) {

	for _, p := range packages {
		if pkg == p.Name {
			if k {
				return p.Npins, p.rjaHigh, p.rjc
			} else {
				return p.Npins, p.rjaLow, p.rjc
			}
		}
	}

	return -1, -1, -1
}
