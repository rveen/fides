package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rveen/fides"
)

func main() {

	var err error
	var md bool

	flag.BoolVar(&md, "md", false, "output markdown format")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage: fides [options] <bom.csv> [db.csv] [work.csv] <mission.csv>")
		os.Exit(1)
	}

	// The BOM, db and working conditions files
	bom := &fides.Bom{}
	var n int
	var files []string
	for n = 0; n < flag.NArg()-1; n++ {
		files = append(files, flag.Arg(n))
	}
	err = bom.FromCsvs(files)
	if err!=nil {
		fmt.Printf(err.Error())
		os.Exit(-1)
	}

	// The mission
	mission := &fides.Mission{}
	mission.FromCsv(flag.Arg(n))

	// The result
	
	var fit float64

	if md {
        fmt.Println("# FIDES 2022 analysis\n\n## FIT values\n")
		fmt.Println("| Name | FIT | Class | Tags | Package | Conditions |")
		fmt.Println("|---|---|---|---|---|---|")
	} else {
		fmt.Println("name, fit, class, tags, package, npins, power")
	}
	for _, c := range bom.Components {

		c.FIT, err = fides.FIT(c, mission)

		sfit := ""

		if err != nil {
			sfit = err.Error()
		} else {
			sfit = fmt.Sprintf("%.4f", c.FIT)
			fit += c.FIT
		}

		tags := ""
		for _, tag := range c.Tags {
			tags += " " + tag
		}

		cond := fmt.Sprintf("V=%f V, P=%f W",c.V,c.P)

		if md {
			fmt.Printf("| %s | %s | %s | %s | %s | %s |\n", strings.ToUpper(c.Name), sfit, c.Class, tags[1:], c.Package, cond)
		} else {
			fmt.Printf("%s, %s, %s, %s, %s, %d, %f\n", strings.ToUpper(c.Name), sfit, c.Class, tags[1:], c.Package, c.Np, c.P)
	    }
	}

	fmt.Printf("\n FIT TOTAL = %f\n\n", fit)

	if md {
		fmt.Printf("## Mission profile\n\n")
		fmt.Print(mission.ToMD())
	}
}
