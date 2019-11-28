package main_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "sesam-cimrdf"
)

func NewInputOutput(input string, output string, buf *bytes.Buffer) *bufio.ReadWriter {
	sz := len(input)
	r := bufio.NewReaderSize(bytes.NewBufferString(input), sz)
	w := bufio.NewWriterSize(buf, sz*3)
	return bufio.NewReadWriter(r, w)
}

func NewContent(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func NewOuter(output string, omit string) (outer string, err error) {
	var model []json.RawMessage
	err = json.Unmarshal([]byte(output), &model)
	if err != nil {
		return "", fmt.Errorf("not valid JSON array of objects, got error %s\n", err)
	}
	result := make([]map[string]interface{}, len(model))

	for i, o := range model {
		var object map[string]interface{}
		err = json.Unmarshal(o, &object)
		if err != nil {
			return "", fmt.Errorf("not valid JSON object, got error %s\n", err)
		}
		delete(object, omit)
		result[i] = object
	}
	var buf []byte
	buf, err = json.Marshal(result)
	outer = string(buf)
	return outer, nil
}

func NewInner(output string, keep string) (inner []string, err error) {
	var model []json.RawMessage
	err = json.Unmarshal([]byte(output), &model)
	if err != nil {
		return inner, fmt.Errorf("not valid JSON array of objects, got error %s\n", err)
	}
	result := make([]string, len(model))
	for i, o := range model {
		var object map[string]json.RawMessage
		err = json.Unmarshal(o, &object)
		if err != nil {
			return inner, fmt.Errorf("not valid JSON object, got error %s\n", err)
		}
		if val, exists := object[keep]; exists {
			var value string
			err = json.Unmarshal(val, &value)
			result[i] = value
		}
	}
	inner = result
	return inner, nil
}

const (
	namespaces string = `"ns": {
		"_": "https://foo.bar/",
		"cim": "http://iec.ch/TC57/2017/CIM-schema-cim100#",
		"cim15": "http://iec.ch/TC57/2010/CIM-schema-cim15#",
		"cim16": "http://iec.ch/TC57/2013/CIM-schema-cim16#",
		"cim17": "http://iec.ch/TC57/2016/CIM-schema-cim17#",
		"dm": "http://iec.ch/TC57/61970-552/DifferenceModel/1#",
		"md": "http://iec.ch/TC57/61970-552/ModelDescription/1#",
		"iev": "http://iec.ch/TC1/60050-6xx/Electropedia/1#",
		"entsoe": "http://entsoe.eu/CIM/SchemaExtension/3/2#",
		"nek": "http://nek.no/NK57/CIM/CIM100-Extension/1/0#",
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
		"xsd": "http://www.w3.org/2001/XMLSchema#"
	}`
	NL        string = "\n"
	headerXML string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
	headerRDF string = `<rdf:RDF xmlns:cim="http://iec.ch/TC57/2017/CIM-schema-cim100#" xmlns:md="http://iec.ch/TC57/61970-552/ModelDescription/1#" "xmlns:nek": "http://nek.no/NK57/CIM/CIM100-Extension/1/0#" "xmlns:entsoe": "http://entsoe.eu/CIM/SchemaExtension/3/2#" "xmlns:iev": "http://iec.ch/TC1/60050-6xx/Electropedia/1#" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">`
	footerRDF string = `</rdf:RDF>`
)

var _ = Describe("Microservice CIM JSON-to-RDF/XML conversions", func() {

	var (
		defaults Options = Options{"json": "json"}
		input    string
		output   string
		buf      bytes.Buffer
		rw       *bufio.ReadWriter
		sz       int
		err      error
		content  string
		outer    string
	)

	Describe("when parsing simple JSON structures", func() {

		Context("with JSON empty array", func() {
			BeforeEach(func() {
				input = `[]`
				output = `[]`
				rw = NewInputOutput(input, output, &buf)
				err = Convert(rw, &defaults, sz)
				rw.Flush()
			})
			AfterEach(func() {
				buf.Reset()
			})
			It("replies with", func() {
				By("no error")
				Expect(err).To(BeNil())
				By("correct CIM RDF/XML")
				Expect(buf.String()).To(MatchJSON(output))
			})
		})

		/*
			Context("with entity without '_id' field", func() {
				BeforeEach(func() {
					input = `[{"key":"val","fields":2}]`
					output = input
					rw = NewInputOutput(input, output, &buf)
					err = Convert(rw, &defaults, sz)
					rw.Flush()
				})
				AfterEach(func() {
					buf.Reset()
				})
				It("replies with", func() {
					By("no error")
					Expect(err).To(BeNil())
					By("correct CIM RDF/XML")
					Expect(buf.String()).To(MatchJSON(output))
				})
			})

			Context("with entity containing '_id' field", func() {
				BeforeEach(func() {
					input = `[{"_id":"urn:uuid:0001070f-175c-511f-ba20-06bc1c36b47e", "key":"val","fields":2}]`
					output = `[{"_id":"urn:uuid:0001070f-175c-511f-ba20-06bc1c36b47e", "key":"val","fields":2}]`
					rw = NewInputOutput(input, output, &buf)
					err = Convert(rw, &defaults, sz)
					rw.Flush()
				})
				AfterEach(func() {
					buf.Reset()
				})
				It("replies with", func() {
					By("no error")
					Expect(err).To(BeNil())
					By("correct CIM RDF/XML")
					Expect(buf.String()).To(MatchJSON(output))
				})
			})

		*/

	})

	Describe("when parsing simple JSON structures", func() {

		Context("with empty CIM array", func() {
			BeforeEach(func() {
				input = `[{"json":[]}]`
				output = `[{"xml":""}]`
				rw = NewInputOutput(input, output, &buf)
				err = Convert(rw, &defaults, sz)
				rw.Flush()
			})
			AfterEach(func() {
				buf.Reset()
			})
			It("delivers", func() {
				By("no error")
				Expect(err).To(BeNil())
				By("correct CIM RDF/XML")
				Expect(buf.String()).To(MatchJSON(output))
			})
		})

		Context("with empty entity in CIM array", func() {
			BeforeEach(func() {
				input = `[{"json":[
					{}
				]}]`
				output = `[{"xml":""}]`
				rw = NewInputOutput(input, output, &buf)
				err = Convert(rw, &defaults, sz)
				rw.Flush()
			})
			AfterEach(func() {
				buf.Reset()
			})
			It("delivers", func() {
				By("no error")
				Expect(err).To(BeNil())
				By("correct CIM RDF/XML")
				Expect(buf.String()).To(MatchJSON(output))
			})
		})

		Context("with a single CIM component", func() {
			BeforeEach(func() {
				input = `[{` + namespaces + `
					,"json":[
						{
							"$ids": [
								"urn:uuid:00000000-0000-0000-0000-000000000000",
								"~:cim-class:00000",
								"~:Class:00000000-0000-0000-0000-000000000000"
							],
							"_id": "urn:uuid:00000000-0000-0000-0000-000000000000",
						  "cim:Class.property": "value",
						  "rdf:type": "~:cim:Class"
						}
					]
					}]`
				content = fmt.Sprintf("%s\n%s\n", headerXML, headerRDF)
				content += `
				  <cim:Class rdf:ID="_00000000-0000-0000-0000-000000000000">
				  <cim:Class.property>value</cim:Class.property>
					</cim:Class>
					`
				content += fmt.Sprintf("%s\n", footerRDF)
				output = `[{` + namespaces + `,"xml":` + NewContent(content) + ` }]`
				outer, err = NewOuter(output, "xml")
				if err != nil {
					fmt.Printf("ERR outer: %s\n", err)
					fmt.Println(output)
				}
				rw = NewInputOutput(input, output, &buf)
				err = Convert(rw, &defaults, sz)
				rw.Flush()
			})
			AfterEach(func() {
				buf.Reset()
			})
			It("delivers", func() {
				By("no error")
				Expect(err).To(BeNil())
				By("valid JSON")
				naked, err := NewOuter(buf.String(), "xml")
				Expect(err).To(BeNil())
				Expect(naked).To(MatchJSON(outer))
				By("correct CIM XML/RDF")
				inner, err := NewInner(buf.String(), "xml")
				Expect(err).To(BeNil())
				Expect(inner[0]).To(MatchXML(content))
			})
		})

		Context("with a single CIM component having local refs", func() {
			BeforeEach(func() {
				input = `[{` + namespaces + `
					,"json":[
						{
							"$ids": [
								"urn:uuid:00000000-0000-0000-0000-000000000000",
								"~:cim-class:00000",
								"~:Class:00000000-0000-0000-0000-000000000000"
							],
							"_id": "urn:uuid:00000000-0000-0000-0000-000000000000",
						  "cim:Class.property": "value",
						  "cim:Class.ref": "~:cim:Values.item",
						  "cim:Class.Other": "~:AltClass:00000000-1100-0000-0011-000000000000",
						  "rdf:type": "~:cim:Class"
						}
					]
					}]`
				content = fmt.Sprintf("%s\n%s\n", headerXML, headerRDF)
				content += `
				  <cim:Class rdf:ID="_00000000-0000-0000-0000-000000000000">
				  <cim:Class.property>value</cim:Class.property>
				  <cim:Class.ref rdf:resource="http://iec.ch/TC57/2017/CIM-schema-cim100#Values.item"/>
				  <cim:Class.Other rdf:resource="#_00000000-1100-0000-0011-000000000000"/>
					</cim:Class>
					`
				content += fmt.Sprintf("%s\n", footerRDF)
				output = `[{` + namespaces + `,"xml":` + NewContent(content) + ` }]`
				outer, err = NewOuter(output, "xml")
				if err != nil {
					fmt.Printf("ERR outer: %s\n", err)
					fmt.Println(output)
				}
				rw = NewInputOutput(input, output, &buf)
				err = Convert(rw, &defaults, sz)
				rw.Flush()
			})
			AfterEach(func() {
				buf.Reset()
			})
			It("delivers", func() {
				By("no error")
				Expect(err).To(BeNil())
				By("valid JSON")
				naked, err := NewOuter(buf.String(), "xml")
				Expect(err).To(BeNil())
				Expect(naked).To(MatchJSON(outer))
				By("correct CIM XML/RDF")
				inner, err := NewInner(buf.String(), "xml")
				Expect(err).To(BeNil())
				Expect(inner[0]).To(MatchXML(content))
			})
		})

		Context("with several CIM components having local refs", func() {
			BeforeEach(func() {
				input = `[{` + namespaces + `
					,"json":[
						{
							"$ids": [
								"urn:uuid:00000000-0000-0000-0000-000000000000",
								"~:cim-class:00000",
								"~:Class:00000000-0000-0000-0000-000000000000"
							],
							"_id": "urn:uuid:00000000-0000-0000-0000-000000000000",
						  "cim:Class.property": "value",
						  "cim:Class.ref": "~:cim:Values.item",
						  "cim:Class.Other": "~:AltClass:00000000-1100-0000-0011-000000000000",
						  "rdf:type": "~:cim:Class"
						},
						{
							"$ids": [
								"urn:uuid:00000000-1100-0000-0011-000000000000",
								"~:cim-class:00000",
								"~:AltClass:00000000-1100-0000-0011-000000000000"
							],
							"_id": "urn:uuid:00000000-1100-0000-0011-000000000000",
						  "cim:AltClass.property": "value",
						  "cim:AltClass.ref": "~:cim:Values.more",
						  "cim:AltClass.Other": "~:Class:00000000-0000-0000-0000-000000000000",
						  "rdf:type": [ "~:cim:AltClass" , "~:cim:Ident" ]
						}
					]
					}]`
				content = fmt.Sprintf("%s\n%s\n", headerXML, headerRDF)
				content += `
				  <cim:Class rdf:ID="_00000000-0000-0000-0000-000000000000">
				  <cim:Class.property>value</cim:Class.property>
				  <cim:Class.ref rdf:resource="http://iec.ch/TC57/2017/CIM-schema-cim100#Values.item"/>
				  <cim:Class.Other rdf:resource="#_00000000-1100-0000-0011-000000000000"/>
					</cim:Class>
				  <cim:AltClass rdf:ID="_00000000-1100-0000-0011-000000000000">
				  <cim:AltClass.property>value</cim:AltClass.property>
				  <cim:AltClass.ref rdf:resource="http://iec.ch/TC57/2017/CIM-schema-cim100#Values.more"/>
				  <cim:AltClass.Other rdf:resource="#_00000000-0000-0000-0000-000000000000"/>
					</cim:AltClass>
					`
				content += fmt.Sprintf("%s\n", footerRDF)
				output = `[{` + namespaces + `,"xml":` + NewContent(content) + ` }]`
				outer, err = NewOuter(output, "xml")
				if err != nil {
					fmt.Printf("ERR outer: %s\n", err)
					fmt.Println(output)
				}
				rw = NewInputOutput(input, output, &buf)
				err = Convert(rw, &defaults, sz)
				rw.Flush()
			})
			AfterEach(func() {
				buf.Reset()
			})
			It("delivers", func() {
				By("no error")
				Expect(err).To(BeNil())
				By("valid JSON")
				naked, err := NewOuter(buf.String(), "xml")
				Expect(err).To(BeNil())
				Expect(naked).To(MatchJSON(outer))
				By("correct CIM XML/RDF")
				inner, err := NewInner(buf.String(), "xml")
				Expect(err).To(BeNil())
				Expect(inner[0]).To(MatchXML(content))
			})
		})

	})

})
