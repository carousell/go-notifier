package notifier

type isTags interface {
	isTags()
	value() map[string]string
}

type Tags map[string]string

func (tags Tags) isTags() {}

func (tags Tags) value() map[string]string {
	return map[string]string(tags)
}
