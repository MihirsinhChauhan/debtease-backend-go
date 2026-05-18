package financial_events

// EffectSummary documents how each event type should be interpreted when projecting or auditing state.
func EffectSummary(eventType string) string {
	switch eventType {
	case EventPaymentRecorded:
		return "User-recorded cash payment applied to a debt; splits into principal and interest components against balances."
	case EventDebtInstrumentCreated:
		return "New debt instrument opened with initial balances and loan/card attributes."
	case EventBalanceManuallyAdjusted:
		return "Outstanding balance overwritten via API (manual correction); must not silently overwrite ledger history."
	case EventEMIAutopApplied:
		return "Scheduled EMI reducing principal on due date (worker path)."
	case EventMoratoriumInterestCapitalized:
		return "Moratorium monthly interest capitalized into outstanding principal."
	case EventLoanMoratoriumExited:
		return "Moratorium ended; EMI recomputed from remaining tenure and balance."
	case EventInterestAccrued:
		return "Interest accrued (e.g. daily credit-card path) increasing accrued_interest / card buckets."
	case EventStatementCycleAdvanced:
		return "Credit card billing cycle rolled; unbilled moved to billed; grace restarted."
	default:
		return ""
	}
}
