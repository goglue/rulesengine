package rulesengine

type (
	// LoggerFunc is func type that accepts the different [Rule] attributes to
	// be logged.
	LoggerFunc func(
		fieldName string,
		operator Operator,
		actual,
		expected any,
	)

	// Options type are the configurations that enables/disables the debugging
	// of the engine.
	Options struct {
		Logger LoggerFunc
		Timing bool
	}
)

// DefaultOptions method returns a default options type as a builder pattern.
func DefaultOptions() Options {
	return Options{
		Timing: false,
		Logger: func(
			fieldName string,
			operator Operator,
			actual, expected any,
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
