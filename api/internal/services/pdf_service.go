package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PDFService handles PDF file operations
type PDFService struct {
	uploadDir string
}

// NewPDFService creates a new PDF service
func NewPDFService(uploadDir string) *PDFService {
	if uploadDir == "" {
		uploadDir = os.TempDir()
	}
	return &PDFService{
		uploadDir: uploadDir,
	}
}

// SaveUploadedFile saves an uploaded file and returns the path
func (s *PDFService) SaveUploadedFile(reader io.Reader, filename string) (string, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filepath.Base(filename), ext)
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, os.Getpid(), ext)
	filePath := filepath.Join(s.uploadDir, uniqueName)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

// ExtractText extracts text from a PDF file
func (s *PDFService) ExtractText(pdfPath string) (string, error) {
	// Try pdftotext first (from poppler-utils)
	text, err := s.extractWithPdftotext(pdfPath)
	if err == nil && text != "" {
		return text, nil
	}

	// Fall back to reading raw file and extracting text
	return s.extractTextManually(pdfPath)
}

// extractWithPdftotext uses pdftotext command if available
func (s *PDFService) extractWithPdftotext(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}
	return string(output), nil
}

// extractTextManually extracts text from PDF manually (basic implementation)
func (s *PDFService) extractTextManually(pdfPath string) (string, error) {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	var result strings.Builder

	// Simple PDF text extraction - look for text in parentheses
	// This is a very basic implementation and won't work for all PDFs
	for i := 0; i < len(content)-1; i++ {
		// Look for text in parentheses (common in PDF streams)
		if content[i] == '(' {
			j := i + 1
			for j < len(content) && content[j] != ')' {
				if content[j] == '\\' && j+1 < len(content) {
					j++
				}
				if content[j] >= 32 && content[j] < 127 {
					result.WriteByte(content[j])
				}
				j++
			}
			i = j
		}
	}

	text := result.String()

	// If we got nothing useful, just return readable characters from the file
	if len(strings.TrimSpace(text)) < 50 {
		text = extractReadableText(content)
	}

	return text, nil
}

// extractReadableText extracts readable ASCII text from content
func extractReadableText(content string) string {
	var result strings.Builder
	var word strings.Builder

	for _, c := range content {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '@' || c == '.' || c == '-' || c == '_' {
			word.WriteRune(c)
		} else {
			if word.Len() > 2 { // Only include words longer than 2 chars
				result.WriteString(word.String())
				result.WriteString(" ")
			}
			word.Reset()
		}
	}

	if word.Len() > 2 {
		result.WriteString(word.String())
	}

	return result.String()
}

// CleanupFile removes a temporary file
func (s *PDFService) CleanupFile(path string) error {
	return os.Remove(path)
}
