package alice

// File contains global settings Alice uses

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logLevel zerolog.Level = zerolog.TraceLevel

/*
SetLogLevel sets the level of logging Alice uses
Levels are:
	-2 Off
	-1 Trace
	0 Debug
	1 Info
	2 Warning
	3 Error
	4 Fatal
	5 Panic
*/
func SetLogLevel(level int) {
	if level == -2 {
		logLevel = zerolog.Disabled
		return
	}
	logLevel = zerolog.Level(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
