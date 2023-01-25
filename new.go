package progressbar

func Add(maxBytes int64, title string, opts ...Opt) MultiPB {
	defaultMPB.Add(maxBytes, title, opts...)
	return defaultMPB
}

func New(opts ...MOpt) MultiPB {
	bar := multiBar(opts...)
	return bar
}
