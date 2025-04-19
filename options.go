package rulesengine

type (
	LoggerFunc func(
		fieldName string,
		operator Operator,
		actual,
		expected interface{},
	)

	Options struct {
		Logger LoggerFunc
		Timing bool
	}
)

func DefaultOptions() Options {
	return Options{
		Timing: false,
		Logger: func(
			fieldName string,
			operator Operator,
			actual, expected interface{},
		) {
		},
	}
}

func (o Options) WithTiming() Options {
	o.Timing = true
	return o
}

func (o Options) WithLogger(logger LoggerFunc) Options {
	o.Logger = logger
	return o
}
