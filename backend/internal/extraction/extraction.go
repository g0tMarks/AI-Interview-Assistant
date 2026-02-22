package extraction

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/xavier268/mydocx"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrEmptyDocument     = errors.New("document contains no text")
)

// ExtractTextFromPDF extracts text from a PDF using ledongthuc/pdf (BSD-3-Clause, no license required).
// Uses a temp file because the library only supports opening by path.
func ExtractTextFromPDF(r io.ReadSeeker) (string, error) {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp("", "rubric-pdf-*.pdf")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.Write(fileBytes); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	f, pdfReader, err := pdf.Open(tmpPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	plainTextReader, err := pdfReader.GetPlainText()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(plainTextReader); err != nil {
		return "", err
	}

	text := strings.TrimSpace(buf.String())
	if text == "" {
		return "", ErrEmptyDocument
	}

	return text, nil
}

// ExtractTextFromDOCX extracts text from a DOCX file reader using mydocx (MIT license).
func ExtractTextFromDOCX(r io.Reader) (string, error) {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	content, err := mydocx.ExtractTextBytes(fileBytes)
	if err != nil {
		return "", err
	}

	if len(content) == 0 {
		return "", ErrEmptyDocument
	}

	// Flatten map[container][]paragraphs into a single string (document.xml first, then others)
	var parts []string
	for _, paragraphs := range content {
		for _, p := range paragraphs {
			if t := strings.TrimSpace(p); t != "" {
				parts = append(parts, t)
			}
		}
	}

	text := strings.TrimSpace(strings.Join(parts, "\n\n"))
	if text == "" {
		return "", ErrEmptyDocument
	}

	return text, nil
}

// ExtractText extracts text from a file based on its content type or filename extension.
// Supports PDF and DOCX formats.
func ExtractText(r io.ReadSeeker, contentType, filename string) (string, error) {
	// Determine format from content type or filename
	isPDF := strings.HasPrefix(contentType, "application/pdf") ||
		strings.HasSuffix(strings.ToLower(filename), ".pdf")
	isDOCX := strings.HasPrefix(contentType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document") ||
		strings.HasPrefix(contentType, "application/msword") ||
		strings.HasSuffix(strings.ToLower(filename), ".docx") ||
		strings.HasSuffix(strings.ToLower(filename), ".doc")

	switch {
	case isPDF:
		return ExtractTextFromPDF(r)
	case isDOCX:
		return ExtractTextFromDOCX(r)
	default:
		return "", ErrUnsupportedFormat
	}
}
