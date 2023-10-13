package complete

// Predictor can predict completion options.
type Predictor interface {
	// Predict returns prediction options for a given prefix. The prefix is what currently is typed
	// as a hint for what to return, but the returned values can have any prefix. The returned
	// values will be filtered by the prefix when needed regardless. The prefix may be empty which
	// means that no value was typed.
	Predict(prefix string) []string
}
