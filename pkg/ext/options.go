package ext

import "strings"

// Option type
type Option func(*Options)

// Options struct
type Options struct {
	ShowPercentage  bool
	CompleteMessage string
}

func newOptions(opts ...Option) Options {
	options := Options{
		ShowPercentage:  false,
		CompleteMessage: "ready",
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

// ShowPercentage to enable percentage view
func ShowPercentage() Option {
	return func(o *Options) {
		o.ShowPercentage = true
	}
}

// CompleteMessage to set task complete message
func CompleteMessage(s string) Option {
	return func(o *Options) {
		if msg := strings.TrimSpace(s); len(msg) > 0 {
			o.CompleteMessage = msg
		}
	}
}
