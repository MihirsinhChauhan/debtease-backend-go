package parser

import (
	"time"

	"github.com/shopspring/decimal"
)

// ParseOptions carries per-request parse settings. Password is never persisted.
type ParseOptions struct {
	Password string
}

// NormalizedStatement is issuer-agnostic output from any parser.
type NormalizedStatement struct {
	Issuer            string
	ParserVariant     string
	Last4             string
	StatementStart    *time.Time
	StatementEnd      *time.Time
	StatementDate     *time.Time
	DueDate           *time.Time
	StatementAmount   *decimal.Decimal
	MinimumDue        *decimal.Decimal
	PreviousBalance   *decimal.Decimal
	PaymentsReceived  *decimal.Decimal
	NewBalance        *decimal.Decimal
	CreditLimit       *decimal.Decimal
	AvailableCredit   *decimal.Decimal
	Metadata          map[string]interface{}
	MissingRequired   []string
}

// IssuerParser detects and parses statement text for one bank.
type IssuerParser interface {
	IssuerID() string
	DetectConfidence(text string) int
	Parse(text string, opts ParseOptions) (*NormalizedStatement, error)
}
