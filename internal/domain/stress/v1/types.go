package stress_v1

import (
	classification_v1 "DebtEase/internal/domain/classification/v1"

	"github.com/shopspring/decimal"
)

// StressLevel is the primary qualitative outcome of the stress engine.
type StressLevel string

const (
	StressResilient StressLevel = "RESILIENT"
	StressSensitive StressLevel = "SENSITIVE"
	StressFragile   StressLevel = "FRAGILE"
)

// SurvivabilityBand is a coarse estimate of how long the system can sustain EMIs
// under shocked conditions. This is intentionally banded rather than precise months.
type SurvivabilityBand string

const (
	SurvivabilityImmediate SurvivabilityBand = "IMMEDIATE_STRESS"
	SurvivabilityShort     SurvivabilityBand = "SHORT"
	SurvivabilityModerate  SurvivabilityBand = "MODERATE"
	SurvivabilityStable    SurvivabilityBand = "STABLE"
)

// DebtFailureCategory groups debts into ordered buckets for simulated failure.
type DebtFailureCategory string

const (
	FailureBNPLOrZeroEMI     DebtFailureCategory = "BNPL_OR_ZERO_EMI"
	FailureShortTermUnsecured DebtFailureCategory = "SHORT_TERM_UNSECURED"
	FailureLongTermUnsecured  DebtFailureCategory = "LONG_TERM_UNSECURED"
	FailureSecured            DebtFailureCategory = "SECURED"
)

// StressDebtEntry is a minimal, normalized view of a single debt for stress simulation.
// Handler/service layers are responsible for mapping database or engine outputs into this type.
type StressDebtEntry struct {
	ID              string
	EMI             decimal.Decimal
	RemainingMonths int
	Secured         bool
	InstrumentType  classification_v1.DebtInstrumentType
}

// StressInput is the top-level input to the stress engine.
// It represents a snapshot of the user's portfolio and cashflow at a given time.
type StressInput struct {
	Debts             []StressDebtEntry
	MonthlyNetIncome  decimal.Decimal
	FixedObligations  decimal.Decimal
}

// StressOutcome is the deterministic output of the stress engine.
type StressOutcome struct {
	StressLevel           StressLevel
	FlexibilityAfterShock decimal.Decimal
	SurvivabilityBand     SurvivabilityBand
	FailedDebtCategories  []DebtFailureCategory
	KeyDrivers            []string
	EngineVersion         string
}

