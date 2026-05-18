package financial_events_test

import (
	financial_events "DebtEase/internal/domain/financial_events/v1"
	"testing"
)

func TestEffectSummary_knownEventsNonEmpty(t *testing.T) {
	types := []string{
		financial_events.EventPaymentRecorded,
		financial_events.EventDebtInstrumentCreated,
		financial_events.EventBalanceManuallyAdjusted,
		financial_events.EventEMIAutopApplied,
		financial_events.EventMoratoriumInterestCapitalized,
		financial_events.EventLoanMoratoriumExited,
		financial_events.EventInterestAccrued,
		financial_events.EventStatementCycleAdvanced,
	}
	for _, et := range types {
		t.Run(et, func(t *testing.T) {
			if financial_events.EffectSummary(et) == "" {
				t.Fatalf("missing effect summary for %q", et)
			}
		})
	}
}
