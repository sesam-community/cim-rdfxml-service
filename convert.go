package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	// 	"net/http"
	"strings"

	"github.com/google/uuid"
	// 	"github.com/julienschmidt/httprouter"
)

const (
	headerXML string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
	headerRDF string = `<rdf:RDF xmlns:cim="http://iec.ch/TC57/2017/CIM-schema-cim100#" xmlns:md="http://iec.ch/TC57/61970-552/ModelDescription/1#" xmlns:nek="http://nek.no/NK57/CIM/CIM100-Extension/1/0#" xmlns:entsoe="http://entsoe.eu/CIM/SchemaExtension/3/2#" xmlns:iev="http://iec.ch/TC1/60050-6xx/Electropedia/1#" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">`
	footerRDF string = `</rdf:RDF>`
	lenURN    int    = 45 // length of string "urn:uuid:00000000-0000-0000-0000-000000000000"
	posUUID   int    = 9  // length of string "urn:uuid:"
)

var (
	skipKeys map[string]bool = map[string]bool{"rdf:type": true}
)

// Convert transforms JSON to CIM RDF/XML
// func Convert(dec *json.Decoder, w *bufio.Writer, cfg *Options, sz int) error {
func Convert(rw *bufio.ReadWriter, config *Options, sz int) error {

	szDefault := 3 * 1024 * 1024 // 3MB
	if sz < szDefault {
		sz = szDefault
	}
	result := bytes.NewBuffer(make([]byte, sz)) // this will actually be the model XML-string; needs a Reset() because it is filled with 0x00 bytes
	result.Reset()
	defer func() {
		result.Reset() // empty after use as a precaution
	}()

	batch := json.NewDecoder(*rw)
	var err error
	t, err := batch.Token() // read opening bracket '['
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	if t.(json.Delim) != '[' {
		return fmt.Errorf("expected JSON array opening bracket '[', but found '%s'", t)
	}

	if _, err = rw.WriteRune('['); err != nil { // for the outer batch
		return fmt.Errorf("error writing response: %s", err)
	}

	cfg := *config
	if jsonField, exist := cfg["json"]; exist {

		// nswarn := false
		total := 0
		for batch.More() {
			var model map[string]json.RawMessage
			if err := batch.Decode(&model); err != nil {
				if strings.Contains(err.Error(), "map[string]interface") {
					return fmt.Errorf("expected JSON object inside array, but got error instead")
				}
				return fmt.Errorf("expected JSON object inside array, but got error: %s", err)
			}

			// tmp, _ := json.Marshal(model)
			// fmt.Printf("model: %v\n", string(tmp))

			xField := "xml"
			if val, exist := cfg["xml"]; exist {
				xField = fmt.Sprintf("%v", val)
			}
			// fmt.Printf("xField: %s\n", xField)

			nsField := "ns"
			if val, exist := cfg["ns"]; exist {
				nsField = fmt.Sprintf("%v", val)
			}

			var ns map[string]string
			nField := fmt.Sprintf("%v", nsField)
			if val, exist := model[nField]; exist {
				if err = json.Unmarshal(val, &ns); err != nil {
					return fmt.Errorf("expected the map of namespaces '%s' to be a JSON object with string values, but got error: %s", nField, err)
				}
			} else if val, exist := cfg[nField]; exist {

				switch nsValue := val.(type) {
				case map[string]string:
					ns = nsValue
				default:
				}

			}

			strictModel := make(map[string]interface{}, len(model))
			jField := fmt.Sprintf("%v", jsonField)
			if val, exist := model[jField]; exist {

				//

				// result.WriteString(fmt.Sprintf("\"%s\":", xField))

				dec := json.NewDecoder(bytes.NewReader(val))
				t, err := dec.Token() // read opening bracket '['
				if err != nil {
					return fmt.Errorf("%s", err)
				}
				if t.(json.Delim) != '[' {
					return fmt.Errorf("expected JSON array opening bracket '[', but found '%s'", t)
				}

				//
				//
				//

				xCount := 0
				for dec.More() {
					var entity map[string]json.RawMessage
					if err := dec.Decode(&entity); err != nil {
						if strings.Contains(err.Error(), "map[string]json.RawMessage") {
							return fmt.Errorf("expected JSON object inside array, but got error instead")
						}
						return fmt.Errorf("expected JSON object inside array, but got error: %s", err)
					}
					if len(entity) == 0 {
						continue
					}

					name, class, id, err := identity(&entity)
					if err != nil {
						// return err
						fmt.Fprintf(os.Stderr, "%s\n", err)
						continue // skipping bad errors
					}

					if xCount == 0 {
						// result.WriteRune('"')
						result.WriteString(headerXML)
						result.WriteRune('\n')
						result.WriteString(headerRDF)
						result.WriteRune('\n')
					}
					result.WriteString(fmt.Sprintf("  <%s:%s rdf:about=\"_%s\">\n", name, class, id[posUUID:]))

					local := bytes.NewBufferString("")
					// var data []byte
					// strictEntity := make(map[string]json.RawMessage, len(entity))
					for k, v := range entity {
						if skip, exists := skipKeys[k]; exists {
							if skip {
								continue
							}
						}
						if parts := strings.Split(k, ":"); len(parts) == 2 {
							prefix := parts[0]
							attr := parts[1]
							// FXIME: use val for rdf:resource etc... i.e need to expand/substitute in v
							// if val, exists := ns[prefix]; exists {
							if _, exists := ns[prefix]; exists {

								var value interface{}
								if err = json.Unmarshal(v, &value); err != nil {

									fmt.Printf("-----ERROR-----  '%s' for: %v\n", err, v)

								}

								switch attrValue := value.(type) {
								case nil:
									continue
								case string:
									pieces := strings.Split(attrValue, ":")
									if len(pieces) == 3 && pieces[0] == "~" {
										localRef := pieces[2]
										localNS := pieces[1]
										if len(strings.Split(localRef, "-")) == 5 {
											result.WriteString(fmt.Sprintf("    <%s:%s rdf:resource=\"#_%s\"/>\n", prefix, attr, localRef))
										} else if ref, exists := ns[localNS]; exists {
											result.WriteString(fmt.Sprintf("    <%s:%s rdf:resource=\"%s%s\"/>\n", prefix, attr, ref, localRef))
										} else {
											result.WriteString(fmt.Sprintf("    <%s:%s>%v</%s:%s>\n", prefix, attr, value, prefix, attr))
										}
									} else {
										result.WriteString(fmt.Sprintf("    <%s:%s>%s</%s:%s>\n", prefix, attr, attrValue, prefix, attr))
									}
								case map[string]interface{}:
									localID := uuid.NewSHA1(uuid.Nil, []byte(fmt.Sprintf("%s:%s:%s", id, prefix, attr))).String()
									result.WriteString(fmt.Sprintf("    <%s:%s rdf:resource=\"#_%s\"/>\n", prefix, attr, localID))
									localCount := 0
									subName := name
									subKey := ""
									for localKey, localVal := range attrValue {
										subName = name
										subKey = localKey
										nameSubs := strings.Split(localKey, ":")
										if len(nameSubs) == 2 {
											subName = nameSubs[0]
											subKey = nameSubs[1]
										}
										if localCount == 0 {
											local.WriteString(fmt.Sprintf("    <%s:%s rdf:about=\"_%s\">\n", subName, subKey, localID))
										}
										subs := strings.Split(localKey, ".")
										attrSubs := strings.Split(attr, ".")
										if len(subs) == 2 && len(attrSubs) == 2 {
											if subs[0] == attrSubs[1] {
												result.WriteString(fmt.Sprintf("        <%s:%s>%v</%s:%s>\n", subName, attrSubs, localVal, subName, attrSubs))
											}
										}
										localCount++
									}
									if localCount != 0 {
										local.WriteString(fmt.Sprintf("    </%s:%s>\n", subName, subKey))
									}
								case []interface{}:
									continue
								default:
									result.WriteString(fmt.Sprintf("    <%s:%s>%v</%s:%s>\n", prefix, attr, attrValue, prefix, attr))
								}

							}
						}
					}
					// TODO: make another testing-only flag here to make strictEntity not possible to marshal, for testing HTTP 503 below
					// if data, err = json.Marshal(strictEntity); err != nil {
					// 	return fmt.Errorf("%s", err)
					// }

					// fmt.Printf("data: %v\n", string(data))

					// if _, err = result.Write(data); err != nil { // for the outer batch
					// 	return fmt.Errorf("error writing response: %s", err)
					// }

					result.WriteString(fmt.Sprintf("  </%s:%s>\n", name, class))
					if local.Len() == 0 {
						result.WriteString(local.String())
					}

					//
					//
					//

					xCount++
				}

				if xCount != 0 {
					result.WriteString(footerRDF)
				}

				if _, err = dec.Token(); err != nil { // read closing bracket ']'
					return fmt.Errorf("expected JSON array closing bracket ']', but got error: %s", err)
				}
				// result.WriteRune('"')

				delete(model, jField)
				strictModel[xField] = result.String()

				for k, v := range model {
					var any interface{}
					json.Unmarshal(v, &any)
					strictModel[k] = any
				}

				//

			}

			// TODO: make a testing-only flag here to make model not possible to marshal, for testing HTTP 503 below
			var data []byte
			// if data, err = json.Marshal(model); err != nil {
			// 	return fmt.Errorf("%s", err)
			// }
			if total != 0 {
				if _, err = rw.WriteRune(','); err != nil { // for the outer batch
					return fmt.Errorf("error writing response: %s", err)
				}
			}
			for k := range strictModel {
				if k != "_id" && k[0] == '_' {

					delete(strictModel, k)

				}
			}
			// TODO: make another testing-only flag here to make strictModel not possible to marshal, for testing HTTP 503 below
			if data, err = json.Marshal(strictModel); err != nil {
				return fmt.Errorf("%s", err)
			}

			// fmt.Printf("data: %v\n", string(data))

			if _, err = rw.Write(data); err != nil { // for the outer batch
				return fmt.Errorf("error writing response: %s", err)
			}
			rw.Flush()
			total++
		}
	}
	rw.Flush()

	if _, err = batch.Token(); err != nil { // read closing bracket ']'
		return fmt.Errorf("expected JSON array closing bracket ']', but got error: %s", err)
	}

	// TODO: test-case with a failing rw-ReadWriter (simulating client peer closed connection etc) returning error for testing HTTP 503 below
	if _, err = rw.WriteRune(']'); err != nil { // for the outer batch
		return fmt.Errorf("error writing response: %s", err)
	}
	return nil
}

func identity(entity *map[string]json.RawMessage) (name string, class string, id string, err error) {
	var ids []string
	if val, exist := (*entity)["$ids"]; exist {
		if err = json.Unmarshal(val, &ids); err != nil {
			err = fmt.Errorf("expected '$ids' to be a JSON array of string values, but got error: %s", err)
			return name, class, id, err
		}
	}
	var names []string
	if val, exist := (*entity)["rdf:type"]; exist {
		if err = json.Unmarshal(val, &names); err != nil {
			if err = json.Unmarshal(val, &name); err != nil {
				err = fmt.Errorf("expected 'rdf:type' to be a JSON string value or JSON array of string values, but got error: %s", err)
				return name, class, id, err
			}
			names = []string{name}
		}
	}
	if val, exist := (*entity)["_id"]; exist {
		if err = json.Unmarshal(val, &id); err != nil {
			err = fmt.Errorf("expected '_id' to be a JSON string value, but got error: %s", err)
			return name, class, id, err
		}
	}
	if !strings.HasPrefix(id, "urn:uuid:") || len(id) != lenURN || len(strings.Split(id[posUUID:], "-")) != 5 {
		err = fmt.Errorf("expected '_id' to be a valid RFC 4122 urn:uuid-scheme value")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			// fmt.Printf("\n\n\n")
			e := *entity
			for k, v := range e {
				var in interface{}
				json.Unmarshal(v, &in)
				fmt.Fprintf(os.Stderr, "\"%s\": %v\n", k, string(v))
				// fmt.Printf("\"%s\": %v\n", k, string(v))
			}
			// fmt.Printf("\n\n\n")
		}
		return name, class, id, err
	}
	var (
		hasID      = false
		hasClass   = false
		hasRDFTYPE = false
	)
	for _, val := range ids {
		if val == id {
			hasID = true
		}
		if strings.HasSuffix(val, id[posUUID:]) {
			parts := strings.Split(val, ":")
			if hasClass = len(parts) == 3 && parts[0] == "~" && parts[2] == id[posUUID:]; hasClass {
				class = parts[1]
			}
		}
	}
	if !hasID {
		err = fmt.Errorf("expected '$ids' to contain the '_id' urn:uuid-scheme")
	}
	if !hasClass {
		err = fmt.Errorf("expected '$ids' to contain the CIM class of '_id' as a (NI) namespace identifier of format '~:cim-class:UUID'")
	}
	for _, val := range names {
		if strings.HasSuffix(val, ":"+class) {
			parts := strings.Split(val, ":")
			if hasRDFTYPE = len(parts) == 3 && parts[0] == "~" && parts[2] == class; hasRDFTYPE {
				name = parts[1]
			}
		}
	}
	if !hasRDFTYPE {
		err = fmt.Errorf("expected 'rdf:type' to contain the class (NI) namespace identifier '~:<namespace>:%s'", class)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		// fmt.Printf("\n\n\n")
		e := *entity
		for k, v := range e {
			var in interface{}
			json.Unmarshal(v, &in)
			fmt.Fprintf(os.Stderr, "\"%s\": %v\n", k, string(v))
			// fmt.Printf("\"%s\": %v\n", k, string(v))
		}
		// fmt.Printf("\n\n\n")
	}
	return name, class, id, err
}
