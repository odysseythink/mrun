//go:build viper_logger
// +build viper_logger

package mconf

// WithLogger sets a custom logger.
func WithLogger(l Logger) Option {
	return optionFunc(func(v *MConf) {
		v.logger = l
	})
}
