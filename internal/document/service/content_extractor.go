package service

import (
	"cdk-office/internal/document/domain"
	"cdk-office/pkg/logger"
	"bytes"
	"fmt"
	"github.com/EndFirstCorp/doc2txt"
	"github.com/dslipak/pdf"
	"github.com/nguyenthenguyen/docx"
	"golang.org/x/net/html"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ContentExtractorInterface defines the interface for content extraction service
type ContentExtractorInterface interface {
	ExtractContent(document *domain.Document) (string, error)
}

// ContentExtractor implements the ContentExtractorInterface
type ContentExtractor struct {
	storagePath string
}

// NewContentExtractor creates a new instance of ContentExtractor
func NewContentExtractor(storagePath string) *ContentExtractor {
	return &ContentExtractor{
		storagePath: storagePath,
	}
}

// ExtractContent extracts content from a document
func (ce *ContentExtractor) ExtractContent(document *domain.Document) (string, error) {
	// Determine file path
	filePath := document.FilePath
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(ce.storagePath, filePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("file not found", "file_path", filePath)
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	// Extract content based on file type
	switch strings.ToLower(document.MimeType) {
	case "text/plain":
		return ce.extractTextContent(filePath)
	case "text/html":
		return ce.extractHTMLContent(filePath)
	case "application/pdf":
		return ce.extractPDFContent(filePath)
	case "application/msword":
		return ce.extractDOCContent(filePath)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ce.extractDOCXContent(filePath)
	default:
		// For unsupported file types, return basic file information
		return ce.extractGenericContent(document)
	}
}

// extractTextContent extracts content from a text file
func (ce *ContentExtractor) extractTextContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("failed to open text file", "error", err)
		return "", fmt.Errorf("failed to open text file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		logger.Error("failed to read text file", "error", err)
		return "", fmt.Errorf("failed to read text file: %v", err)
	}

	return string(content), nil
}

// extractHTMLContent extracts content from an HTML file
func (ce *ContentExtractor) extractHTMLContent(filePath string) (string, error) {
	// Open the HTML file
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("failed to open HTML file", "error", err)
		return "", fmt.Errorf("failed to open HTML file: %v", err)
	}
	defer file.Close()

	// Parse the HTML document
	doc, err := html.Parse(file)
	if err != nil {
		logger.Error("failed to parse HTML file", "error", err)
		return "", fmt.Errorf("failed to parse HTML file: %v", err)
	}

	// Extract text content from the HTML document
	var textContent string
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			textContent += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(doc)

	return textContent, nil
}

// extractPDFContent extracts content from a PDF file
func (ce *ContentExtractor) extractPDFContent(filePath string) (string, error) {
	// Open the PDF file
	file, r, err := pdf.Open(filePath)
	if err != nil {
		logger.Error("failed to open PDF file", "error", err)
		return "", fmt.Errorf("failed to open PDF file: %v", err)
	}
	defer file.Close()

	// Extract plain text content from the PDF
	b, err := r.GetPlainText()
	if err != nil {
		logger.Error("failed to extract text from PDF file", "error", err)
		return "", fmt.Errorf("failed to extract text from PDF file: %v", err)
	}

	// Read the extracted text into a string
	var buf bytes.Buffer
	buf.ReadFrom(b)
	
	return buf.String(), nil
}

// extractDOCContent extracts content from a DOC file
func (ce *ContentExtractor) extractDOCContent(filePath string) (string, error) {
	// Open the DOC file
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("failed to open DOC file", "error", err)
		return "", fmt.Errorf("failed to open DOC file: %v", err)
	}
	defer file.Close()

	// Extract text content from the DOC file
	text, err := doc2txt.DocToText(file)
	if err != nil {
		logger.Error("failed to extract text from DOC file", "error", err)
		return "", fmt.Errorf("failed to extract text from DOC file: %v", err)
	}

	return text, nil
}

// extractDOCXContent extracts content from a DOCX file
func (ce *ContentExtractor) extractDOCXContent(filePath string) (string, error) {
	// Open the DOCX file
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		logger.Error("failed to open DOCX file", "error", err)
		return "", fmt.Errorf("failed to open DOCX file: %v", err)
	}
	defer r.Close()

	// Get the content from the DOCX file
	docxFile := r.Editable()
	content := docxFile.GetContent()

	return content, nil
}

// extractGenericContent extracts generic content from a document
func (ce *ContentExtractor) extractGenericContent(document *domain.Document) (string, error) {
	// For unsupported file types, return basic file information
	content := fmt.Sprintf("File: %s\nSize: %d bytes\nMIME Type: %s\n", 
		document.Title, document.FileSize, document.MimeType)
	
	return content, nil
}