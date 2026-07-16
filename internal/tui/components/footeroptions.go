package components

// FooterOptions configures the two-line Pi-style footer.
type FooterOptions struct {
	CWD      string
	Home     string
	Branch   string
	Provider string
	Model    string
	Notice   string
	Used     int
	Context  int

	Dim         func(string) string
	NoticeStyle func(string) string
}

func footerNormalizeOptions(opts FooterOptions) FooterOptions {
	if opts.Used < 0 {
		opts.Used = 0
	}
	if opts.Context <= 0 {
		opts.Context = 128000
	}
	if opts.Dim == nil {
		opts.Dim = footerIdentity
	}
	if opts.NoticeStyle == nil {
		opts.NoticeStyle = opts.Dim
	}
	return opts
}

func footerIdentity(value string) string {
	return value
}
