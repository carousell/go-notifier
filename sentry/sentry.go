package sentry

import (
	"context"
	"github.com/carousell/go-logging"
	"github.com/carousell/go-notifier"
	"github.com/getsentry/raven-go"
	stdopentracing "github.com/opentracing/opentracing-go"
	"runtime"
	"strings"
)

func InitSentry(dsn string) (*sentryNotifier, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return nil, err
	}
	return &sentryNotifier{
		inited: true,
		client: client,
	}, nil
}

type sentryNotifier struct {
	inited bool
	client *raven.Client
}

func (s *sentryNotifier) Notify(err error, rawData ...interface{}) error {
	return s.NotifyWithLevelAndSkip(err, 2, notifier.ErrorLevel, rawData...)
}

func (s *sentryNotifier) NotifyWithLevel(err error, level string, rawData ...interface{}) error {
	return s.NotifyWithLevelAndSkip(err, 2, level, rawData...)
}

func (s *sentryNotifier) NotifyWithLevelAndSkip(err error, skip int, level string, rawData ...interface{}) error {

	if err == nil {
		return nil
	}

	if n, ok := err.(notifier.NotifyExt); ok {
		if !n.ShouldNotify() {
			return err
		}
		n.Notified(true)
	}
	return s.doNotify(err, skip, level, rawData...)
}

func (s *sentryNotifier) doNotify(err error, skip int, level string, rawData ...interface{}) error {

	if err == nil {
		return nil
	}
	sev := notifier.ParseLevel(level)

	// add stack information
	errWithStack, ok := err.(notifier.ErrorExt)
	if !ok {
		errWithStack = notifier.WrapWithSkip(err, "", skip+1)
	}

	list := make([]interface{}, 0)
	for pos := range rawData {
		data := rawData[pos]
		// if we find the error, return error and do not log it
		if e, ok := data.(error); ok {
			if e == err {
				return err
			} else if er, ok := e.(notifier.ErrorExt); ok {
				if err == er.Cause() {
					return err
				}
			}
		} else {
			list = append(list, rawData[pos])
		}
	}

	// try to fetch a traceID and context from rawData
	var traceID string
	ctx := context.Background()
	for _, d := range list {
		if c, ok := d.(context.Context); ok {
			if span := stdopentracing.SpanFromContext(c); span != nil {
				traceID = span.BaggageItem("trace")
			}
			if strings.TrimSpace(traceID) == "" {
				traceID = notifier.GetTraceId(c)
			}
			ctx = c
			break
		}
	}

	parsedData, tagData := notifier.ParseRawData(ctx, list...)
	if s.inited {
		ravenExp := raven.NewException(errWithStack, convToSentry(errWithStack))
		packet := raven.NewPacketWithExtra(errWithStack.Error(), parsedData, ravenExp)

		for _, tags := range tagData {
			packet.AddTags(tags)
		}

		// type assert directly since it's single use case, so we don't consider wrapping it for now
		packet.Level = raven.Severity(sev)
		s.client.Capture(packet, nil)
	}

	logging.GetLogger().Log(ctx, sev.LoggerLevel(), skip+1, "err", errWithStack, "stack", errWithStack.StackFrame())
	return err
}

func (s *sentryNotifier) NotifyOnPanic(rawData ...interface{}) {

}

func convToSentry(in notifier.ErrorExt) *raven.Stacktrace {
	out := new(raven.Stacktrace)
	pcs := in.Callers()
	frames := make([]*raven.StacktraceFrame, 0)

	callersFrames := runtime.CallersFrames(pcs)

	for {
		fr, more := callersFrames.Next()
		if fr.Func != nil {
			frame := raven.NewStacktraceFrame(fr.PC, fr.Function, fr.File, fr.Line, 3, []string{})
			if frame != nil {
				frame.InApp = true
				frames = append(frames, frame)
			}
		}
		if !more {
			break
		}
	}
	for i := len(frames)/2 - 1; i >= 0; i-- {
		opp := len(frames) - 1 - i
		frames[i], frames[opp] = frames[opp], frames[i]
	}
	out.Frames = frames
	return out
}
