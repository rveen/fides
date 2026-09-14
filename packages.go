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

// Rtha returns the default junction-to-ambient thermal resistance of the
// package in K/W, on a substrate with the given thermal conductivity in
// W/(m·K): Ctype·Np^-0.58·K for integrated circuit packages (FIDES 2022,
// p. 119), else the value of the package table (p. 120). It returns -1 if
// there is none: an IC package without a pin count, QFN (whose formula needs
// the package area), or a discrete package without a table value.
//
// It takes the name and pin count that NewPackage keeps apart for IC
// packages (SOIC8: SOIC, 8), and the table row of discrete packages and
// their equivalents.
func (p *Package) Rtha(tcSubstrate float64) float64 {

	if ctype := rthBase(p.Name); ctype > 0 {
		if p.Npins <= 0 {
			return -1
		}
		k := 1.15
		if tcSubstrate >= 15 {
			k = 0.94
		}
		return ctype * math.Pow(float64(p.Npins), -0.58) * k
	}

	rth := p.rjaLow
	if tcSubstrate >= 15 {
		rth = p.rjaHigh
	}
	if math.IsNaN(rth) || rth <= 0 {
		return -1
	}
	return rth
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
	case "TSOP":
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
		if n < 8 || n > 80 {
			return -1, -1, -1, -1
		} else if n < 25 {
			ats = 6.68
			am = 9.9
		} else if n < 57 {
			ats = 6.17
			am = 9.38
		} else {
			// 64 to 80 pins (FIDES 2022, p. 126)
			ats = 5.95
			am = 9.64
		}
		arh = 8.84
		brh = 0.77
		atc = 12.03
		btc = 0.94

	// The following packages up to QFN_04 complete the table of FIDES 2022,
	// pp. 125-127. A "-" for λ0RH (hermetic packages) is arh = 0. Fine-pitch
	// PBGAs (< 1 mm), BGA WLPs and FC-PBGA 0.8 mm cannot be named by a
	// package name and a number of pins, and are not supported.

	case "HQFP", "RQFP", "POWERQFP": // Power QFP
		if n < 160 || n > 304 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 14.17, 2.41, 11.80, 1.36
		if n <= 240 {
			ats, am = 8.29, 12.21
		} else {
			ats, am = 7.96, 11.87
		}

	case "CERPACK":
		if n < 20 || n > 56 {
			return -1, -1, -1, -1
		}
		atc, btc, ats, am = 8.14, 1.01, 8.29, 11.51

	case "CQFP":
		if n < 20 || n > 256 {
			return -1, -1, -1, -1
		}
		atc, btc = 8.14, 1.01
		if n <= 132 {
			ats, am = 8.29, 11.51
		} else {
			ats, am = 6.68, 9.90
		}

	case "PLCC":
		if n < 20 || n > 84 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 10.50, 1.92, 19.45, 3.29
		if n <= 52 {
			ats, am = 8.29, 12.43
		} else {
			ats, am = 7.20, 11.29
		}

	case "JLCC", "JCLCC": // J-lead ceramic leaded chip carrier
		if n < 4 || n > 84 {
			return -1, -1, -1, -1
		}
		atc, btc = 8.07, 0.93
		switch {
		case n <= 32:
			ats, am = 8.29, 11.51
		case n <= 44:
			ats, am = 7.95, 11.17
		case n <= 52:
			ats, am = 7.19, 10.41
		case n <= 68:
			ats, am = 6.21, 9.43
		default:
			ats, am = 5.58, 8.80
		}

	case "CLCC": // ceramic leadless chip carrier
		if n < 4 || n > 84 {
			return -1, -1, -1, -1
		}
		atc, btc = 8.07, 0.93
		switch {
		case n <= 10:
			ats, am = 7.19, 10.41
		case n <= 20:
			ats, am = 5.89, 9.11
		case n <= 32:
			ats, am = 5.58, 8.80
		case n <= 52:
			ats, am = 5.07, 8.29
		default:
			ats, am = 4.36, 7.58
		}

	case "SOJ":
		if n < 24 || n > 44 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc, ats, am = 4.33, 0.83, 8.76, 1.49, 8.29, 12.90

	case "LGA": // plastic land grid array
		if n < 6 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 5.3, 0.84, 9.02, 1.02
		if n <= 20 {
			ats, am = 6.68, 9.90
		} else {
			ats, am = 6.17, 9.38
		}

	case "PBGA": // plastic BGA, 1.27 mm pitch
		if n < 119 || n > 729 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 6.85, 0.80, 10.29, 0.86
		switch {
		case n <= 352:
			ats, am = 7.20, 10.87
		case n <= 432:
			ats, am = 6.68, 10.37
		default:
			ats, am = 6.17, 9.85
		}

	case "PBGABT": // plastic BGA BT, 1.00 mm pitch
		if n < 64 || n > 1156 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 4.77, 0.40, 10.46, 0.76
		if n <= 484 {
			ats, am = 6.68, 10.37
		} else {
			ats, am = 6.17, 9.85
		}

	case "POWERBGA", "TBGA", "SBGA": // power BGA, 1.27 mm pitch
		if n < 256 || n > 956 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 5.59, 0.61, 11.01, 0.79
		if n <= 352 {
			ats, am = 7.20, 10.87
		} else {
			ats, am = 6.68, 10.37
		}

	case "FCBGA": // flip chip PBGA, 1 mm pitch
		if n > 1704 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 8.39, 0.79, 9.82, 0.52
		if n < 672 {
			ats, am = 7.20, 10.87
		} else {
			ats, am = 6.68, 10.37
		}

	case "CBGA":
		if n < 255 || n > 1156 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc = 4.41, 0.63, 10.60, 1.12
		if n <= 560 {
			ats, am = 5.12, 8.80
		} else {
			ats, am = 4.36, 7.35
		}

	case "DBGA": // dimpled BGA
		if n < 255 || n > 1156 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc, ats, am = 4.41, 0.63, 10.60, 1.12, 6.68, 10.37

	case "CCGA", "CICGA": // ceramic land GA + interposer, ceramic column GA
		if n < 255 || n > 1156 {
			return -1, -1, -1, -1
		}
		arh, brh, atc, btc, ats, am = 4.41, 0.63, 10.60, 1.12, 5.95, 9.64

	case "CPGA":
		if n < 68 || n > 655 {
			return -1, -1, -1, -1
		}
		atc, btc = 8.63, 1.02
		if n <= 250 {
			ats, am = 7.96, 11.62
		} else {
			ats, am = 6.68, 10.37
		}

	case "QFNFP", "QFN_04": // QFN 0.4 mm pitch; the name QFN_04 cannot carry a number of pins
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

// rthBase is Ctype of the default junction-to-ambient thermal resistance of
// integrated circuit packages, Ctype·Np^-0.58·K (FIDES 2022, p. 119). QFN
// packages have none here: their formula takes the package area in mm², not
// the number of pins.
func rthBase(pkg string) float64 {

	switch pkg {

	case "CDIP", "CERDIP":
		return 320
	case "RQFP", "HQFP", "POWERQFP":
		return 340
	case "PDIP":
		return 360
	case "PPGA":
		return 380
	case "PLCC":
		return 390
	case "SO", "SOIC", "SOJ":
		return 400
	case "CPGA", "SOP":
		return 410
	case "POWERBGA", "SBGA", "TBGA":
		return 450
	case "JLCC", "JCLCC":
		return 470
	case "LQFP", "VQFP", "TQFP", "CERPACK":
		return 480
	case "FCBGA":
		return 520
	case "PBGA":
		return 530
	case "SSOP", "CQFP":
		return 560
	case "PQFP":
		return 570
	case "TSSOP":
		return 650
	case "PBGABT":
		return 670
	case "TSOP":
		return 750
	case "CBGA":
		return 780
	case "LGA":
		return 260
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

// Rthja_semi returns the number of pins, the junction-to-ambient thermal
// resistance (on a substrate of high thermal conductivity if k is true) and
// the junction-to-case thermal resistance of a discrete package or one of its
// equivalents, from the package table; -1, -1, -1 if it is not there.
func Rthja_semi(pkg string, k bool) (int, float64, float64) {

	p := packages[pkg]
	if p == nil {
		return -1, -1, -1
	}
	if k {
		return p.Npins, p.rjaHigh, p.rjc
	}
	return p.Npins, p.rjaLow, p.rjc
}
