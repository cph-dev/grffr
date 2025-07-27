package options

// WithDebug enables debug logging in the framework.
func WithNoBanner(cfg *Configuration) {
	cfg.Banner = false
}
