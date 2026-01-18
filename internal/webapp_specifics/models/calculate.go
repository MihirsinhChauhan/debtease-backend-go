package models

type LoanInput struct {
	ID               string  `json:"id" validate:"required"`
	Principal        float64 `json:"principal" validate:"required,gt=0"`
	AnnualRate       float64 `json:"annual_rate" validate:"required,gte=0"`
	TenureMonths     int     `json:"tenure_months" validate:"required,gt=0"`
	MoratoriumMonths int     `json:"moratorium_months" validate:"gte=0"`
}

type BaselineCalculateRequest struct {
	EngineVersion string      `json:"engine_version"`
	Loans         []LoanInput `json:"loans" validate:"required,min=1"`
}

type StrategyCalculateRequest struct {
	EngineVersion string      `json:"engine_version" validate:"required"`
	Strategy      string      `json:"strategy" validate:"required,oneof=AVALANCHE SNOWBALL"`
	MonthlyExtra  float64     `json:"monthly_extra" validate:"gte=0"`
	Loans         []LoanInput `json:"loans" validate:"required,min=1"`
}

type PortfolioSummary struct {
	Months         int     `json:"months"`
	TotalPaid      float64 `json:"total_paid"`
	TotalInterest  float64 `json:"total_interest"`
	TotalPrincipal float64 `json:"total_principal"`
}

type EMIRecord struct {
	LoanID         string  `json:"loan_id"`
	Month          int     `json:"month"`
	OpeningBalance float64 `json:"opening_balance"`
	EMI            float64 `json:"emi"`
	Interest       float64 `json:"interest"`
	Principal      float64 `json:"principal"`
	ClosingBalance float64 `json:"closing_balance"`
}

type BaselineCalculateResponse struct {
	EngineVersion string              `json:"engine_version"`
	Summary       PortfolioSummary    `json:"summary"`
	Schedule      map[int][]EMIRecord `json:"schedule"`
}

type Delta struct {
	InterestSaved float64 `json:"interest_saved"`
	MonthsSaved   int     `json:"months_saved"`
}

type StrategyCalculateResponse struct {
	EngineVersion   string              `json:"engine_version"`
	Strategy        string              `json:"strategy"`
	BaselineSummary PortfolioSummary    `json:"baseline_summary"`
	StrategySummary PortfolioSummary    `json:"strategy_summary"`
	Delta           Delta               `json:"delta"`
	Schedule        map[int][]EMIRecord `json:"schedule"`
}
