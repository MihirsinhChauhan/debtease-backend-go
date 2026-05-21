package debt

import (
	"DebtEase/internal/database"

	"github.com/shopspring/decimal"
)

// MapDebtTypeToInstrument maps persisted `debts.type` (debt_type enum) to canonical instrument.
func MapDebtTypeToInstrument(dt database.DebtType) database.DebtInstrumentType {
	switch dt {
	case database.DebtTypeCreditCard:
		return database.DebtInstrumentTypeCreditCard
	case database.DebtTypePersonalLoan:
		return database.DebtInstrumentTypePersonalLoan
	case database.DebtTypeStudentLoan:
		return database.DebtInstrumentTypeEducationLoan
	case database.DebtTypeMortgage:
		return database.DebtInstrumentTypeHomeLoan
	case database.DebtTypeMedical:
		return database.DebtInstrumentTypeMedicalLoan
	case database.DebtTypeOther:
		return database.DebtInstrumentTypePersonalLoan
	case database.DebtTypeHomeLoan:
		return database.DebtInstrumentTypeHomeLoan
	case database.DebtTypeCarLoan:
		return database.DebtInstrumentTypeCarLoan
	case database.DebtTypeConsumerEmi:
		return database.DebtInstrumentTypeConsumerEmi
	case database.DebtTypeBnpl:
		return database.DebtInstrumentTypeBnpl
	default:
		return database.DebtInstrumentTypePersonalLoan
	}
}

// InferInterestModel derives interest behavior from legacy type + APR (TDD §4.2).
func InferInterestModel(dt database.DebtType, annualRate decimal.Decimal) database.InterestModel {
	switch dt {
	case database.DebtTypeCreditCard:
		return database.InterestModelRevolving
	case database.DebtTypeConsumerEmi:
		if annualRate.IsZero() {
			return database.InterestModelZeroCostEmi
		}
		return database.InterestModelFixed
	case database.DebtTypeHomeLoan, database.DebtTypeMortgage:
		return database.InterestModelFloating
	default:
		return database.InterestModelFixed
	}
}

// CharacteristicsForInstrument returns deterministic structural metadata (TDD §8).
func CharacteristicsForInstrument(inst database.DebtInstrumentType) DebtCharacteristics {
	switch inst {
	case database.DebtInstrumentTypeBnpl:
		return DebtCharacteristics{
			Secured: false, Revolving: false, ShortTerm: true,
			FixedObligation: true, InterestBearing: false,
			FragmentationWeight: 0.5, StressPriority: 1,
		}
	case database.DebtInstrumentTypeHomeLoan:
		return DebtCharacteristics{
			Secured: true, Revolving: false, ShortTerm: false,
			FixedObligation: true, InterestBearing: true,
			FragmentationWeight: 1.0, StressPriority: 4,
		}
	case database.DebtInstrumentTypeCreditCard:
		return DebtCharacteristics{
			Secured: false, Revolving: true, ShortTerm: false,
			FixedObligation: false, InterestBearing: true,
			FragmentationWeight: 0.7, StressPriority: 2,
		}
	case database.DebtInstrumentTypeConsumerEmi:
		return DebtCharacteristics{
			Secured: false, Revolving: false, ShortTerm: false,
			FixedObligation: true, InterestBearing: true,
			FragmentationWeight: 0.6, StressPriority: 2,
		}
	case database.DebtInstrumentTypeCarLoan:
		return DebtCharacteristics{
			Secured: true, Revolving: false, ShortTerm: false,
			FixedObligation: true, InterestBearing: true,
			FragmentationWeight: 0.95, StressPriority: 4,
		}
	case database.DebtInstrumentTypeEducationLoan:
		return DebtCharacteristics{
			Secured: false, Revolving: false, ShortTerm: false,
			FixedObligation: true, InterestBearing: true,
			FragmentationWeight: 0.8, StressPriority: 3,
		}
	case database.DebtInstrumentTypeMedicalLoan:
		return DebtCharacteristics{
			Secured: false, Revolving: false, ShortTerm: false,
			FixedObligation: true, InterestBearing: true,
			FragmentationWeight: 0.85, StressPriority: 3,
		}
	default:
		return DebtCharacteristics{
			Secured: false, Revolving: false, ShortTerm: false,
			FixedObligation: true, InterestBearing: true,
			FragmentationWeight: 0.85, StressPriority: 3,
		}
	}
}

// ApplyInterestModelOverrides adjusts interest-bearing flag for zero-cost EMI.
func ApplyInterestModelOverrides(ch DebtCharacteristics, model database.InterestModel) DebtCharacteristics {
	if model == database.InterestModelZeroCostEmi {
		ch.InterestBearing = false
	}
	return ch
}
