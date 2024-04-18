package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rveen/fides"
)

func main() {

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage: fides [options] <bom.csv> [db.csv] [work.csv] <mission.csv>")
		os.Exit(1)
	}

	// The BOM, db and working conditions files
	bom := &fides.Bom{}
	var n int
	for n = 0; n < flag.NArg()-1; n++ {
		err := bom.FromCsv(flag.Arg(n))
		if err != nil {
			fmt.Println(err.Error() + ": " + flag.Arg(n))
			os.Exit(1)
		}
	}

	// The mission
	mission := &fides.Mission{}
	mission.FromCsv(flag.Arg(n))
	// fmt.Print(mission.ToCsv())

	// The result
	var err error
	var fit float64

	fmt.Println("name, fit, class, tags, package, npins\n")
	for _, c := range bom.Components {

		c.FIT, err = fides.FIT(c, mission)

		sfit := ""

		if err != nil {
			sfit = err.Error()
		} else {
			sfit = fmt.Sprintf("%.4f", c.FIT)
			fit += c.FIT
		}
		fmt.Printf("%s, %s, %s, %s, %s, %d\n", c.Name, sfit, c.Class, c.Tags, c.Package, c.Np)
	}

	fmt.Printf("TOTAL, %f, , , ,\n", fit)
}
