package pdf

import (
	"testing"

	"DebtEase/internal/parser"
)

func TestExtractText_UnencryptedEmptyPDF(t *testing.T) {
	data := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	_, err := ExtractText(data, "")
	if err == nil {
		t.Fatal("expected error for empty text pdf")
	}
}

func TestExtractText_PasswordProvidedAttemptsDecrypt(t *testing.T) {
	// Not a real encrypted PDF, but must not return password_required when password is set.
	data := []byte("%PDF-1.4\n/Encrypt\nfake\n%%EOF\n")
	_, err := ExtractText(data, "01011990")
	if err == nil {
		t.Fatal("expected error for fake encrypted pdf")
	}
	if parser.FailureCode(err) == "password_required" {
		t.Fatalf("should not ask for password when one was provided, got %v", err)
	}
}

func TestLooksEncryptedPDF(t *testing.T) {
	if !looksEncryptedPDF([]byte("%PDF-1.4\n/Encrypt\n")) {
		t.Fatal("expected encrypt marker detection")
	}
}
