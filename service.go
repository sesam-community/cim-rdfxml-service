package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

func altmain() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func main() {
	names := map[string]string{

		"_":      "https://foo.bar/",
		"cim":    "http://iec.ch/TC57/2017/CIM-schema-cim100#",
		"cim15":  "http://iec.ch/TC57/2010/CIM-schema-cim15#",
		"cim16":  "http://iec.ch/TC57/2013/CIM-schema-cim16#",
		"cim17":  "http://iec.ch/TC57/2016/CIM-schema-cim17#",
		"dm":     "http://iec.ch/TC57/61970-552/DifferenceModel/1#",
		"entsoe": "http://entsoe.eu/CIM/SchemaExtension/3/2#",
		"iev":    "http://iec.ch/TC1/60050-6xx/Electropedia/1#",
		"md":     "http://iec.ch/TC57/61970-552/ModelDescription/1#",
		"nek":    "http://nek.no/NK57/CIM/CIM100-Extension/1/0#",
		"rdf":    "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"rdfs":   "http://www.w3.org/2000/01/rdf-schema#",
		"xsd":    "http://www.w3.org/2001/XMLSchema#",
	}
	var defaults Options = Options{"json": "cim:Model.all", "ns": "names", "names": names}

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	rw := bufio.NewReadWriter(r, w)
	err := Convert(rw, &defaults, 3*1024*1024)
	rw.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	return http.ListenAndServe(":5000", NewServer(NewOptions(nil)))
}

// Server is a simple microservice
type Server struct {
	router  *httprouter.Router
	options *serverOptions
}

// NewServer sets up and returns microservice Server
func NewServer(opt serverOptions) *Server {
	s := &Server{router: httprouter.New(), options: &opt}
	s.Routes()
	s.Logf(logLIVE, "Started RFC4122 urn:uuid-scheme UUID-v5 microservice with namespace:  %s  (\"%s\").\n", s.options.seed.String(), s.options.namespace)
	var period string
	if time.Now().Year() > 2019 {
		period = fmt.Sprintf("%d-%d", 2019, time.Now().Year())
	} else {
		period = "2019"
	}
	s.Logf(logLIVE, "Copyright Sesam.io %s. All rights reserved.\n", period)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
