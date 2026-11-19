package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"DebtEase/internal/database"
	"DebtEase/internal/parser"
	pdfextract "DebtEase/internal/parser/pdf"
	"DebtEase/internal/parser/issuers/icici"
	"DebtEase/internal/storage"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const maxStatementPDFBytes = 20 << 20 // 20 MB

type StatementService struct {
	db       *database.Queries
	store    storage.StatementStorage
	registry *parser.Registry
}

func NewStatementService(db *database.Queries, store storage.StatementStorage) *StatementService {
	reg := parser.NewRegistry(icici.New())
	return &StatementService{db: db, store: store, registry: reg}
}

type UploadInput struct {
	UserID           uuid.UUID
	DebtID           uuid.UUID
	Filename         string
	Mime             string
	Reader           io.Reader
	Password         string
}

type UploadResult struct {
	Upload    database.StatementUpload
	Statement *database.CreditCardStatement
	ParseErr  error
}

func (s *StatementService) UploadAndParse(ctx context.Context, in UploadInput) (*UploadResult, error) {
	if err := s.validateCreditCardDebt(ctx, in.UserID, in.DebtID); err != nil {
		return nil, err
	}
	if in.Mime != "" && in.Mime != "application/pdf" && in.Mime != "application/octet-stream" {
		return nil, parser.ErrInvalidFile
	}

	data, err := pdfextract.ReadAll(in.Reader, maxStatementPDFBytes)
	if err != nil {
		return nil, err
	}
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		return nil, parser.ErrInvalidFile
	}

	uploadID := uuid.New()
	key := storage.ObjectKey(in.UserID.String(), in.DebtID.String(), uploadID.String())
	if err := s.store.Put(ctx, key, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("store pdf: %w", err)
	}

	sum := sha256.Sum256(data)
	upload, err := s.db.CreateStatementUpload(ctx, database.CreateStatementUploadParams{
		DebtID:           in.DebtID,
		StoragePath:      key,
		OriginalFilename: in.Filename,
		Checksum:         hex.EncodeToString(sum[:]),
		Mime:             "application/pdf",
		Status:           database.StatementUploadStatusUploaded,
	})
	if err != nil {
		return nil, err
	}

	return s.parseExistingUpload(ctx, in.UserID, upload, data, in.Password)
}

func (s *StatementService) RetryParse(ctx context.Context, userID, uploadID uuid.UUID, password string) (*UploadResult, error) {
	upload, err := s.db.GetStatementUploadByID(ctx, database.GetStatementUploadByIDParams{
		ID:     uploadID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	rc, err := s.store.Get(ctx, upload.StoragePath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return s.parseExistingUpload(ctx, userID, upload, data, password)
}

func (s *StatementService) parseExistingUpload(ctx context.Context, userID uuid.UUID, upload database.StatementUpload, data []byte, password string) (*UploadResult, error) {
	_, _ = s.db.UpdateStatementUploadStatus(ctx, database.UpdateStatementUploadStatusParams{
		ID:          upload.ID,
		Status:      database.StatementUploadStatusParsing,
		FailureCode: database.NullStatementFailureCode{},
		ErrorMessage: sql.NullString{},
		StatementID: uuid.NullUUID{},
	})

	text, extractErr := pdfextract.ExtractText(data, password)
	if extractErr != nil {
		return s.failUpload(ctx, upload, extractErr)
	}

	normalized, parseErr := s.registry.Parse(text, parser.ParseOptions{Password: password})
	if parseErr != nil {
		res, err := s.failUpload(ctx, upload, parseErr)
		if err != nil {
			return nil, err
		}
		if normalized != nil && errors.Is(parseErr, parser.ErrParseIncomplete) {
			if stmtRow, persistErr := s.persistStatement(ctx, upload.DebtID, upload.ID, normalized); persistErr == nil {
				res.Statement = stmtRow
				_ = s.syncCreditCardFromParse(ctx, upload.DebtID, normalized)
			}
		}
		res.ParseErr = parseErr
		return res, nil
	}

	stmtRow, err := s.persistStatement(ctx, upload.DebtID, upload.ID, normalized)
	if err != nil {
		return nil, err
	}
	if err := s.syncCreditCardFromParse(ctx, upload.DebtID, normalized); err != nil {
		return nil, err
	}

	updated, err := s.db.UpdateStatementUploadStatus(ctx, database.UpdateStatementUploadStatusParams{
		ID:           upload.ID,
		Status:       database.StatementUploadStatusParsed,
		FailureCode:  database.NullStatementFailureCode{},
		ErrorMessage: sql.NullString{},
		StatementID:  uuid.NullUUID{UUID: stmtRow.ID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &UploadResult{Upload: updated, Statement: stmtRow}, nil
}

func (s *StatementService) failUpload(ctx context.Context, upload database.StatementUpload, parseErr error) (*UploadResult, error) {
	code := database.StatementFailureCode(parser.FailureCode(parseErr))
	updated, err := s.db.UpdateStatementUploadStatus(ctx, database.UpdateStatementUploadStatusParams{
		ID:     upload.ID,
		Status: database.StatementUploadStatusFailed,
		FailureCode: database.NullStatementFailureCode{
			StatementFailureCode: code,
			Valid:                true,
		},
		ErrorMessage: sql.NullString{String: parseErr.Error(), Valid: true},
		StatementID:  uuid.NullUUID{},
	})
	if err != nil {
		return nil, err
	}
	return &UploadResult{Upload: updated, ParseErr: parseErr}, nil
}

func (s *StatementService) persistStatement(ctx context.Context, debtID, uploadID uuid.UUID, n *parser.NormalizedStatement) (*database.CreditCardStatement, error) {
	meta, _ := json.Marshal(n.Metadata)
	if meta == nil {
		meta = []byte("{}")
	}
	params := database.CreateCreditCardStatementParams{
		DebtID:   debtID,
		Source:   database.StatementSourceMANUALPDF,
		UploadID: uploadID,
		Metadata: meta,
	}
	if n.StatementStart != nil {
		params.StatementStart = sql.NullTime{Time: *n.StatementStart, Valid: true}
	}
	if n.StatementEnd != nil {
		params.StatementEnd = sql.NullTime{Time: *n.StatementEnd, Valid: true}
	}
	if n.StatementDate != nil {
		params.StatementDate = sql.NullTime{Time: *n.StatementDate, Valid: true}
	}
	if n.DueDate != nil {
		params.DueDate = sql.NullTime{Time: *n.DueDate, Valid: true}
	}
	params.StatementAmount = nullDecimal(n.StatementAmount)
	params.MinimumDue = nullDecimal(n.MinimumDue)
	params.PreviousBalance = nullDecimal(n.PreviousBalance)
	params.PaymentsReceived = nullDecimal(n.PaymentsReceived)
	params.NewBalance = nullDecimal(n.NewBalance)
	params.CreditLimit = nullDecimal(n.CreditLimit)
	params.AvailableCredit = nullDecimal(n.AvailableCredit)
	if n.Issuer != "" {
		params.IssuerDetected = sql.NullString{String: n.Issuer, Valid: true}
	}
	if n.ParserVariant != "" {
		params.ParserVariant = sql.NullString{String: n.ParserVariant, Valid: true}
	}
	row, err := s.db.CreateCreditCardStatement(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *StatementService) syncCreditCardFromParse(ctx context.Context, debtID uuid.UUID, n *parser.NormalizedStatement) error {
	var minDue decimal.Decimal
	if n.MinimumDue != nil {
		minDue = *n.MinimumDue
	}
	var lastStmt, due sql.NullTime
	if n.StatementDate != nil {
		lastStmt = sql.NullTime{Time: *n.StatementDate, Valid: true}
	}
	if n.DueDate != nil {
		due = sql.NullTime{Time: *n.DueDate, Valid: true}
	}
	var last4 sql.NullString
	if n.Last4 != "" {
		last4 = sql.NullString{String: n.Last4, Valid: true}
	}
	var limit decimal.NullDecimal
	if n.CreditLimit != nil {
		limit = decimal.NullDecimal{Decimal: *n.CreditLimit, Valid: true}
	}
	var issuer sql.NullString
	if n.Issuer != "" {
		issuer = sql.NullString{String: n.Issuer, Valid: true}
	}
	if err := s.db.UpdateCreditCardFromStatement(ctx, database.UpdateCreditCardFromStatementParams{
		DebtID:            debtID,
		MinimumDue:        minDue,
		LastStatementDate: lastStmt,
		PaymentDueDate:    due,
		Last4:             last4,
		CreditLimit:       limit,
		Issuer:            issuer,
	}); err != nil {
		return err
	}

	var principalUpdate decimal.Decimal
	if n.CreditLimit != nil && n.CreditLimit.IsPositive() {
		principalUpdate = *n.CreditLimit
	}
	if minDue.IsPositive() || principalUpdate.IsPositive() || due.Valid {
		if err := s.db.UpdateDebtFromStatementParse(ctx, database.UpdateDebtFromStatementParseParams{
			ID:               debtID,
			MinPayment:       minDue,
			PrincipalDecimal: principalUpdate,
			DueDate:          due,
		}); err != nil {
			return err
		}
	}

	balance := n.NewBalance
	if balance == nil {
		balance = n.StatementAmount
	}
	if balance != nil {
		return s.db.UpdateDebtBalance(ctx, database.UpdateDebtBalanceParams{
			ID:                 debtID,
			OutstandingBalance: *balance,
			IsOverdue:          sql.NullBool{},
		})
	}
	return nil
}

func (s *StatementService) validateCreditCardDebt(ctx context.Context, userID, debtID uuid.UUID) error {
	debt, err := s.db.GetDebtByID(ctx, database.GetDebtByIDParams{ID: debtID, UserID: userID})
	if err != nil {
		return err
	}
	if debt.Type != database.DebtTypeCreditCard {
		return fmt.Errorf("debt is not a credit card")
	}
	return nil
}

func nullDecimal(v *decimal.Decimal) decimal.NullDecimal {
	if v == nil {
		return decimal.NullDecimal{}
	}
	return decimal.NullDecimal{Decimal: *v, Valid: true}
}

func (s *StatementService) ListStatements(ctx context.Context, userID, debtID uuid.UUID) ([]database.CreditCardStatement, error) {
	if err := s.validateCreditCardDebt(ctx, userID, debtID); err != nil {
		return nil, err
	}
	return s.db.ListCreditCardStatementsByDebtID(ctx, database.ListCreditCardStatementsByDebtIDParams{
		DebtID: debtID,
		UserID: userID,
	})
}

func (s *StatementService) GetStatement(ctx context.Context, userID, statementID uuid.UUID) (database.CreditCardStatement, error) {
	return s.db.GetCreditCardStatementByID(ctx, database.GetCreditCardStatementByIDParams{
		ID:     statementID,
		UserID: userID,
	})
}

func (s *StatementService) GetUpload(ctx context.Context, userID, uploadID uuid.UUID) (database.StatementUpload, error) {
	return s.db.GetStatementUploadByID(ctx, database.GetStatementUploadByIDParams{
		ID:     uploadID,
		UserID: userID,
	})
}

func (s *StatementService) OpenPDF(ctx context.Context, userID, uploadID uuid.UUID) (io.ReadCloser, string, error) {
	upload, err := s.GetUpload(ctx, userID, uploadID)
	if err != nil {
		return nil, "", err
	}
	rc, err := s.store.Get(ctx, upload.StoragePath)
	if err != nil {
		return nil, "", err
	}
	return rc, upload.OriginalFilename, nil
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
