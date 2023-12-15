package strutils

// LevenshteinParams represents a set of parameter values for the various formulas involved
// in the calculation of the Levenshtein string metrics.
type LevenshteinParams struct {
	insCost        int
	subCost        int
	delCost        int
	maxCost        int
	minScore       float64
	bonusPrefix    int
	bonusScale     float64
	bonusThreshold float64
}

var (
	defaultLevenshteinParams = NewParams()
)

// NewParams creates a new set of parameters and initializes it with the default values.
func NewParams() *LevenshteinParams {
	return &LevenshteinParams{
		insCost:        1,
		subCost:        1,
		delCost:        1,
		maxCost:        0,
		minScore:       0,
		bonusPrefix:    4,
		bonusScale:     .1,
		bonusThreshold: .7,
	}
}

// Clone returns a pointer to a copy of the receiver parameter set, or of a new
// default parameter set if the receiver is nil.
func (p *LevenshteinParams) Clone() *LevenshteinParams {
	if p == nil {
		return NewParams()
	}
	return &LevenshteinParams{
		insCost:        p.insCost,
		subCost:        p.subCost,
		delCost:        p.delCost,
		maxCost:        p.maxCost,
		minScore:       p.minScore,
		bonusPrefix:    p.bonusPrefix,
		bonusScale:     p.bonusScale,
		bonusThreshold: p.bonusThreshold,
	}
}

// InsCost overrides the default value of 1 for the cost of insertion.
// The new value must be zero or positive.
func (p *LevenshteinParams) InsCost(v int) *LevenshteinParams {
	if v >= 0 {
		p.insCost = v
	}
	return p
}

// SubCost overrides the default value of 1 for the cost of substitution.
// The new value must be zero or positive.
func (p *LevenshteinParams) SubCost(v int) *LevenshteinParams {
	if v >= 0 {
		p.subCost = v
	}
	return p
}

// DelCost overrides the default value of 1 for the cost of deletion.
// The new value must be zero or positive.
func (p *LevenshteinParams) DelCost(v int) *LevenshteinParams {
	if v >= 0 {
		p.delCost = v
	}
	return p
}

// MaxCost overrides the default value of 0 (meaning unlimited) for the maximum cost.
// The calculation of Distance() stops when the result is guaranteed to exceed
// this maximum, returning a lower-bound rather than exact value.
// The new value must be zero or positive.
func (p *LevenshteinParams) MaxCost(v int) *LevenshteinParams {
	if v >= 0 {
		p.maxCost = v
	}
	return p
}

// MinScore overrides the default value of 0 for the minimum similarity score.
// Scores below this threshold are returned as 0 by Similarity() and Match().
// The new value must be zero or positive. Note that a minimum greater than 1
// can never be satisfied, resulting in a score of 0 for any pair of strings.
func (p *LevenshteinParams) MinScore(v float64) *LevenshteinParams {
	if v >= 0 {
		p.minScore = v
	}
	return p
}

// BonusPrefix overrides the default value for the maximum length of
// common prefix to be considered for bonus by Match().
// The new value must be zero or positive.
func (p *LevenshteinParams) BonusPrefix(v int) *LevenshteinParams {
	if v >= 0 {
		p.bonusPrefix = v
	}
	return p
}

// BonusScale overrides the default value for the scaling factor used by Match()
// in calculating the bonus.
// The new value must be zero or positive. To guarantee that the similarity score
// remains in the interval 0..1, this scaling factor is not allowed to exceed
// 1 / BonusPrefix.
func (p *LevenshteinParams) BonusScale(v float64) *LevenshteinParams {
	if v >= 0 {
		p.bonusScale = v
	}

	// the bonus cannot exceed (1-sim), or the score may become greater than 1.
	if float64(p.bonusPrefix)*p.bonusScale > 1 {
		p.bonusScale = 1 / float64(p.bonusPrefix)
	}

	return p
}

// BonusThreshold overrides the default value for the minimum similarity score
// for which Match() can assign a bonus.
// The new value must be zero or positive. Note that a threshold greater than 1
// effectively makes Match() become the equivalent of Similarity().
func (p *LevenshteinParams) BonusThreshold(v float64) *LevenshteinParams {
	if v >= 0 {
		p.bonusThreshold = v
	}
	return p
}
