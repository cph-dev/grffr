package options

// WithDebug enables debug logging in the framework.
func WithDebug(cfg *Configuration) {
	cfg.Debug = true
}
