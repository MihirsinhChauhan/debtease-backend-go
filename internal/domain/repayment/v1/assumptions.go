package repayment_v1

const EngineVersion = "v1"

var Assumptions = []string{
	"Fixed interest rate for entire tenure",
	"Monthly compounding",
	"EMI-based repayment",
	"Prepayments directly reduce principal",
	"No bank penalties considered",
	"No floating rates",
	"No historical payment reconstruction",
	"Outstanding principal is user-provided truth",
}
