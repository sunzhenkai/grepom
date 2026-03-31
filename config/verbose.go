package config

import "fmt"

// SetVerbose controls whether verbose logging is enabled.
// This is set by the CLI layer.
var verboseEnabled bool

func SetVerbose(v bool) {
	verboseEnabled = v
}

func Verbose(format string, args ...interface{}) {
	if verboseEnabled {
		fmt.Printf("[verbose] "+format+"\n", args...)
	}
}
