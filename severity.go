package notifier

import "github.com/carousell/logging"

func (s severity) String() string {
	return string(s)
}

func (s severity) LoggerLevel() logging.Level {
	switch s {
	case warningSeverity:
		return logging.WarnLevel
	case infoSeverity:
		return logging.InfoLevel
	case debugSeverity:
		return logging.DebugLevel
	case errorSeverity:
		return logging.ErrorLevel
	default:
		return logging.ErrorLevel
	}
}

func ParseLevel(s string) severity {
	sev, ok := levelSeverityMap[s]
	if !ok {
		sev = errorSeverity // by default
	}
	return sev
}
