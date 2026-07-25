package models

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type StatementUploadResponse struct {
	ID               string  `json:"id"`
	DebtID           string  `json:"debt_id"`
	StatementID      *string `json:"statement_id,omitempty"`
	OriginalFilename string  `json:"original_filename"`
	Status           string  `json:"status"`
	FailureCode      *string `json:"failure_code,omitempty"`
	Code             *string `json:"code,omitempty"`
	ErrorMessage     *string `json:"error_message,omitempty"`
	UploadedAt       string  `json:"uploaded_at"`
	Statement        *CreditCardStatementResponse `json:"statement,omitempty"`
}

type CreditCardStatementResponse struct {
	ID               string          `json:"id"`
	DebtID           string          `json:"debt_id"`
	StatementStart   *string         `json:"statement_start,omitempty"`
	StatementEnd     *string         `json:"statement_end,omitempty"`
	StatementDate    *string         `json:"statement_date,omitempty"`
	DueDate          *string         `json:"due_date,omitempty"`
	StatementAmount  *decimal.Decimal `json:"statement_amount,omitempty"`
	MinimumDue       *decimal.Decimal `json:"minimum_due,omitempty"`
	PreviousBalance  *decimal.Decimal `json:"previous_balance,omitempty"`
	PaymentsReceived *decimal.Decimal `json:"payments_received,omitempty"`
	NewBalance       *decimal.Decimal `json:"new_balance,omitempty"`
	CreditLimit      *decimal.Decimal `json:"credit_limit,omitempty"`
	AvailableCredit  *decimal.Decimal `json:"available_credit,omitempty"`
	Source           string          `json:"source"`
	UploadID         string          `json:"upload_id"`
	IssuerDetected   *string         `json:"issuer_detected,omitempty"`
	ParserVariant    *string         `json:"parser_variant,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	CreatedAt        string          `json:"created_at"`
}

type RetryStatementParseRequest struct {
	Password string `json:"password"`
}

type StatementErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
