package http

type Options struct {
	isVerbose bool
}

type Option func(*Options)

func SetVerbose(isVerbose bool) Option {
	return func(options *Options) {
		options.isVerbose = isVerbose
	}
}
