package debt

import (
	"database/sql"
	"time"

	"DebtEase/internal/database"

	"github.com/shopspring/decimal"
)

// CreateObligationInput captures API inputs needed to derive obligation rows.
type CreateObligationInput struct {
	DebtType      database.DebtType
	InterestRate  decimal.Decimal
	MinPayment    decimal.Decimal
	Principal     decimal.Decimal
	EMIAmount     *decimal.Decimal
	TenureMonths  *int32
	PaymentDueDay *int32
	DueDate       time.Time
}

// CreateObligationDerived is the normalized obligation slice for persistence.
type CreateObligationDerived struct {
	InstrumentType      database.DebtInstrumentType
	InterestModel       database.InterestModel
	MonthlyObligation   decimal.Decimal
	MinimumObligation   decimal.Decimal
	RemainingMonths     sql.NullInt32
	DueDay              sql.NullInt32
	IsSecured           bool
	IsRevolving         bool
	IsShortTerm         bool
	IsInterestBearing   bool
	FragmentationWeight decimal.Decimal
	StressPriority      int16
	EngineVersion       string
	PolicyVersion       string
}

// NormalizeCreate derives canonical obligation fields from user input (TDD §7).
func NormalizeCreate(in CreateObligationInput) CreateObligationDerived {
	inst := MapDebtTypeToInstrument(in.DebtType)
	model := InferInterestModel(in.DebtType, in.InterestRate)
	ch := CharacteristicsForInstrument(inst)
	ch = ApplyInterestModelOverrides(ch, model)
	// Classifier is source of truth for structural flags; interest-bearing still respects rate for generic types.
	isInterestBearing := ch.InterestBearing
	if !in.InterestRate.IsZero() && model != database.InterestModelZeroCostEmi {
		isInterestBearing = true
	}
	if inst == database.DebtInstrumentTypeBnpl {
		isInterestBearing = false
	}

	monthly := in.MinPayment
	if in.EMIAmount != nil && !in.EMIAmount.IsZero() {
		monthly = *in.EMIAmount
	} else if installmentType(in.DebtType) && in.TenureMonths != nil && *in.TenureMonths > 0 {
		monthly = calculateEMI(in.Principal, in.InterestRate, int(*in.TenureMonths))
	}

	var rem sql.NullInt32
	if in.TenureMonths != nil {
		rem = sql.NullInt32{Int32: *in.TenureMonths, Valid: true}
	}

	var dueDay sql.NullInt32
	if in.PaymentDueDay != nil {
		dueDay = sql.NullInt32{Int32: *in.PaymentDueDay, Valid: true}
	} else {
		d := int32(in.DueDate.Day())
		dueDay = sql.NullInt32{Int32: d, Valid: true}
	}

	fw := decimal.NewFromFloat(ch.FragmentationWeight).Round(2)

	return CreateObligationDerived{
		InstrumentType:      inst,
		InterestModel:       model,
		MonthlyObligation:   monthly,
		MinimumObligation:   in.MinPayment,
		RemainingMonths:     rem,
		DueDay:              dueDay,
		IsSecured:           ch.Secured,
		IsRevolving:         ch.Revolving,
		IsShortTerm:         ch.ShortTerm,
		IsInterestBearing:   isInterestBearing,
		FragmentationWeight: fw,
		StressPriority:      int16(ch.StressPriority),
		EngineVersion:       DefaultEngineVersion,
		PolicyVersion:       DefaultPolicyVersion,
	}
}

func installmentType(dt database.DebtType) bool {
	switch dt {
	case database.DebtTypePersonalLoan, database.DebtTypeStudentLoan, database.DebtTypeMortgage,
		database.DebtTypeHomeLoan, database.DebtTypeCarLoan, database.DebtTypeConsumerEmi:
		return true
	default:
		return false
	}
}

func calculateEMI(principal, annualRate decimal.Decimal, months int) decimal.Decimal {
	if months == 0 {
		return decimal.Zero
	}
	monthlyRate := annualRate.Div(decimal.NewFromInt(1200))
	if monthlyRate.IsZero() {
		return principal.Div(decimal.NewFromInt(int64(months)))
	}
	onePlusR := decimal.NewFromInt(1).Add(monthlyRate)
	onePlusRPowN := onePlusR.Pow(decimal.NewFromInt(int64(months)))
	numerator := principal.Mul(monthlyRate).Mul(onePlusRPowN)
	denominator := onePlusRPowN.Sub(decimal.NewFromInt(1))
	return numerator.Div(denominator).Round(2)
}
