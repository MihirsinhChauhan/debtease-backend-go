package pdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"DebtEase/internal/parser"

	"github.com/ledongthuc/pdf"
	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ExtractText opens a PDF (optionally password-protected) and returns plain text.
func ExtractText(data []byte, password string) (string, error) {
	if len(data) == 0 {
		return "", parser.ErrInvalidFile
	}
	if !isPDF(data) {
		return "", parser.ErrInvalidFile
	}

	plain, err := extractWithLedongthuc(data)
	if err == nil {
		return plain, nil
	}

	if password == "" {
		if needsPassword(data, err) {
			return "", parser.ErrPasswordRequired
		}
		return "", parser.ErrInvalidFile
	}

	// Password supplied — always try decrypt before giving up.
	decrypted, decryptErr := decrypt(data, password)
	if decryptErr != nil {
		if isPasswordError(decryptErr) || needsPassword(data, err) {
			return "", parser.ErrPasswordInvalid
		}
		return "", parser.ErrInvalidFile
	}

	plain, extractErr := extractWithLedongthuc(decrypted)
	if extractErr == nil {
		return plain, nil
	}

	if fallback, fbErr := extractTextFallback(decrypted); fbErr == nil {
		return fallback, nil
	}

	if plain, ctxErr := extractWithPasswordContext(data, password); ctxErr == nil {
		return plain, nil
	}

	if needsPassword(data, err) || needsPassword(decrypted, extractErr) {
		return "", parser.ErrPasswordInvalid
	}
	return "", parser.ErrInvalidFile
}

func isPDF(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "%PDF"
}

func needsPassword(data []byte, extractErr error) bool {
	return isEncrypted(data) || looksEncryptedPDF(data) || likelyNeedsPassword(extractErr)
}

func looksEncryptedPDF(data []byte) bool {
	// Heuristic: encrypted PDFs include an /Encrypt dictionary.
	return bytes.Contains(data, []byte("/Encrypt"))
}

func likelyNeedsPassword(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "encrypt") ||
		strings.Contains(msg, "password") ||
		strings.Contains(msg, "permission") ||
		strings.Contains(msg, "cipher")
}

func isEncrypted(data []byte) bool {
	conf := model.NewDefaultConfiguration()
	ctx, err := pdfcpuapi.ReadContext(bytes.NewReader(data), conf)
	if err != nil {
		return strings.Contains(strings.ToLower(err.Error()), "encrypt")
	}
	return ctx != nil && ctx.Encrypt != nil
}

func decrypt(data []byte, password string) ([]byte, error) {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = password
	conf.OwnerPW = password
	var out bytes.Buffer
	if err := pdfcpuapi.Decrypt(bytes.NewReader(data), &out, conf); err != nil {
		return nil, err
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("decrypt produced empty output")
	}
	return out.Bytes(), nil
}

func isPasswordError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "password") ||
		strings.Contains(msg, "decrypt") ||
		strings.Contains(msg, "permission") ||
		strings.Contains(msg, "cipher")
}

func extractWithLedongthuc(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	total := r.NumPage()
	for i := 1; i <= total; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
		b.WriteByte('\n')
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", fmt.Errorf("empty pdf text")
	}
	return out, nil
}

// extractWithPasswordContext reads encrypted PDFs in-place using pdfcpu + password.
func extractWithPasswordContext(data []byte, password string) (string, error) {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = password
	conf.OwnerPW = password
	ctx, err := pdfcpuapi.ReadContext(bytes.NewReader(data), conf)
	if err != nil {
		return "", err
	}
	return extractTextFromContext(ctx)
}

func extractTextFallback(data []byte) (string, error) {
	conf := model.NewDefaultConfiguration()
	ctx, err := pdfcpuapi.ReadContext(bytes.NewReader(data), conf)
	if err != nil {
		return "", err
	}
	return extractTextFromContext(ctx)
}

func extractTextFromContext(ctx *model.Context) (string, error) {
	var b strings.Builder
	for pageNr := 1; pageNr <= ctx.PageCount; pageNr++ {
		rc, err := pdfcpuapi.ExtractPage(ctx, pageNr)
		if err != nil {
			continue
		}
		buf, err := io.ReadAll(rc)
		if err != nil {
			continue
		}
		chunk := string(buf)
		if strings.TrimSpace(chunk) != "" {
			b.WriteString(chunk)
			b.WriteByte('\n')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", fmt.Errorf("empty pdf text")
	}
	return out, nil
}

// ReadAll reads the full stream into memory (max enforced by caller).
func ReadAll(r io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(r, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, parser.ErrInvalidFile
	}
	return data, nil
}
