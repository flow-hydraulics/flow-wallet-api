package test

import (
	"io"
	"log"
)

var testLogger = log.New(io.Discard, "", log.LstdFlags)

func Logger() *log.Logger {
	return testLogger
}
