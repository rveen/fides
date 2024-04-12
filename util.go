package fides

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/rveen/golib/document"
	"github.com/rveen/ogdl"
)

// TODO read OGDL
func LoadWork(file string) *ogdl.Graph {

	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	doc, _ := document.New(string(b))
	return doc.Data()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsTag(tags []string, tag string) bool {

	for _, field := range tags {
		if field == tag {
			return true
		}
	}
	return false
}

// Read a CVS file into and array of maps
func csvRead(file string) ([]map[string]string, error) {

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	m, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	// The first line contains the field names or keys
	keys := m[0]
	for j := 0; j < len(keys); j++ {
		// Clean up (remove space and convert to lower case)
		keys[j] = strings.ToLower(strings.TrimSpace(keys[j]))
	}
	var rr []map[string]string

	for i := 1; i < len(m); i++ {

		l := m[i]
		r := make(map[string]string)

		for j := 0; j < len(l); j++ {
			// Clean up (remove space and convert to lower case)
			value := strings.ToLower(strings.TrimSpace(l[j]))
			// Add to map
			r[keys[j]] = value
		}
		rr = append(rr, r)
	}

	return rr, nil
}
