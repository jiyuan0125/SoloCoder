package natsort

type Options struct {
	Ascending       bool
	CaseSensitive   bool
	IgnoreLeadingZeros bool
}

func DefaultOptions() Options {
	return Options{
		Ascending:       true,
		CaseSensitive:   false,
		IgnoreLeadingZeros: true,
	}
}
