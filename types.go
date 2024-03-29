package notifier

// We use string as argument to determine the log level from day one,
// so it's not a big deal to make breaking changes to force everyone to adopt.
// It turns out we expose string type constants for users to use and keep severity type as private for now
type severity string

const (
	ErrorLevel    = "error"
	DebugLevel    = "debug"
	InfoLevel     = "info"
	WarningLevel  = "warning"
	FatalLevel    = "fatal"
	CriticalLevel = "critical"
)

const (
	errorSeverity    = severity(ErrorLevel)
	debugSeverity    = severity(DebugLevel)
	infoSeverity     = severity(InfoLevel)
	warningSeverity  = severity(WarningLevel)
	fatalSeverity    = severity(FatalLevel)
	criticalSeverity = severity(FatalLevel)
)

var (
	// mainly for mapping string back used
	levelSeverityMap = map[string]severity{
		ErrorLevel:    errorSeverity,
		DebugLevel:    debugSeverity,
		InfoLevel:     infoSeverity,
		WarningLevel:  warningSeverity,
		FatalLevel:    fatalSeverity,
		CriticalLevel: fatalSeverity,
	}
)

type Notifier interface {
	Notify(err error, rawData ...interface{}) error
	NotifyWithLevel(err error, level string, rawData ...interface{}) error
	NotifyOnPanic(rawData ...interface{})
}
