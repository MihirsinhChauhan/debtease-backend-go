package debt

// DebtCharacteristics mirrors TDD §4.3 (engine-facing semantics).
type DebtCharacteristics struct {
	Secured             bool
	Revolving           bool
	ShortTerm           bool
	FixedObligation     bool
	InterestBearing     bool
	FragmentationWeight float64
	StressPriority      int
}
