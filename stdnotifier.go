package notifier

import (
	"context"
	"github.com/carousell/go-logging"
)

type stdNotifier struct{}

// NewStdNotifier returns Notifier which logs on default logger
func NewStdNotifier() Notifier {
	return &stdNotifier{}
}

func (s *stdNotifier) Notify(err error, rawData ...interface{}) error {
	if err == nil {
		return nil
	}
	logging.GetLogger().Log(context.Background(), logging.ErrorLevel, 3, "err", err, "rawData", rawData)
	return err
}

func (s *stdNotifier) NotifyWithLevel(err error, level string, rawData ...interface{}) error {
	if err == nil {
		return nil
	}
	sev := ParseLevel(level)
	logging.GetLogger().Log(context.Background(), sev.LoggerLevel(), 3, "err", err, "rawData", rawData)
	return err
}

func (s *stdNotifier) NotifyOnPanic(rawData ...interface{}) {
	logging.GetLogger().Log(context.Background(), logging.ErrorLevel, 3, "is_panic", "true", "rawData", rawData)
}
