package fides

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/golden.txt with the current results")

const golden = "testdata/golden.txt"

// TestRegression computes the FIT of every component in testdata/bom.csv and
// compares it with testdata/golden.txt.
//
// The golden file was first recorded with the library as it was before the
// fitcalc M1.0 changes. A change that is meant to alter results updates the
// golden file (go test -run Regression -update) in the same commit, so the
// diff shows exactly which components it affects.
func TestRegression(t *testing.T) {

	bom := &Bom{}
	if err := bom.FromCsvs([]string{"testdata/bom.csv", "testdata/db.csv"}); err != nil {
		t.Fatal(err)
	}
	mission := &Mission{}
	if err := mission.FromCsv("testdata/mission.csv"); err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, c := range bom.Components {
		fit, err := FIT(c, mission)
		if err != nil {
			got[c.Name] = "error: " + err.Error()
		} else {
			got[c.Name] = strconv.FormatFloat(fit, 'g', 10, 64)
		}
	}

	if *update {
		writeGolden(t, got)
		return
	}

	want := readGolden(t)
	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("%s: missing from results", name)
			continue
		}
		if !sameResult(g, w) {
			t.Errorf("%s: got %s, want %s", name, g, w)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("%s: not in %s", name, golden)
		}
	}
}

// sameResult compares FIT values with a relative tolerance, and errors
// literally.
func sameResult(a, b string) bool {
	fa, erra := strconv.ParseFloat(a, 64)
	fb, errb := strconv.ParseFloat(b, 64)
	if erra != nil || errb != nil {
		return a == b
	}
	return math.Abs(fa-fb) <= 1e-9*math.Max(math.Abs(fa), math.Abs(fb))
}

func writeGolden(t *testing.T, got map[string]string) {
	var names []string
	for name := range got {
		names = append(names, name)
	}
	sort.Strings(names)

	var sb strings.Builder
	for _, name := range names {
		fmt.Fprintf(&sb, "%s\t%s\n", name, got[name])
	}
	if err := os.WriteFile(golden, []byte(sb.String()), 0644); err != nil {
		t.Fatal(err)
	}
}

func readGolden(t *testing.T) map[string]string {
	f, err := os.Open(golden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	defer f.Close()

	want := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		name, result, ok := strings.Cut(sc.Text(), "\t")
		if ok {
			want[name] = result
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return want
}
