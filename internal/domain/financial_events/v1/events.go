package financial_events

// DefaultCurrency matches persisted financial_events.currency default (schema).
const DefaultCurrency = "INR"

// Domain ledger event types (append-only). Not accounting debit/credit semantics.
const (
	EventPaymentRecorded               = "PAYMENT_RECORDED"
	EventDebtInstrumentCreated         = "DEBT_INSTRUMENT_CREATED"
	EventBalanceManuallyAdjusted       = "BALANCE_MANUALLY_ADJUSTED"
	EventEMIAutopApplied               = "EMI_AUTOP_APPLIED"
	EventMoratoriumInterestCapitalized = "MORATORIUM_INTEREST_CAPITALIZED"
	EventLoanMoratoriumExited          = "LOAN_MORATORIUM_EXITED"
	EventInterestAccrued               = "INTEREST_ACCRUED"
	EventStatementCycleAdvanced        = "STATEMENT_CYCLE_ADVANCED"
)
