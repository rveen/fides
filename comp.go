package fides

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Component struct {
	Name        string  // Denominator or reference
	Value       float64 // Value in SI unit
	Tolerance   float64
	Code        string // Part number
	Description string // Free description

	Class string
	Tags  []string
	Block string // For classifying per block or function

	Package string
	N       int // Number of devices
	Np      int // TODO Number of pins
	Rtha    float64

	Vp, V, P, I, T                float64 // Working conditions (T is the delta over ambient)
	Vpmax, Vmax, Pmax, Imax, Tmax float64 // Device limits

	// Temperature coefficient. Set to NaN for undefined
	TC float64

	FIT float64
}

type Bom struct {
	Components []*Component
}

func (bom *Bom) FromCsv(file string) error {

	m, err := csvRead(file)
	if err != nil {
		return err
	}

	// Check what fields are in this file, and update only those
	// If 'code' is the first field, consider it as a component database
	// If 'name' is the first field, then merge data into that component (or create it)

	byCode := true
	if _, ok := m[0]["name"]; ok {
		byCode = false
	}

	for i := 0; i < len(m); i++ {
		r := m[i]

		// get component or insert

		var cc []*Component
		if byCode {
			cc = getComps(r["code"], bom, byCode)
		} else {
			cc = getComps(r["name"], bom, byCode)
		}

		for _, c := range cc {

			if val, ok := r["class"]; ok {
				c.Class = strings.ToUpper(val)
			}
			if val, ok := r["code"]; ok {
				c.Code = val
			}
			if val, ok := r["block"]; ok {
				c.Block = val
			}
			if val, ok := r["tags"]; ok {
				c.Tags = strings.Fields(val)
			}
			if val, ok := r["value"]; ok {
				c.Value = getValue(val)
			}
			if val, ok := r["tolerance"]; ok {
				c.Tolerance = getValue(val)
			}
			if val, ok := r["description"]; ok {
				c.Description = val
			}
			if val, ok := r["package"]; ok {
				c.Package = strings.ToUpper(val)
			}
			if val, ok := r["ndevices"]; ok {
				c.N, _ = strconv.Atoi(val)
			}
			if val, ok := r["npins"]; ok {
				c.Np, _ = strconv.Atoi(val)
			}
			if val, ok := r["vmax"]; ok {
				c.Vmax, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["v"]; ok {
				c.V, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["vpmax"]; ok {
				c.Vpmax, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["vp"]; ok {
				c.Vp, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["pmax"]; ok {
				c.Pmax, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["p"]; ok {
				c.P, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["imax"]; ok {
				c.Imax, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["i"]; ok {
				c.I, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["tmax"]; ok {
				c.Tmax, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["t"]; ok {
				c.T, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["rtha"]; ok {
				c.Rtha, _ = strconv.ParseFloat(val, 64)
			}
			if val, ok := r["tc"]; ok {
				c.TC, _ = strconv.ParseFloat(val, 64)
			}
		}
	}

	return nil
}

func (c *Component) ToCsv() string {

	s := "name, class, tags, code, value, tolerance\n"

	s += fmt.Sprintf("%s, ", c.Name)
	s += fmt.Sprintf("%s, ", c.Class)
	s += fmt.Sprintf("%s, ", c.Tags)
	s += fmt.Sprintf("%s, ", c.Code)
	s += fmt.Sprintf("%.12f, ", c.Value)
	s += fmt.Sprintf("%f, ", c.Tolerance)

	return s + "\n"
}

func getComps(key string, bom *Bom, byCode bool) []*Component {

	var cc []*Component

	for _, c := range bom.Components {

		if byCode && c.Code == key {
			cc = append(cc, c)
		} else if !byCode && c.Name == key {
			cc = append(cc, c)
			break
		}
	}

	if len(cc) == 0 && !byCode {
		c := &Component{Name: key}
		cc = append(cc, c)
		bom.Components = append(bom.Components, c)
	}

	return cc
}

// get the value out of a string
// 9.1 k == 9k1 = 9100
func getValue(s string) float64 {

	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	}

	s = strings.ToLower(s)

	var v1 []rune
	var v2 []rune
	var k []rune

	first := true

	for i, c := range s {
		if (c >= '0' && c <= '9') || c == '.' {
			if first {
				v1 = append(v1, c)
			} else {
				v2 = append(v2, c)
			}
		} else if c == ' ' || c == '\t' {
			continue
		} else {
			if i == 0 {
				return math.NaN()
			}
			k = append(k, c)
			first = false
		}
	}

	n1, _ := strconv.ParseFloat(string(v1), 64)

	n2 := 0.0
	if len(v2) != 0 {
		n2, _ = strconv.ParseFloat(string(v2), 64)
	}

	n1 = n1 + n2/10.0

	ks := string(k)

	switch ks {
	case "k":
		return n1 * 1e3
	case "m":
		return n1 * 1e6
	case "meg":
		return n1 * 1e6
	case "":
		return n1
	case "u":
		fallthrough
	case "uf":
		return n1 * 1e-6
	case "n":
		fallthrough
	case "nf":
		return n1 * 1e-9
	case "p":
		fallthrough
	case "pf":
		return n1 * 1e-12
	default:
		return n1
	}
}
