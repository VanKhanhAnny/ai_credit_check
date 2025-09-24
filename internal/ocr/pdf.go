package ocr

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ExtractTextFromPDF first tries `pdftotext` (embedded text). If empty, it falls back
// to rendering pages with `pdftoppm` and OCRing them with Google Cloud Vision.
// Requires Poppler tools (pdftotext, pdftoppm) on PATH.
func ExtractTextFromPDF(ctx context.Context, pdfPath string, lang string, dpi int) (string, error) {
	// Try to extract embedded text
	txt, err := runPdfToText(ctx, pdfPath)
	if err == nil && len(strings.TrimSpace(txt)) > 10 {
		// Only use embedded text if it has substantial content (more than 10 non-whitespace chars)
		return txt, nil
	}
	// If pdftotext fails, returns empty text, or returns only control characters, fall back to OCR
	if dpi <= 0 {
		dpi = 300
	}
	return ExtractTextFromPDFVision(ctx, pdfPath, lang, dpi)
}

func runPdfToText(ctx context.Context, pdfPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", pdfPath, "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext error: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}