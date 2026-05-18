package financial_events

// Common metadata JSON keys written alongside domain events (camelCase JSON).
const (
	MetaDebtID             = "debt_id"
	MetaPaymentID          = "payment_id"
	MetaDebtType           = "debt_type"
	MetaOpeningBalance     = "opening_balance"
	MetaEMIAmount          = "emi_amount"
	MetaPrincipalApplied   = "principal_applied"
	MetaInterestApplied    = "interest_applied"
	MetaPaymentType        = "payment_type"
	MetaPreviousBalance    = "previous_balance"
	MetaNewBalance         = "new_balance"
	MetaInterestDelta      = "interest_delta"
	MetaPrincipalComponent = "principal_component"
	MetaInterestComponent  = "interest_component"
	MetaNewEMI             = "new_emi"
)
