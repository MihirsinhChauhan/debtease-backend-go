package icici

import (
	"strings"

	"DebtEase/internal/parser"
)

const (
	issuerID       = "ICICI"
	variantRubyx   = "rubyx"
	detectICICI    = 40
	detectRubyx    = 30
	detectStmt     = 20
)

// Parser implements ICICI statement parsing with variant routing.
type Parser struct{}

func New() *Parser { return &Parser{} }

func (p *Parser) IssuerID() string { return issuerID }

func (p *Parser) DetectConfidence(text string) int {
	upper := strings.ToUpper(text)
	score := 0
	if strings.Contains(upper, "ICICI") {
		score += detectICICI
	}
	if strings.Contains(upper, "RUBYX") {
		score += detectRubyx
	}
	if strings.Contains(upper, "CREDIT CARD") && strings.Contains(upper, "STATEMENT") {
		score += detectStmt
	}
	return score
}

func (p *Parser) Parse(text string, _ parser.ParseOptions) (*parser.NormalizedStatement, error) {
	upper := strings.ToUpper(text)
	if strings.Contains(upper, "RUBYX") {
		return parseRubyx(text)
	}
	return parseRubyx(text)
}
