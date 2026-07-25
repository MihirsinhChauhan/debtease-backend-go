package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/parser"
	"DebtEase/internal/service"
	"DebtEase/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func statementService(cfg *api.Config) *service.StatementService {
	return service.NewStatementService(cfg.DB, cfg.StatementStorage)
}

func HandleUploadStatement(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		debtID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid debt id")
			return
		}

		if err := r.ParseMultipartForm(21 << 20); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		password := r.FormValue("password")
		mime := header.Header.Get("Content-Type")
		if mime == "" {
			mime = "application/pdf"
		}

		svc := statementService(cfg)
		result, err := svc.UploadAndParse(r.Context(), service.UploadInput{
			UserID:   userID,
			DebtID:   debtID,
			Filename: header.Filename,
			Mime:     mime,
			Reader:   file,
			Password: password,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			if errors.Is(err, parser.ErrInvalidFile) {
				utils.RespondWithErrorCode(w, http.StatusBadRequest, parser.APIErrorCode(err), "Invalid PDF file")
				return
			}
			logger.Logger.Errorw("UploadAndParse failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to upload statement")
			return
		}

		if result.ParseErr != nil {
			status := http.StatusUnprocessableEntity
			utils.RespondWithJSON(w, status, toStatementUploadResponse(result))
			return
		}
		utils.RespondWithJSON(w, http.StatusCreated, toStatementUploadResponse(result))
	}
}

func HandleRetryStatementParse(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		uploadID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid upload id")
			return
		}

		var req models.RetryStatementParseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		svc := statementService(cfg)
		result, err := svc.RetryParse(r.Context(), userID, uploadID, req.Password)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithError(w, http.StatusNotFound, "Upload not found")
				return
			}
			logger.Logger.Errorw("RetryParse failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to parse statement")
			return
		}

		if result.ParseErr != nil {
			utils.RespondWithJSON(w, http.StatusUnprocessableEntity, toStatementUploadResponse(result))
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, toStatementUploadResponse(result))
	}
}

func HandleListStatements(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		debtID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid debt id")
			return
		}

		svc := statementService(cfg)
		items, err := svc.ListStatements(r.Context(), userID, debtID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			if strings.Contains(err.Error(), "not a credit card") {
				utils.RespondWithError(w, http.StatusBadRequest, "Debt is not a credit card")
				return
			}
			logger.Logger.Errorw("ListStatements failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list statements")
			return
		}

		resp := make([]models.CreditCardStatementResponse, 0, len(items))
		for _, item := range items {
			resp = append(resp, toStatementResponse(item))
		}
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func HandleGetStatement(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		statementID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid statement id")
			return
		}

		svc := statementService(cfg)
		stmt, err := svc.GetStatement(r.Context(), userID, statementID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithError(w, http.StatusNotFound, "Statement not found")
				return
			}
			logger.Logger.Errorw("GetStatement failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get statement")
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, toStatementResponse(stmt))
	}
}

func HandleDownloadStatementPDF(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		uploadID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid upload id")
			return
		}

		svc := statementService(cfg)
		rc, filename, err := svc.OpenPDF(r.Context(), userID, uploadID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithError(w, http.StatusNotFound, "Upload not found")
				return
			}
			logger.Logger.Errorw("OpenPDF failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to download PDF")
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		_, _ = io.Copy(w, rc)
	}
}

func toStatementUploadResponse(result *service.UploadResult) models.StatementUploadResponse {
	u := result.Upload
	resp := models.StatementUploadResponse{
		ID:               u.ID.String(),
		DebtID:           u.DebtID.String(),
		OriginalFilename: u.OriginalFilename,
		Status:           string(u.Status),
		UploadedAt:       u.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.StatementID.Valid {
		sid := u.StatementID.UUID.String()
		resp.StatementID = &sid
	}
	if u.FailureCode.Valid {
		code := string(u.FailureCode.StatementFailureCode)
		resp.FailureCode = &code
		apiCode := failureCodeToAPI(code)
		resp.Code = &apiCode
	}
	if u.ErrorMessage.Valid {
		msg := u.ErrorMessage.String
		resp.ErrorMessage = &msg
	}
	if result.Statement != nil {
		stmt := toStatementResponse(*result.Statement)
		resp.Statement = &stmt
	}
	return resp
}

func failureCodeToAPI(code string) string {
	switch code {
	case "password_required":
		return "PDF_PASSWORD_REQUIRED"
	case "password_invalid":
		return "PDF_PASSWORD_INVALID"
	case "issuer_unknown":
		return "ISSUER_UNKNOWN"
	case "parse_incomplete":
		return "PARSE_INCOMPLETE"
	case "invalid_file":
		return "INVALID_FILE"
	default:
		return "INTERNAL_ERROR"
	}
}

func toStatementResponse(s database.CreditCardStatement) models.CreditCardStatementResponse {
	resp := models.CreditCardStatementResponse{
		ID:        s.ID.String(),
		DebtID:    s.DebtID.String(),
		Source:    string(s.Source),
		UploadID:  s.UploadID.String(),
		Metadata:  s.Metadata,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if s.StatementStart.Valid {
		v := service.FormatDate(s.StatementStart.Time)
		resp.StatementStart = &v
	}
	if s.StatementEnd.Valid {
		v := service.FormatDate(s.StatementEnd.Time)
		resp.StatementEnd = &v
	}
	if s.StatementDate.Valid {
		v := service.FormatDate(s.StatementDate.Time)
		resp.StatementDate = &v
	}
	if s.DueDate.Valid {
		v := service.FormatDate(s.DueDate.Time)
		resp.DueDate = &v
	}
	resp.StatementAmount = nullDecimalPtr(s.StatementAmount)
	resp.MinimumDue = nullDecimalPtr(s.MinimumDue)
	resp.PreviousBalance = nullDecimalPtr(s.PreviousBalance)
	resp.PaymentsReceived = nullDecimalPtr(s.PaymentsReceived)
	resp.NewBalance = nullDecimalPtr(s.NewBalance)
	resp.CreditLimit = nullDecimalPtr(s.CreditLimit)
	resp.AvailableCredit = nullDecimalPtr(s.AvailableCredit)
	if s.IssuerDetected.Valid {
		v := s.IssuerDetected.String
		resp.IssuerDetected = &v
	}
	if s.ParserVariant.Valid {
		v := s.ParserVariant.String
		resp.ParserVariant = &v
	}
	return resp
}

func nullDecimalPtr(v decimal.NullDecimal) *decimal.Decimal {
	if !v.Valid {
		return nil
	}
	d := v.Decimal
	return &d
}

// loadCreditCardResponse loads nested credit card for debt responses.
func loadCreditCardResponse(ctx context.Context, q *database.Queries, debt database.Debt) *models.CreditCardResponse {
	if debt.Type != database.DebtTypeCreditCard {
		return nil
	}
	cc, err := q.GetCreditCardByDebtID(ctx, debt.ID)
	if err != nil {
		return nil
	}
	return creditCardToResponse(cc)
}

func creditCardToResponse(cc database.CreditCard) *models.CreditCardResponse {
	var lastStmt, due *time.Time
	if cc.LastStatementDate.Valid {
		t := cc.LastStatementDate.Time
		lastStmt = &t
	}
	if cc.PaymentDueDate.Valid {
		t := cc.PaymentDueDate.Time
		due = &t
	}
	var minDue *decimal.Decimal
	if !cc.MinimumDue.IsZero() {
		d := cc.MinimumDue
		minDue = &d
	}
	var last4 *string
	if cc.Last4.Valid {
		s := cc.Last4.String
		last4 = &s
	}
	var limit *decimal.Decimal
	if cc.CreditLimit.Valid {
		d := cc.CreditLimit.Decimal
		limit = &d
	}
	var issuer *string
	if cc.Issuer.Valid {
		s := cc.Issuer.String
		issuer = &s
	}
	return &models.CreditCardResponse{
		BillingDay:        cc.BillingDay,
		GracePeriodDays:   cc.GracePeriodDays,
		MinPaymentPercent: cc.MinPaymentPercent,
		LastStatementDate: lastStmt,
		PaymentDueDate:    due,
		BilledBalance:     cc.BilledBalance,
		UnbilledBalance:   cc.UnbilledBalance,
		InGracePeriod:     cc.InGracePeriod,
		MinimumDue:        minDue,
		Last4:             last4,
		CreditLimit:       limit,
		Issuer:            issuer,
	}
}
