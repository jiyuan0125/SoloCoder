package bptree

type Options struct {
	Order          int
	AllowDuplicates bool
}

func DefaultOptions() *Options {
	return &Options{
		Order:           4,
		AllowDuplicates: false,
	}
}

func (o *Options) validate() {
	if o.Order < 3 {
		o.Order = 3
	}
}
