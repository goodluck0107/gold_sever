package signal

import "context"

type Option func(options *Options)

type Options struct {
	fnPIPE, fnTSTP, fnTRAP, fnINT func(ctx context.Context)
}

func WithPIPE(fn func(ctx context.Context)) Option {
	return func(options *Options) {
		options.fnPIPE = fn
	}
}

// Ctrl Z
func WithTSTP(fn func(ctx context.Context)) Option {
	return func(options *Options) {
		options.fnTSTP = fn
	}
}

func WithTRAP(fn func(ctx context.Context)) Option {
	return func(options *Options) {
		options.fnTRAP = fn
	}
}

// Ctrl C
func WithINT(fn func(ctx context.Context)) Option {
	return func(options *Options) {
		options.fnINT = fn
	}
}

func makeOptions(opts ...Option) Options {
	ret := Options{}
	for _, o := range opts {
		o(&ret)
	}
	return ret
}
