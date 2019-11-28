package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// HandleDefault receives URL POST requests without field or namespace components,
// but reroutes with a default field and namespace
func (s *Server) HandleDefault(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.HandleFieldNamespace(w, r, []httprouter.Param{{Key: "field", Value: "_id"}, {Key: "namespace", Value: "rdf:type"}})
}

// HandleField receives URL POST requests without namespace component,
// but reroutes with a default namespace
func (s *Server) HandleField(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	p = append(p, httprouter.Param{Key: "namespace", Value: "rdf:type"})
	s.HandleFieldNamespace(w, r, p)
}

// HandleFieldNamespace receives URL POST requests with JSON body consisting of array of objects,
// and returns requested field values transformed to UUID based on SHA1 of field values
// https://github.com/google/uuid
// https://tools.ietf.org/html/rfc4122   (URN:UUID-scheme)
// https://en.wikipedia.org/wiki/Uniform_Resource_Name
// https://en.wikipedia.org/wiki/Universally_unique_identifier
func (s *Server) HandleFieldNamespace(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	if r.ContentLength == 0 {
		s.Errorf("error: missing JSON array of entities\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sz := 3 * 1024 * 1024 // 3MB
	if int(r.ContentLength) > sz {
		sz = int(r.ContentLength)
	}
	result := bytes.NewBuffer(make([]byte, sz))
	result.Reset()
	defer func() {
		result.Reset()
	}()

	var err error
	dec := json.NewDecoder(r.Body)
	t, err := dec.Token() // read opening bracket '['
	if err != nil {
		s.Errorf("%s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if t.(json.Delim) != '[' {
		s.Errorf("expected JSON array opening bracket '[', but found '%s'\n", t)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result.WriteRune('[')

	nswarn := false
	total := 0
	for dec.More() {
		var entity map[string]interface{}
		if err := dec.Decode(&entity); err != nil {
			if strings.Contains(err.Error(), "map[string]interface") {
				s.Errorf("expected JSON object inside array, but got error instead\n")
			} else {
				s.Errorf("expected JSON object inside array, but got error: %s\n", err)
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		keyspecs := strings.Split(p.ByName("field"), ";")
		for _, keyspec := range keyspecs {
			// key := p.ByName("field")
			key := keyspec // key variable mutates (is substituted), so keeping the original specification as well
			prefix := ""
			if key[0] == '_' && key != "_id" {
				prefix = "#_" // key given wanting automatic RDF resource local label reference format
				key = key[1:]
			}

			ns := p.ByName("namespace")
			if key[0] == ':' && ns == "" {
				// FIXME: just make some special meaning for <nil> namespace ? (when disabled HTTP 307 redirects for trailing slash in router)
				ns = "rdf:type"
			}
			if _, exist := entity[key[1:]]; exist && key[0] == ':' {
				key = key[1:]
			} else if val, exist := entity[key]; !exist {
				// key shortcut given needing expanding
				var nskey string
				if key[0] == ':' {
					if key[1] == '.' {
						nskey = key[1:]
					} else {
						nskey = key
					}
				} else {
					ns = "" // no automatic namespace
					if key[0] == '.' {
						nskey = key
					} else {
						nskey = ":" + key
					}
				}
				for k := range entity {
					if strings.HasSuffix(k, nskey) {
						key = k // k includes pipeline namespace, key is now expanded from shortcut
						break
					}
				}
			} else {
				if strings.Contains(fmt.Sprintf("%v", val), ":") {
					ns = "" // key value already includes desired namespace
				}
			}

			if ns == "rdf:type" {
				// want automatic namespacing
				if val, exist := entity[ns]; exist {
					switch value := val.(type) {
					case []interface{}:
						many := val.([]interface{})
						if len(many) == 0 {
							ns = "" // empty array
						} else if len(many) == 1 {
							ns = fmt.Sprintf("%v", many[0])
						} else {
							ns = fmt.Sprintf("%v", many[0])
							if !nswarn {
								s.Logf(logWARN, "warning '%s', multiple 'rdf:type' (using '%v', please indicate): %v\n", keyspec, many[0], many)
								nswarn = true
							}
						}
					default:
						ns = fmt.Sprintf("%v", value)
					}
				} else {
					if !nswarn {
						s.Logf(logWARN, "warning '%s', no 'rdf:type' found\n", keyspec)
						nswarn = true
					}
					ns = "" // no RDF type information, so setting blank namespace
				}
				if !nswarn && ns == "" {
					s.Logf(logWARN, "warning '%s', empty 'rdf:type'\n", keyspec)
					nswarn = true
				}
			} else if strings.HasSuffix(ns, ":") {
				if !strings.HasPrefix(ns, "~:") {
					ns = "~:" + ns
				}
				if val, exist := entity["rdf:type"]; exist {
					choice := ""
					switch value := val.(type) {
					case []interface{}:
						n := 0
						for _, v := range value {
							rdfType := v.(string)
							if strings.HasPrefix(rdfType, ns) {
								if n == 0 { // choose first prefix match
									choice = rdfType
								}
								n++ // count matches for possible warning
							}
						}
						if choice == "" {
							if !nswarn {
								s.Logf(logWARN, "warning '%s', prefix '%s' not in 'rdf:type'\n", keyspec, ns)
								nswarn = true
							}
						} else if n != 1 {
							if !nswarn {
								s.Logf(logWARN, "warning '%s', multiple 'rdf:type' (using '%v', please indicate): %v\n", keyspec, ns, value)
								nswarn = true
							}
						} else {
							ns = "" // empty array
						}
						ns = choice
					default:
						choice = fmt.Sprintf("%v", value)
						if strings.HasPrefix(choice, ns) {
							ns = choice
						} else {
							if !nswarn {
								s.Logf(logWARN, "warning '%s', prefix '%s' doesn't match 'rdf:type' %v\n", keyspec, ns, value)
								nswarn = true
							}
							ns = ""
						}
					}
				} else {
					if !nswarn {
						s.Logf(logWARN, "warning '%s', no 'rdf:type' found\n", keyspec)
						nswarn = true
					}
					ns = "" // no RDF type information, so setting blank namespace
				}
			} else {
				// given a complete namespace
			}
			if strings.HasPrefix(ns, "~:") {
				ns = ns[2:]
			}

			ns = strings.Trim(ns, " ") // forced empty if namespace-parameter was %20 (i.e ' ')
			if val, exist := entity[key]; exist {
				if len(ns) != 0 && !strings.HasSuffix(ns, ":") {
					ns += ":"
				}
				switch value := val.(type) {
				case []interface{}:
					many := val.([]interface{})
					shaids := make([]interface{}, len(many))
					for i, v := range many {
						shaid := uuid.NewSHA1(s.options.seed, []byte(fmt.Sprintf("%s%v", ns, v))) // format is "namespace:value" since non-empty namespace always includes ':'
						shaids[i] = fmt.Sprintf("%s%s", prefix, shaid.String())
						s.Logf(logDEBUG, "[%s]:%d '%s%v'\t  ->  %s   (%x)\n", key, i, ns, v, shaid.String(), [16]byte(shaid))
					}
					entity[key] = shaids
				default:
					shaid := uuid.NewSHA1(s.options.seed, []byte(fmt.Sprintf("%s%v", ns, value))) // format is "namespace:value" since non-empty namespace always includes ':'
					entity[key] = fmt.Sprintf("%s%s", prefix, shaid.String())
					s.Logf(logDEBUG, "[%s] '%s%v'\t  ->  %s   (%x)\n", key, ns, value, shaid.String(), [16]byte(shaid))
				}
			}

		}
		// TODO: make a testing-only flag here to make entity not possible to marshal, for testing HTTP 503 below
		var data []byte
		if data, err = json.Marshal(entity); err != nil {
			s.Errorf("%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if total != 0 {
			result.WriteRune(',')
		}
		strictEntity := make(map[string]interface{}, len(entity))
		for k, v := range entity {
			if k == "_id" || k == "" || k[0] != '_' {
				strictEntity[k] = v
			}
		}
		// TODO: make another testing-only flag here to make strictEntity not possible to marshal, for testing HTTP 503 below
		if data, err = json.Marshal(strictEntity); err != nil {
			s.Errorf("%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result.Write(data)
		total++
	}

	if _, err = dec.Token(); err != nil { // read closing bracket ']'
		s.Errorf("expected JSON array closing bracket ']', but got error: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result.WriteRune(']')

	// TODO: test-case with a failing w-ResponseWriter (simulating client peer closed connection etc) returning error for testing HTTP 503 below
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err = fmt.Fprint(w, result.String()); err != nil {
		s.Errorf("error writing response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
