package rollbar

import (
	"context"
	"github.com/carousell/go-logging"
	"github.com/carousell/go-notifier"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/stvp/rollbar"
	"strings"
)

// InitRollbar inits rollbar configuration
func InitRollbar(token, env string) (*rollbarNotifier, error) {
	rollbar.Token = token
	rollbar.Environment = env
	return &rollbarNotifier{
		token:       token,
		environment: env,
		inited:      true,
	}, nil
}

type rollbarNotifier struct {
	inited      bool
	token       string
	environment string
}

func (r *rollbarNotifier) Notify(err error, rawData ...interface{}) error {
	return r.NotifyWithLevelAndSkip(err, 2, notifier.ErrorLevel, rawData...)
}

func (r *rollbarNotifier) NotifyWithLevel(err error, level string, rawData ...interface{}) error {
	return r.NotifyWithLevelAndSkip(err, 2, level, rawData...)
}

func (r *rollbarNotifier) NotifyWithLevelAndSkip(err error, skip int, level string, rawData ...interface{}) error {

	if err == nil {
		return nil
	}

	if n, ok := err.(notifier.NotifyExt); ok {
		if !n.ShouldNotify() {
			return err
		}
		n.Notified(true)
	}
	return r.doNotify(err, skip, level, rawData...)
}

func (r *rollbarNotifier) doNotify(err error, skip int, level string, rawData ...interface{}) error {

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

	parsedData, _ := notifier.ParseRawData(ctx, list...)
	if r.inited {
		var fields []*rollbar.Field
		if len(list) > 0 {
			for k, v := range parsedData {
				fields = append(fields, &rollbar.Field{Name: k, Data: v})
			}
		}
		if traceID != "" {
			fields = append(fields, &rollbar.Field{Name: "traceId", Data: traceID})
		}
		fields = append(fields, &rollbar.Field{
			Name: "server",
			Data: map[string]interface{}{
				"hostname": notifier.GetHostname(),
				"root":     notifier.GetServerRoot(),
			},
		})
		rollbar.ErrorWithStack(sev.String(), errWithStack, convToRollbar(errWithStack.StackFrame()), fields...)
	}

	logging.GetLogger().Log(ctx, sev.LoggerLevel(), skip+1, "err", errWithStack, "stack", errWithStack.StackFrame())
	return err
}

func (r *rollbarNotifier) NotifyOnPanic(rawData ...interface{}) {

}

func convToRollbar(in []notifier.StackFrame) rollbar.Stack {
	out := rollbar.Stack{}
	for _, s := range in {
		out = append(out, rollbar.Frame{
			Filename: s.File,
			Method:   s.Func,
			Line:     s.Line,
		})
	}
	return out
}
