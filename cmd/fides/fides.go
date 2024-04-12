package main

import (
	"flag"
	"fmt"
	"log"
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
	fmt.Print(mission.ToCsv())

	// The result
	var err error
	for _, c := range bom.Components {
		c.FIT, err = fides.FIT(c, mission)
		if err != nil {
			log.Printf("%s: %s\n", c.Name, err.Error())
		}
	}

	for _, c := range bom.Components {
		fmt.Print(c.ToCsv())
		fmt.Printf("- FIT: %s: %f\n", c.Name, c.FIT)
	}
}
