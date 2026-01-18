package repayment_v1

type StrategyType string

const (
	StrategyBaseline  StrategyType = "BASELINE"
	StrategySnowball  StrategyType = "SNOWBALL"
	StrategyAvalanche StrategyType = "AVALANCHE"
)

type Loan struct {
	ID               string
	Principal        float64
	AnnualRate       float64
	TenureMonths     int
	MoratoriumMonths int
}

type EMIRecord struct {
	LoanID          string
	Month           int
	OpeningBalance  float64
	EMI             float64
	Interest        float64
	Principal       float64
	ClosingBalance  float64
}

type PortfolioSummary struct {
	Months         int
	TotalPaid      float64
	TotalInterest  float64
	TotalPrincipal float64
}

type CalculationInput struct {
	Loans        []Loan
	Strategy     StrategyType
	MonthlyExtra float64
}

type CalculationResult struct {
	Schedule      map[int][]EMIRecord
	Summary       PortfolioSummary
	EngineVersion string
}
