package stress_v1

import (
	classification_v1 "DebtEase/internal/domain/classification/v1"
	"sort"

	"github.com/shopspring/decimal"
)

// categoryForDebt maps a normalized debt into a failure bucket.
// Priority is defined in ComputeFailureOrder.
func categoryForDebt(d StressDebtEntry) DebtFailureCategory {
	if d.InstrumentType == classification_v1.InstrumentBNPL {
		return FailureBNPLOrZeroEMI
	}
	if !d.Secured && d.RemainingMonths <= 12 {
		return FailureShortTermUnsecured
	}
	if !d.Secured {
		return FailureLongTermUnsecured
	}
	return FailureSecured
}

// failurePriority defines the deterministic failure ordering across categories.
func failurePriority(cat DebtFailureCategory) int {
	switch cat {
	case FailureBNPLOrZeroEMI:
		return 0
	case FailureShortTermUnsecured:
		return 1
	case FailureLongTermUnsecured:
		return 2
	case FailureSecured:
		return 3
	default:
		return 100
	}
}

// ComputeFailureOrder simulates which categories would fail first if the buffer
// after shock is negative. It consumes the deficit by dropping EMIs in the
// configured priority order and returns the distinct categories that would fail.
func ComputeFailureOrder(
	bufferAfterShock decimal.Decimal,
	shockedEMIs []decimal.Decimal,
	debts []StressDebtEntry,
) []DebtFailureCategory {
	if bufferAfterShock.GreaterThanOrEqual(decimalZero) {
		return nil
	}

	if len(shockedEMIs) != len(debts) {
		// Programming error; keep behavior deterministic but do not panic in production.
		return nil
	}

	deficit := bufferAfterShock.Neg() // make positive

	type idxWithPriority struct {
		Idx      int
		Priority int
	}

	order := make([]idxWithPriority, len(debts))
	for i, d := range debts {
		cat := categoryForDebt(d)
		order[i] = idxWithPriority{
			Idx:      i,
			Priority: failurePriority(cat),
		}
	}

	sort.Slice(order, func(i, j int) bool {
		return order[i].Priority < order[j].Priority
	})

	var failed []DebtFailureCategory
	seen := make(map[DebtFailureCategory]struct{})

	for _, item := range order {
		if deficit.LessThanOrEqual(decimalZero) {
			break
		}
		idx := item.Idx
		if idx < 0 || idx >= len(shockedEMIs) {
			continue
		}
		emi := shockedEMIs[idx]
		deficit = deficit.Sub(emi)

		cat := categoryForDebt(debts[idx])
		if _, ok := seen[cat]; !ok {
			seen[cat] = struct{}{}
			failed = append(failed, cat)
		}
	}

	return failed
}

