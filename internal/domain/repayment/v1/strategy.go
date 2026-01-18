package repayment_v1

type loanState struct {
	Loan    Loan
	Balance float64
	EMI     float64
	Active  bool
}

func ApplyStrategy(
	loans []Loan,
	strategy StrategyType,
	monthlyExtra float64,
) map[int][]EMIRecord {

	states := make(map[string]*loanState)
	for _, l := range loans {
		states[l.ID] = &loanState{
			Loan:    l,
			Balance: l.Principal,
			EMI:     CalculateEMI(l.Principal, l.AnnualRate, l.TenureMonths),
			Active:  true,
		}
	}

	portfolio := make(map[int][]EMIRecord)
	month := 1

	for anyActive(states) {
		var records []EMIRecord

		// Normal EMI pass
		for _, s := range states {
			if !s.Active {
				continue
			}

			r := s.Loan.AnnualRate / 12 / 100
			opening := s.Balance
			interest := round(opening*r, 2)
			emi := min(s.EMI, opening+interest)
			principal := round(emi-interest, 2)
			closing := round(opening-principal, 2)

			if closing <= 0 {
				closing = 0
				s.Active = false
			}

			s.Balance = closing

			records = append(records, EMIRecord{
				LoanID:         s.Loan.ID,
				Month:          month,
				OpeningBalance: opening,
				EMI:            emi,
				Interest:       interest,
				Principal:      principal,
				ClosingBalance: closing,
			})
		}

		// Prepayment
		if monthlyExtra > 0 && strategy != StrategyBaseline {
			target := selectTarget(states, strategy)
			if target != nil {
				extra := min(monthlyExtra, target.Balance)
				target.Balance = round(target.Balance-extra, 2)

				for i := range records {
					if records[i].LoanID == target.Loan.ID {
						records[i].Principal += extra
						records[i].ClosingBalance = target.Balance
						break
					}
				}

				if target.Balance <= 0 {
					target.Active = false
				}
			}
		}

		portfolio[month] = records
		month++
	}

	return portfolio
}

func selectTarget(states map[string]*loanState, strategy StrategyType) *loanState {
	var selected *loanState

	for _, s := range states {
		if !s.Active {
			continue
		}

		if selected == nil {
			selected = s
			continue
		}

		switch strategy {
		case StrategySnowball:
			if s.Balance < selected.Balance {
				selected = s
			}
		case StrategyAvalanche:
			if s.Loan.AnnualRate > selected.Loan.AnnualRate {
				selected = s
			}
		}
	}

	return selected
}

func anyActive(states map[string]*loanState) bool {
	for _, s := range states {
		if s.Active {
			return true
		}
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
