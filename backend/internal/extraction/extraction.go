package extraction

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrEmptyDocument     = errors.New("document contains no text")
)

// ExtractTextFromPDF extracts text from a PDF file reader.
func ExtractTextFromPDF(r io.ReadSeeker) (string, error) {
	pdfReader, err := model.NewPdfReader(r)
	if err != nil {
		return "", err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	if numPages == 0 {
		return "", ErrEmptyDocument
	}

	var textBuilder strings.Builder

	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			continue // Skip pages that can't be read
		}

		ex, err := extractor.New(page)
		if err != nil {
			continue // Skip pages that can't be extracted
		}

		pageText, err := ex.ExtractText()
		if err != nil {
			continue // Skip pages with extraction errors
		}

		if pageText != "" {
			textBuilder.WriteString(pageText)
			if i < numPages {
				textBuilder.WriteString("\n\n")
			}
		}
	}

	text := strings.TrimSpace(textBuilder.String())
	if text == "" {
		return "", ErrEmptyDocument
	}

	return text, nil
}

// ExtractTextFromDOCX extracts text from a DOCX file reader.
func ExtractTextFromDOCX(r io.Reader) (string, error) {
	// Read all bytes first since document.Read needs io.ReaderAt
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	// Create a ReaderAt from bytes
	readerAt := bytes.NewReader(fileBytes)
	doc, err := document.Read(readerAt, int64(len(fileBytes)))
	if err != nil {
		return "", err
	}

	// Extract text using the ExtractText method
	docText := doc.ExtractText()
	if docText == nil {
		return "", ErrEmptyDocument
	}

	// Get plain text from the DocText object
	text := docText.Text()
	text = strings.TrimSpace(text)
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
