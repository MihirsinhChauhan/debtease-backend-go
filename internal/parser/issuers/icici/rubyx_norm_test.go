package icici

import (
	"os"
	"path/filepath"
	"testing"

	"DebtEase/internal/parser"
	pdfextract "DebtEase/internal/parser/pdf"

	"github.com/shopspring/decimal"
)

const fixtureRubyxNormHeader = `ICICI Bank Credit Card Rubyx
Á5,520.00Á1,10,237.29July 18, 2026August 5, 2026
Statement period : June 19, 2026 to July 18, 2026
4501XXXXXXXX1004
Credit Limit (Including cash)Available Credit (Including cash)Cash LimitAvailable CashÁ3,00,000.00Á1,89,762.71Á20,000.00Á0.00
`

func TestParseRubyxNormHeader(t *testing.T) {
	p := New()
	stmt, err := p.Parse(fixtureRubyxNormHeader, parser.ParseOptions{})
	if err != nil {
		t.Fatalf("parse failed: %v missing=%v", err, stmt.MissingRequired)
	}
	assertDecimalEqual(t, "minimum due", stmt.MinimumDue, "5520.00")
	assertDecimalEqual(t, "statement amount", stmt.StatementAmount, "110237.29")
	assertDecimalEqual(t, "credit limit", stmt.CreditLimit, "300000.00")
	if stmt.Last4 != "1004" {
		t.Fatalf("expected last4 1004, got %s", stmt.Last4)
	}
	if stmt.StatementDate == nil || stmt.StatementDate.Day() != 18 || stmt.StatementDate.Month() != 7 {
		t.Fatalf("unexpected statement date: %v", stmt.StatementDate)
	}
	if stmt.DueDate == nil || stmt.DueDate.Day() != 5 || stmt.DueDate.Month() != 8 {
		t.Fatalf("unexpected due date: %v", stmt.DueDate)
	}
}

func TestParseRealRubyxNormPDF_AllAmounts(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "test_docs", "4501XXXXXXXX1004_589675_Retail_Rubyx_NORM.pdf")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	text, err := pdfextract.ExtractText(data, "mihi0407")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	stmt, err := New().Parse(text, parser.ParseOptions{})
	if err != nil {
		t.Fatalf("parse failed: %v missing=%v", err, stmt.MissingRequired)
	}
	logDec := func(label string, d *decimal.Decimal) {
		if d != nil {
			t.Logf("%s=%s", label, d.String())
			if d.CoefficientInt64() > 9_999_999_999 {
				t.Fatalf("%s overflows NUMERIC(12,2): %s", label, d.String())
			}
		}
	}
	logDec("StatementAmount", stmt.StatementAmount)
	logDec("MinimumDue", stmt.MinimumDue)
	logDec("PreviousBalance", stmt.PreviousBalance)
	logDec("PaymentsReceived", stmt.PaymentsReceived)
	logDec("NewBalance", stmt.NewBalance)
	logDec("CreditLimit", stmt.CreditLimit)
	logDec("AvailableCredit", stmt.AvailableCredit)
}

func TestParseRealRubyxNormPDF(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "test_docs", "4501XXXXXXXX1004_589675_Retail_Rubyx_NORM.pdf")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	text, err := pdfextract.ExtractText(data, "mihi0407")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	p := New()
	stmt, err := p.Parse(text, parser.ParseOptions{})
	if err != nil {
		t.Fatalf("parse failed: %v missing=%v", err, stmt.MissingRequired)
	}
	assertDecimalEqual(t, "minimum due", stmt.MinimumDue, "5520.00")
	assertDecimalEqual(t, "statement amount", stmt.StatementAmount, "110237.29")
	if stmt.Last4 != "1004" {
		t.Fatalf("expected last4 1004, got %s", stmt.Last4)
	}
}

func assertDecimalEqual(t *testing.T, label string, actual *decimal.Decimal, want string) {
	t.Helper()
	if actual == nil {
		t.Fatalf("%s is nil", label)
	}
	expected := decimalFromString(want)
	if !actual.Equal(expected) {
		t.Fatalf("%s: got %s want %s", label, actual.String(), expected.String())
	}
}
