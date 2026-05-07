package classification_v1

import "github.com/shopspring/decimal"

// TreatmentCategory is the primary output of the classification domain.
type TreatmentCategory string

const (
	SafeToHold      TreatmentCategory = "SAFE_TO_HOLD"
	NeedsMonitoring TreatmentCategory = "NEEDS_MONITORING"
	AvoidAdding     TreatmentCategory = "AVOID_ADDING"
	ActivelyReduce  TreatmentCategory = "ACTIVELY_REDUCE"
)

// Flag explains why a treatment was assigned (explainability).
type Flag string

const (
	FlagHighCashflowRigidity  Flag = "HIGH_CASHFLOW_RIGIDITY"
	FlagShortTermDebtStacking Flag = "SHORT_TERM_DEBT_STACKING"
	FlagLowShockBuffer        Flag = "LOW_SHOCK_BUFFER"
	FlagUnsecuredOverexposure Flag = "UNSECURED_OVEREXPOSURE"
)

// ConfidenceLevel indicates how confident the classification is.
type ConfidenceLevel string

const (
	ConfidenceLow    ConfidenceLevel = "LOW"
	ConfidenceMedium ConfidenceLevel = "MEDIUM"
	ConfidenceHigh   ConfidenceLevel = "HIGH"
)

// DebtInstrumentType categorizes the debt for fragmentation and secured/unsecured logic.
type DebtInstrumentType string

const (
	InstrumentBNPL         DebtInstrumentType = "bnpl"
	InstrumentLOC          DebtInstrumentType = "loc"
	InstrumentCreditCard   DebtInstrumentType = "credit_card"
	InstrumentPersonalLoan DebtInstrumentType = "personal_loan"
	InstrumentStudentLoan  DebtInstrumentType = "student_loan"
	InstrumentMortgage     DebtInstrumentType = "mortgage"
	InstrumentMedical      DebtInstrumentType = "medical"
	InstrumentOther        DebtInstrumentType = "other"
)

// DebtEntry is a single normalized debt for classification.
// Handler/service layer maps database.Debt into this.
type DebtEntry struct {
	ID              string
	EMI             decimal.Decimal
	RemainingMonths int
	Secured         bool
	InstrumentType  DebtInstrumentType
}

// FinancialContext is the user's financial state (from Financial State domain).
type FinancialContext struct {
	MonthlyNetIncome          decimal.Decimal
	FixedObligations          decimal.Decimal
	ExistingEMIs              decimal.Decimal // portfolio total EMI
	ShortTermObligationsCount int             // BNPL, 0-cost EMI, LOC count
}

// ClassificationInput is the input to the classifier.
type ClassificationInput struct {
	Debts   []DebtEntry
	Context FinancialContext
}

// ContextMetrics are the five computed context metrics.
type ContextMetrics struct {
	EMIRatio           decimal.Decimal // Total_Monthly_EMI / Monthly_Net_Income
	FragmentationIndex decimal.Decimal // Active_Debts + BNPL_Count * 0.5
	RigidityScore      decimal.Decimal // SUM(EMI * Remaining_Months) / Monthly_Income
	ShockBufferMonths  decimal.Decimal // (Income - Fixed_Obligations - EMIs) / EMIs
	UnsecuredRatio     decimal.Decimal // Unsecured_EMI / Total_EMI
}

// ClassificationResult is the output of Classify().
type ClassificationResult struct {
	Treatment     TreatmentCategory
	Flags         []Flag
	Confidence    ConfidenceLevel
	Metrics       ContextMetrics
	EngineVersion string
}
