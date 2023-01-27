package progressbar

import "time"

func Add(maxBytes int64, title string, opts ...Opt) MultiPB {
	defaultMPB.Add(maxBytes, title, opts...)
	return defaultMPB
}

func New(opts ...MOpt) MultiPB {
	bar := multiBar(opts...)
	return bar
}

type SchemaData struct {
	Indent  string
	Prepend string
	Bar     string
	Percent string
	Title   string
	Current string
	Total   string
	Elapsed string
	Speed   string
	Append  string

	PercentFloat float64
	ElapsedTime  time.Duration
}
