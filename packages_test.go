package fides

import (
	"math"
	"testing"
)

// TestRtha: the default junction-to-ambient thermal resistance of IC
// packages comes from their pin count (FIDES 2022, p. 119), that of discrete
// packages from the table (p. 120), for their equivalents too.
func TestRtha(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) <= 1e-9*math.Abs(b) }

	for _, tc := range []struct {
		name string
		tc   float64 // substrate thermal conductivity, W/(m·K)
		want float64
	}{
		{"SOIC8", 0, 400 * math.Pow(8, -0.58) * 1.15},
		{"SOIC8", 20, 400 * math.Pow(8, -0.58) * 0.94},
		{"TSSOP16", 0, 650 * math.Pow(16, -0.58) * 1.15},
		{"PDIP14", 0, 360 * math.Pow(14, -0.58) * 1.15},
		{"LQFP64", 0, 480 * math.Pow(64, -0.58) * 1.15},
		{"SOT23", 0, 443},
		{"SOT23", 20, 360},
		{"TO236AB", 0, 443},
		{"SOT323", 0, 516},
		{"SC70", 0, 516},
		{"TO220", 0, 58},
	} {
		if got := NewPackage(tc.name).Rtha(tc.tc); !near(got, tc.want) {
			t.Errorf("%s on %g W/(m·K): Rtha %g K/W, want %g", tc.name, tc.tc, got, tc.want)
		}
	}

	// No default: an IC package without pins, QFN (whose formula needs the
	// package area), a discrete package without a table value
	for _, name := range []string{"SOIC", "QFN32", "SO8P"} {
		if got := NewPackage(name).Rtha(0); got != -1 {
			t.Errorf("%s: Rtha %g, want -1", name, got)
		}
	}
}

func TestRthjaSemi(t *testing.T) {
	if n, rth, rjc := Rthja_semi("SC70", false); n != 3 || rth != 516 || rjc != 164 {
		t.Errorf("SC70: %d pins, %g K/W, Rjc %g; want those of SOT323", n, rth, rjc)
	}
	if n, _, _ := Rthja_semi("NOPE", false); n != -1 {
		t.Errorf("unknown package: %d pins, want -1", n)
	}
}
