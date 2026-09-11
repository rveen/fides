package fides

import (
	"encoding/csv"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func readCases(t *testing.T, path string) []map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]string
	for _, rec := range records[1:] {
		row := map[string]string{}
		for i, k := range records[0] {
			row[strings.TrimSpace(k)] = strings.TrimSpace(rec[i])
		}
		rows = append(rows, row)
	}
	return rows
}

// num parses a number of a case, or returns def if the field is empty.
func num(t *testing.T, s string, def float64) float64 {
	t.Helper()
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// TestGuideCases runs the cases of testdata/validation/cases.csv through the
// library and compares them with the FIDES 2022 reference values that
// testdata/validation/fides_cases.py computes from the guide
// (reference-cases.csv). The library has to choose the model the case names
// from the class, tags and package.
func TestGuideCases(t *testing.T) {

	want := map[string]float64{}
	for _, r := range readCases(t, "testdata/validation/reference-cases.csv") {
		want[r["name"]] = num(t, r["fit"], math.NaN())
	}

	missions := map[string]*Mission{}
	cases := readCases(t, "testdata/validation/cases.csv")
	for _, r := range cases {
		name := r["name"]

		m := missions[r["mission"]]
		if m == nil {
			m = NewMission()
			if err := m.FromCsv("testdata/" + r["mission"]); err != nil {
				t.Fatal(err)
			}
			missions[r["mission"]] = m
		}

		c := NewComponent(name)
		c.Class = r["class"]
		c.Tags = strings.Fields(r["tags"])
		c.Package = strings.ToUpper(r["package"])
		c.Value = num(t, r["value"], 0)
		c.N = int(num(t, r["n"], 0))
		c.Np = int(num(t, r["np"], 0))
		c.V = num(t, r["v"], math.NaN())
		c.I = num(t, r["i"], math.NaN())
		c.P = num(t, r["p"], math.NaN())
		c.T = num(t, r["t"], 0)
		c.Vmax = num(t, r["vmax"], 0)
		c.Imax = num(t, r["imax"], 0)
		c.Pmax = num(t, r["pmax"], 0)
		c.Tmax = num(t, r["tmax"], 0)
		c.Layers = int(num(t, r["layers"], 0))
		c.Mounts = int(num(t, r["mounts"], 0))
		c.PiClass = num(t, r["piclass"], 0)
		c.PiTechno = num(t, r["pitechno"], 0)

		got, err := FIT(c, m)
		if err != nil {
			t.Errorf("%s (%s): %v", name, r["model"], err)
			continue
		}
		w, ok := want[name]
		if !ok {
			t.Errorf("%s: no reference value", name)
			continue
		}
		if math.Abs(got-w) > 1e-9*w {
			t.Errorf("%s (%s): library %.9g FIT, guide %.9g FIT (%+.4f %%)", name, r["model"], got, w, 100*(got/w-1))
		}
	}
	if len(cases) < 99 {
		t.Errorf("%d cases, want the 99 of cases.csv", len(cases))
	}
}
