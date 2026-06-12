package agent

import "strings"

type logLevel int

const (
	logLevelInfo logLevel = iota
	logLevelDebug
)

// parseLogLevel converts a LOG_LEVEL environment value into a logLevel,
// defaulting to info for empty or unrecognized values.
func parseLogLevel(value string) logLevel {
	if strings.EqualFold(value, "debug") {
		return logLevelDebug
	}

	return logLevelInfo
}
