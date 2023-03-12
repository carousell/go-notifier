package notifier

import "sync"

var (
	defaultNotifier Notifier
	mu              sync.Mutex
	once            sync.Once
)

type notifier struct {
	baseNotifier Notifier
}

func Notify(err error, rawData ...interface{}) error {
	return GetNotifier().Notify(err, rawData...)
}

func NotifyWithLevel(err error, level string, rawData ...interface{}) error {
	return GetNotifier().NotifyWithLevel(err, level, rawData...)
}

func NotifyOnPanic(rawData ...interface{}) {
	GetNotifier().NotifyOnPanic(rawData...)
}

func NewNotifier(n Notifier) Notifier {
	return n
}

func GetNotifier() Notifier {
	return defaultNotifier
}

func SetNotifier(l Notifier) {
	if l != nil {
		mu.Lock()
		defer mu.Unlock()
		defaultNotifier = l
	}
}
func RegisterNotifier(n Notifier) {
	once.Do(func() {
		if defaultNotifier == nil {
			defaultNotifier = NewNotifier(n)
		}
	})
}
