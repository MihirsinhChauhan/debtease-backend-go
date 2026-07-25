package icici

import (
	"testing"

	"DebtEase/internal/parser"

	"github.com/shopspring/decimal"
)

const fixtureRubyxText = `ICICI Bank
Credit Card Statement
Rubyx Dual Credit Card
Visa / RuPay

Statement Date: 12-Jul-2026
Payment Due Date: 02-Aug-2026
Statement Period: 13-Jun-2026 to 12-Jul-2026

Total Amount Due: 24,580.00
Minimum Amount Due: 1,240.00
Previous Balance: 22,100.00
Payments & Other Credits: 5,000.00
New Balance: 24,580.00

Credit Limit: 1,20,000.00
Available Credit Limit: 95,420.00

Card Number: 4375 XXXX XXXX 1234
Secondary Card: 6074 XXXX XXXX 5678
`

func TestDetectConfidence_Rubyx(t *testing.T) {
	p := New()
	score := p.DetectConfidence(fixtureRubyxText)
	if score < 60 {
		t.Fatalf("expected confidence >= 60, got %d", score)
	}
}

func TestParseRubyx(t *testing.T) {
	p := New()
	stmt, err := p.Parse(fixtureRubyxText, parser.ParseOptions{})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if stmt.ParserVariant != variantRubyx {
		t.Fatalf("expected variant rubyx, got %s", stmt.ParserVariant)
	}
	if stmt.Last4 != "1234" {
		t.Fatalf("expected last4 1234, got %s", stmt.Last4)
	}
	if stmt.StatementAmount == nil || !stmt.StatementAmount.Equal(decimalFromString("24580.00")) {
		t.Fatalf("unexpected statement amount: %v", stmt.StatementAmount)
	}
	if stmt.MinimumDue == nil || !stmt.MinimumDue.Equal(decimalFromString("1240.00")) {
		t.Fatalf("unexpected minimum due: %v", stmt.MinimumDue)
	}
	cards, ok := stmt.Metadata["card_numbers"].([]string)
	if !ok || len(cards) < 2 {
		t.Fatalf("expected multiple card numbers in metadata, got %v", stmt.Metadata["card_numbers"])
	}
}

func decimalFromString(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}
