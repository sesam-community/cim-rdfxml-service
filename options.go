package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Options for the microservice
type Options map[string]interface{}

type serverOptions struct {
	log       io.Writer
	level     int
	seed      uuid.UUID
	namespace string
	options   *Options
}

// NewOptions returns default microservice options
func NewOptions(opt *Options) serverOptions {
	var seed uuid.UUID = uuid.Nil
	var log io.Writer = os.Stdout
	level := ""
	num := logERROR
	namespace := strings.Trim(os.Getenv("UUID_SEED"), " ")
	if len(namespace) == 0 {
		if opt != nil {
			// TODO: unit-tests for all of the below
			if val, exist := (*opt)["seed"]; exist && len(strings.Trim(val.(string), " ")) != 0 {
				namespace = fmt.Sprintf("%v", val)
				seed = uuid.NewSHA1(uuid.Nil, []byte(namespace))
			} else if val, exist := (*opt)["SEED"]; exist && len(strings.Trim(val.(string), " ")) != 0 {
				namespace = fmt.Sprintf("%v", val)
				seed = uuid.NewSHA1(uuid.Nil, []byte(namespace))
			} else if val, exist := (*opt)["uuid"]; exist && len(strings.Trim(val.(string), " ")) != 0 {
				seed = val.(uuid.UUID)
			} else if val, exist := (*opt)["UUID"]; exist && len(strings.Trim(val.(string), " ")) != 0 {
				seed = val.(uuid.UUID)
			}
		}
	} else {
		seed = uuid.NewSHA1(uuid.Nil, []byte(namespace))
	}
	if seed == uuid.Nil {
		fmt.Fprintf(os.Stderr, "%s\n", "fatal: missing environment 'UUID_SEED' or option 'seed' or 'uuid' for microservice.")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}

	if opt != nil {
		if val, exist := (*opt)["level"]; exist {
			level = fmt.Sprintf("%v", val)
		}
		if val, exist := (*opt)["log"]; exist {
			log = val.(io.Writer)
		}
	}
	val := os.Getenv("LOG_LEVEL")
	if len(val) != 0 {
		level = val
	}
	if len(level) != 0 {
		level = strings.ToUpper(level)
		var k int
		for k, val = range logLevel {
			if val == level {
				num = k
				break
			}
		}
	}
	return serverOptions{log: log, level: num, seed: seed, namespace: namespace, options: opt}
}

var logLevel = []string{"OFF", "CUSTOM", "QUIET", "LIVE", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE", "ALL"}

const (
	logOFF = iota
	logCUSTOM
	logQUIET
	logLIVE
	logFATAL
	logERROR
	logWARN
	logINFO
	logDEBUG
	logTRACE
	logALL
)

// Log to configured output with an INFO level
func (s *Server) Log(l string) {
	if strings.HasSuffix(l, "\n") {
		s.Logf(logINFO, l)
	} else {
		s.Logf(logINFO, "%s\n", l)
	}
}

// Logf to configured output with given level, format and parameters
func (s *Server) Logf(level int, format string, args ...interface{}) {
	if s.options.level >= level {
		fmt.Fprintf(s.options.log, format, args...)
	}
}

// Error logs to configured output with an ERROR level
func (s *Server) Error(l string) {
	if strings.HasSuffix(l, "\n") {
		s.Logf(logERROR, l)
	} else {
		s.Logf(logERROR, "%s\n", l)
	}
}

// Errorf logs to configured output with ERROR level, format and parameters
func (s *Server) Errorf(format string, args ...interface{}) {
	s.Logf(logERROR, format, args...)
}
