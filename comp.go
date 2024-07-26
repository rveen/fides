package fides

import (
	"fmt"
	e "github.com/rveen/electronics"
	"github.com/rveen/golib/csv"
	"strconv"
	"strings"
	"sort"
)

type Component struct {
	Name        string  // Denominator or reference
	Value       float64 // Value in SI unit
	Tolerance   float64
	Type        string // Part number
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

func (bom *Bom) Sort(field string) {
	sort.Sort(bom)
}

func (bom *Bom) Len() int { return len(bom.Components) }

func (bom *Bom) Swap(i, j int) { bom.Components[i], bom.Components[j] = bom.Components[j], bom.Components[i] }

func (bom *Bom) Less(i, j int) bool {

	a1 := expandNumber(bom.Components[i].Name, 6)
	a2 := expandNumber(bom.Components[j].Name, 6)

	if a1 < a2 {
		return true
	}
	return false
}

// The component number is expanded from C1 to C00001, for example.
// This is to allow sorting.
func expandNumber(s string, n int) string {

	a := []byte(s)
	var prefix, number []byte

	// Prefix
	for i, c := range a {
		if c >= '0' && c <= '9' {
			a = a[i:]
			break
		}
		prefix = append(prefix, c)
	}

	// Number
	j := 0
	for _, c := range a {
		if c < '0' || c > '9' {
			break
		}
		number = append(number, c)
		j++
	}
	a = a[j:]

	n = n - len(number)
	for i := 0; i < n; i++ {
		number = append([]byte{'0'}, number...)
	}

	// Rest
	prefix = append(prefix, number...)
	prefix = append(prefix, a...)

	return string(prefix)
}

func (bom *Bom) FromCsvs(files []string) error {

	m := csv.ReadTyped(files)

	for key, r := range m {

		c := &Component{Name: key}
		bom.Components = append(bom.Components, c)

		if val, ok := r["class"]; ok {
			c.Class = strings.ToUpper(val)
		}
		if val, ok := r["type"]; ok {
			c.Type = val
		}
		if val, ok := r["block"]; ok {
			c.Block = val
		}
		if val, ok := r["tags"]; ok {
			c.Tags = append(c.Tags, strings.Fields(val)...)
		}
		if val, ok := r["value"]; ok {
			c.Value = e.Value(val)
		}
		if val, ok := r["tolerance"]; ok {
			c.Tolerance = e.Value(val)
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

	bom.Sort("")

	return nil
}

func (c *Component) ToCsv() string {

	s := "name, class, tags, code, value, tolerance\n"

	s += fmt.Sprintf("%s, ", c.Name)
	s += fmt.Sprintf("%s, ", c.Class)
	s += fmt.Sprintf("'%s', ", c.Tags)
	s += fmt.Sprintf("'%s', ", c.Type)
	s += fmt.Sprintf("%.12f, ", c.Value)
	s += fmt.Sprintf("%f, ", c.Tolerance)

	return s + "\n"
}
