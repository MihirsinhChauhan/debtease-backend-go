package icici

import (
	"regexp"
	"strings"
	"time"

	"DebtEase/internal/parser"

	"github.com/shopspring/decimal"
)

// ICICI Rubyx NORM statements embed summary values without labels, e.g.:
// Á5,520.00Á1,10,237.29July 18, 2026August 5, 2026
var rubyxSummaryHeaderRe = regexp.MustCompile(
	`(?i)[Á₹]([\d,]+\.\d{2})[Á₹]([\d,]+\.\d{2})` +
		`([A-Za-z]+\s+\d{1,2},\s+\d{4})` +
		`([A-Za-z]+\s+\d{1,2},\s+\d{4})`,
)

// Previous BalancePurchases / ChargesCash AdvancesPayments / CreditsÁ0.00Á1,14,315.29Á0.00Á4,078.00
var rubyxBalanceBandRe = regexp.MustCompile(
	`(?i)Previous Balance\s*Purchases\s*/?\s*Charges\s*Cash Advances\s*Payments\s*/?\s*Credits` +
		`[Á₹]([\d,]+\.\d{2})[Á₹]([\d,]+\.\d{2})[Á₹]([\d,]+\.\d{2})[Á₹]([\d,]+\.\d{2})`,
)

func parseRubyx(text string) (*parser.NormalizedStatement, error) {
	stmt := &parser.NormalizedStatement{
		Issuer:        issuerID,
		ParserVariant: variantRubyx,
		Metadata:      map[string]interface{}{},
	}

	parseRubyxSummaryHeader(text, stmt)

	if v := FindLabelValue(text, "Statement Date"); v != "" {
		if t, ok := ParseDate(v); ok {
			stmt.StatementDate = &t
		}
	}
	if stmt.StatementDate == nil {
		if v := FindLabelValue(text, "Statement date"); v != "" {
			if t, ok := ParseDate(v); ok {
				stmt.StatementDate = &t
			}
		}
	}

	if v := FindLabelValue(text, "Payment Due Date"); v != "" {
		if t, ok := ParseDate(v); ok && isLikelyDueDate(t, stmt.StatementDate) {
			stmt.DueDate = &t
		}
	}
	if stmt.DueDate == nil {
		if v := FindLabelValue(text, "Due Date"); v != "" {
			if t, ok := ParseDate(v); ok && isLikelyDueDate(t, stmt.StatementDate) {
				stmt.DueDate = &t
			}
		}
	}

	periodRe := regexp.MustCompile(`(?i)statement\s+period\s*:?\s*([0-9A-Za-z\-/ ,]+?)\s+to\s+([0-9A-Za-z\-/ ,]+)`)
	if m := periodRe.FindStringSubmatch(text); len(m) == 3 {
		if t, ok := ParseDate(strings.TrimSpace(m[1])); ok {
			stmt.StatementStart = &t
		}
		if t, ok := ParseDate(strings.TrimSpace(m[2])); ok {
			stmt.StatementEnd = &t
			if stmt.StatementDate == nil {
				stmt.StatementDate = &t
			}
		}
	}

	if stmt.StatementAmount == nil {
		setAmountPtr(text, "Total Amount Due", &stmt.StatementAmount)
	}
	if stmt.StatementAmount == nil {
		setAmountPtr(text, "Statement Amount", &stmt.StatementAmount)
	}
	if stmt.MinimumDue == nil {
		setAmountPtr(text, "Minimum Amount Due", &stmt.MinimumDue)
	}

	parseRubyxCreditSummary(text, stmt)
	parseRubyxBalanceBand(text, stmt)

	setAmountPtr(text, "Previous Balance", &stmt.PreviousBalance)
	setAmountPtr(text, "Payments & Other Credits", &stmt.PaymentsReceived)
	if stmt.PaymentsReceived == nil {
		setAmountPtr(text, "Payments Received", &stmt.PaymentsReceived)
	}
	if stmt.NewBalance == nil {
		setAmountPtr(text, "New Balance", &stmt.NewBalance)
	}
	if stmt.NewBalance == nil {
		setAmountPtr(text, "Closing Balance", &stmt.NewBalance)
	}
	if stmt.CreditLimit == nil {
		setAmountPtr(text, "Credit Limit", &stmt.CreditLimit)
	}
	if stmt.AvailableCredit == nil {
		setAmountPtr(text, "Available Credit Limit", &stmt.AvailableCredit)
	}
	if stmt.AvailableCredit == nil {
		setAmountPtr(text, "Available Credit", &stmt.AvailableCredit)
	}

	last4s := CollectLast4s(text)
	if len(last4s) > 0 {
		stmt.Last4 = last4s[0]
		if len(last4s) > 1 {
			stmt.Metadata["card_numbers"] = last4s
			if networks := detectNetworks(text); len(networks) > 0 {
				stmt.Metadata["networks"] = networks
			}
		}
	}

	if stmt.NewBalance == nil && stmt.StatementAmount != nil {
		bal := *stmt.StatementAmount
		stmt.NewBalance = &bal
	}

	validateRequired(stmt)
	if len(stmt.MissingRequired) > 0 {
		return stmt, parser.ErrParseIncomplete
	}
	return stmt, nil
}

func parseRubyxSummaryHeader(text string, stmt *parser.NormalizedStatement) {
	m := rubyxSummaryHeaderRe.FindStringSubmatch(text)
	if len(m) != 5 {
		return
	}
	if amt, ok := ParseAmount(m[1]); ok {
		min := amt
		stmt.MinimumDue = &min
	}
	if amt, ok := ParseAmount(m[2]); ok {
		total := amt
		stmt.StatementAmount = &total
		stmt.NewBalance = &total
	}
	if t, ok := ParseDate(m[3]); ok {
		stmt.StatementDate = &t
	}
	if t, ok := ParseDate(m[4]); ok {
		stmt.DueDate = &t
	}
}

// Credit Limit (Including cash)Available Credit ...Á3,00,000.00Á1,89,762.71
var rubyxCreditBandRe = regexp.MustCompile(
	`(?i)Credit Limit \(Including cash\).*?[Á₹]([\d,]+\.\d{2})[Á₹]([\d,]+\.\d{2})`,
)

func parseRubyxCreditSummary(text string, stmt *parser.NormalizedStatement) {
	m := rubyxCreditBandRe.FindStringSubmatch(text)
	if len(m) != 3 {
		return
	}
	if stmt.CreditLimit == nil {
		if amt, ok := ParseAmount(m[1]); ok {
			limit := amt
			stmt.CreditLimit = &limit
		}
	}
	if stmt.AvailableCredit == nil {
		if amt, ok := ParseAmount(m[2]); ok {
			avail := amt
			stmt.AvailableCredit = &avail
		}
	}
}

func parseRubyxBalanceBand(text string, stmt *parser.NormalizedStatement) {
	m := rubyxBalanceBandRe.FindStringSubmatch(text)
	if len(m) != 5 {
		return
	}
	if stmt.PreviousBalance == nil {
		if amt, ok := ParseAmount(m[1]); ok {
			prev := amt
			stmt.PreviousBalance = &prev
		}
	}
	// m[2] is purchases/charges — not stored on NormalizedStatement today.
	if stmt.PaymentsReceived == nil {
		if amt, ok := ParseAmount(m[4]); ok {
			payments := amt
			stmt.PaymentsReceived = &payments
		}
	}
}

func isLikelyDueDate(candidate time.Time, statementDate *time.Time) bool {
	if statementDate == nil {
		return true
	}
	return candidate.After(*statementDate)
}

var maxStatementFieldAmount = decimal.NewFromInt(99_999_999)

func setAmountPtr(text, label string, target **decimal.Decimal) {
	if *target != nil {
		return
	}
	if v := FindLabelValue(text, label); v != "" {
		if amt, ok := ParseAmount(v); ok && amt.LessThanOrEqual(maxStatementFieldAmount) {
			*target = &amt
		}
	}
}

func detectNetworks(text string) []string {
	upper := strings.ToUpper(text)
	var nets []string
	if strings.Contains(upper, "VISA") {
		nets = append(nets, "VISA")
	}
	if strings.Contains(upper, "RUPAY") || strings.Contains(upper, "RU PAY") {
		nets = append(nets, "RUPAY")
	}
	return nets
}

func validateRequired(stmt *parser.NormalizedStatement) {
	if stmt.StatementDate == nil {
		stmt.MissingRequired = append(stmt.MissingRequired, "statement_date")
	}
	if stmt.DueDate == nil {
		stmt.MissingRequired = append(stmt.MissingRequired, "due_date")
	}
	if stmt.StatementAmount == nil && stmt.NewBalance == nil {
		stmt.MissingRequired = append(stmt.MissingRequired, "statement_amount")
	}
	if stmt.MinimumDue == nil {
		stmt.MissingRequired = append(stmt.MissingRequired, "minimum_due")
	}
}
