package parser

import "errors"

var (
	ErrPasswordRequired = errors.New("pdf password required")
	ErrPasswordInvalid  = errors.New("pdf password invalid")
	ErrIssuerUnknown    = errors.New("issuer unknown")
	ErrParseIncomplete  = errors.New("parse incomplete")
	ErrInvalidFile      = errors.New("invalid file")
)

// FailureCode maps parser errors to API/database codes.
func FailureCode(err error) string {
	switch {
	case errors.Is(err, ErrPasswordRequired):
		return "password_required"
	case errors.Is(err, ErrPasswordInvalid):
		return "password_invalid"
	case errors.Is(err, ErrIssuerUnknown):
		return "issuer_unknown"
	case errors.Is(err, ErrParseIncomplete):
		return "parse_incomplete"
	case errors.Is(err, ErrInvalidFile):
		return "invalid_file"
	default:
		return "internal_error"
	}
}

// APIErrorCode returns the uppercase client-facing code.
func APIErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrPasswordRequired):
		return "PDF_PASSWORD_REQUIRED"
	case errors.Is(err, ErrPasswordInvalid):
		return "PDF_PASSWORD_INVALID"
	case errors.Is(err, ErrIssuerUnknown):
		return "ISSUER_UNKNOWN"
	case errors.Is(err, ErrParseIncomplete):
		return "PARSE_INCOMPLETE"
	case errors.Is(err, ErrInvalidFile):
		return "INVALID_FILE"
	default:
		return "INTERNAL_ERROR"
	}
}
