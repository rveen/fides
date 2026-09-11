package fides

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
)

// FITAll computes FIT for every component. A nil error entry means success;
// NaN in the corresponding fits entry means that component failed to compute.
func (b *Bom) FITAll(m *Mission) ([]float64, []error) {
	fits := make([]float64, len(b.Components))
	errs := make([]error, len(b.Components))
	for i, c := range b.Components {
		fits[i], errs[i] = FIT(c, m)
	}
	return fits, errs
}

// TotalFIT returns the sum of all computable FIT values and the per-component
// errors for any components that failed. Components with NaN FIT are skipped.
func (b *Bom) TotalFIT(m *Mission) (float64, []error) {
	fits, errs := b.FITAll(m)
	var total float64
	for _, f := range fits {
		if !math.IsNaN(f) {
			total += f
		}
	}
	return total, errs
}

// FITByBlock groups computed FIT values by the Block field on each Component.
// Components with NaN FIT or an empty Block are skipped.
func (b *Bom) FITByBlock(m *Mission) (map[string]float64, []error) {
	fits, errs := b.FITAll(m)
	groups := make(map[string]float64)
	for i, c := range b.Components {
		if math.IsNaN(fits[i]) || c.Block == "" {
			continue
		}
		groups[c.Block] += fits[i]
	}
	return groups, errs
}

// Report writes a tabular FIT summary to w sorted by descending FIT.
// Columns: reference, class, FIT, % of total, cumulative %.
// Components that returned an error are listed at the end with their error.
func (b *Bom) Report(m *Mission, w io.Writer) error {
	fits, errs := b.FITAll(m)

	var total float64
	for _, f := range fits {
		if !math.IsNaN(f) {
			total += f
		}
	}

	type row struct {
		name  string
		class string
		fit   float64
		pct   float64
	}

	var good []row
	var bad []struct{ name, msg string }

	for i, c := range b.Components {
		if errs[i] != nil || math.IsNaN(fits[i]) {
			msg := "error"
			if errs[i] != nil {
				msg = errs[i].Error()
			}
			bad = append(bad, struct{ name, msg string }{c.Name, msg})
			continue
		}
		pct := 0.0
		if total > 0 {
			pct = 100 * fits[i] / total
		}
		good = append(good, row{c.Name, c.Class, fits[i], pct})
	}

	sort.Slice(good, func(i, j int) bool { return good[i].fit > good[j].fit })

	const sep = "----------------------------------------------------------------------"
	fmt.Fprintf(w, "%-12s %-5s %12s %8s %13s\n", "Name", "Class", "FIT", "%", "Cumulative%")
	fmt.Fprintln(w, sep)

	var cumulative float64
	for _, r := range good {
		cumulative += r.pct
		fmt.Fprintf(w, "%-12s %-5s %12.4f %8.2f %13.2f\n",
			strings.ToUpper(r.name), r.class, r.fit, r.pct, cumulative)
	}

	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "%-18s %12.4f\n\n", "TOTAL", total)

	if len(bad) > 0 {
		fmt.Fprintln(w, "Components with errors:")
		for _, b := range bad {
			fmt.Fprintf(w, "  %-12s  %s\n", strings.ToUpper(b.name), b.msg)
		}
	}
	return nil
}
