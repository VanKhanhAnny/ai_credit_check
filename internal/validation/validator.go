package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"extraction/internal/types"
)

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid bool
	Errors  []string
	Warnings []string
	Score   float64 // Quality score from 0.0 to 1.0
}

// Validator validates file results and batch processing results
type Validator struct {
	MinTextLength    int
	MaxFileSize      int64
	AllowedFileTypes []string
	RequiredFields   []string
}

// NewValidator creates a new validator with default settings
func NewValidator() *Validator {
	return &Validator{
		MinTextLength: 10,
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		AllowedFileTypes: []string{"pdf", "image", "text", "word", "excel", "powerpoint"},
		RequiredFields: []string{"client_name", "tax_code_mst"},
	}
}

// ValidateFileResult validates a single file result
func (v *Validator) ValidateFileResult(result types.FileResult) ValidationResult {
	var errors []string
	var warnings []string
	score := 1.0

	// Check for processing errors
	if result.Error != "" {
		errors = append(errors, fmt.Sprintf("Processing error: %s", result.Error))
		score -= 0.5
	}

	// Check file size
	if result.FileSize > v.MaxFileSize {
		errors = append(errors, fmt.Sprintf("File too large: %d bytes (max: %d)", result.FileSize, v.MaxFileSize))
		score -= 0.2
	}

	// Check file type
	if !v.isAllowedFileType(result.FileType) {
		errors = append(errors, fmt.Sprintf("Unsupported file type: %s", result.FileType))
		score -= 0.3
	}

	// Check extracted text quality
	if result.ExtractedText == "" {
		errors = append(errors, "No text extracted from file")
		score -= 0.4
	} else {
		textQuality := v.validateTextQuality(result.ExtractedText)
		if textQuality.Score < 0.5 {
			warnings = append(warnings, "Low quality text extraction")
			score -= 0.2
		}
	}

	// Check processing time
	if result.ProcessingTime > 30*time.Second {
		warnings = append(warnings, fmt.Sprintf("Slow processing time: %v", result.ProcessingTime))
		score -= 0.1
	}

	// Check document source
	if result.DocumentSource == "" || result.DocumentSource == "unknown" {
		warnings = append(warnings, "Unknown document source type")
		score -= 0.1
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Score:    score,
	}
}

// ValidateBatchResult validates a batch processing result
func (v *Validator) ValidateBatchResult(batchResult *types.BatchResult) ValidationResult {
	var errors []string
	var warnings []string
	score := 1.0

	// Check overall success rate
	successRate := float64(batchResult.ProcessedFiles) / float64(batchResult.TotalFiles)
	if successRate < 0.8 {
		errors = append(errors, fmt.Sprintf("Low success rate: %.1f%%", successRate*100))
		score -= 0.3
	} else if successRate < 0.9 {
		warnings = append(warnings, fmt.Sprintf("Moderate success rate: %.1f%%", successRate*100))
		score -= 0.1
	}

	// Check processing time efficiency
	if batchResult.TotalDuration > 10*time.Minute {
		warnings = append(warnings, fmt.Sprintf("Long processing time: %v", batchResult.TotalDuration))
		score -= 0.1
	}

	// Validate individual file results
	validResults := 0
	totalScore := 0.0
	for _, result := range batchResult.Results {
		validation := v.ValidateFileResult(result)
		if validation.IsValid {
			validResults++
		}
		totalScore += validation.Score
	}

	// Calculate average quality score
	if len(batchResult.Results) > 0 {
		avgScore := totalScore / float64(len(batchResult.Results))
		if avgScore < 0.7 {
			warnings = append(warnings, fmt.Sprintf("Low average quality score: %.2f", avgScore))
			score -= 0.2
		}
	}

	// Check for duplicate files
	duplicates := v.findDuplicateFiles(batchResult.Results)
	if len(duplicates) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d potential duplicate files", len(duplicates)))
		score -= 0.1
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Score:    score,
	}
}

// validateTextQuality validates the quality of extracted text
func (v *Validator) validateTextQuality(text string) ValidationResult {
	var errors []string
	var warnings []string
	score := 1.0

	// Check minimum length
	if len(text) < v.MinTextLength {
		errors = append(errors, fmt.Sprintf("Text too short: %d characters (min: %d)", len(text), v.MinTextLength))
		score -= 0.4
	}

	// Check for excessive whitespace
	whitespaceRatio := float64(strings.Count(text, " ")+strings.Count(text, "\n")+strings.Count(text, "\t")) / float64(len(text))
	if whitespaceRatio > 0.5 {
		warnings = append(warnings, "High whitespace ratio in extracted text")
		score -= 0.2
	}

	// Check for common OCR errors
	ocrErrors := v.detectOCRErrors(text)
	if len(ocrErrors) > 0 {
		warnings = append(warnings, fmt.Sprintf("Potential OCR errors detected: %d", len(ocrErrors)))
		score -= 0.1
	}

	// Check for meaningful content
	if !v.hasMeaningfulContent(text) {
		warnings = append(warnings, "Text may not contain meaningful content")
		score -= 0.2
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Score:    score,
	}
}

// isAllowedFileType checks if the file type is allowed
func (v *Validator) isAllowedFileType(fileType string) bool {
	for _, allowed := range v.AllowedFileTypes {
		if fileType == allowed {
			return true
		}
	}
	return false
}

// detectOCRErrors detects common OCR errors in text
func (v *Validator) detectOCRErrors(text string) []string {
	var errors []string
	
	// Common OCR error patterns
	patterns := map[string]string{
		`[0-9]+[a-zA-Z]+[0-9]+`: "Mixed numbers and letters (possible OCR error)",
		`[a-zA-Z]{1,2}[0-9]{3,}`: "Short letters followed by numbers (possible OCR error)",
		`[0-9]{3,}[a-zA-Z]{1,2}`: "Numbers followed by short letters (possible OCR error)",
		`[^a-zA-Z0-9\s.,!?;:()\-]{3,}`: "Excessive special characters",
	}
	
	for pattern, description := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		if len(matches) > 0 {
			errors = append(errors, fmt.Sprintf("%s: %v", description, matches[:min(3, len(matches))]))
		}
	}
	
	return errors
}

// hasMeaningfulContent checks if text contains meaningful content
func (v *Validator) hasMeaningfulContent(text string) bool {
	// Check for common meaningful words
	meaningfulWords := []string{
		"company", "business", "license", "address", "name", "date", "number",
		"client", "customer", "invoice", "bill", "payment", "amount", "total",
		"document", "certificate", "agreement", "contract", "statement",
	}
	
	textLower := strings.ToLower(text)
	wordCount := 0
	
	for _, word := range meaningfulWords {
		if strings.Contains(textLower, word) {
			wordCount++
		}
	}
	
	// If we find at least 2 meaningful words, consider it meaningful
	return wordCount >= 2
}

// findDuplicateFiles finds potential duplicate files
func (v *Validator) findDuplicateFiles(results []types.FileResult) []string {
	var duplicates []string
	fileHashes := make(map[string][]string)
	
	for _, result := range results {
		if result.Error == "" && result.FileSize > 0 {
			// Simple hash based on filename and size
			hash := fmt.Sprintf("%s_%d", result.FileName, result.FileSize)
			fileHashes[hash] = append(fileHashes[hash], result.SourceURL)
		}
	}
	
	for hash, urls := range fileHashes {
		if len(urls) > 1 {
			duplicates = append(duplicates, fmt.Sprintf("Hash %s: %v", hash, urls))
		}
	}
	
	return duplicates
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetValidationSummary returns a summary of validation results
func (v *Validator) GetValidationSummary(results []types.FileResult) map[string]interface{} {
	totalFiles := len(results)
	validFiles := 0
	totalScore := 0.0
	var allErrors []string
	var allWarnings []string
	
	for _, result := range results {
		validation := v.ValidateFileResult(result)
		if validation.IsValid {
			validFiles++
		}
		totalScore += validation.Score
		allErrors = append(allErrors, validation.Errors...)
		allWarnings = append(allWarnings, validation.Warnings...)
	}
	
	avgScore := 0.0
	if totalFiles > 0 {
		avgScore = totalScore / float64(totalFiles)
	}
	
	return map[string]interface{}{
		"total_files":      totalFiles,
		"valid_files":      validFiles,
		"invalid_files":    totalFiles - validFiles,
		"success_rate":     float64(validFiles) / float64(totalFiles) * 100,
		"average_score":    avgScore,
		"total_errors":     len(allErrors),
		"total_warnings":   len(allWarnings),
		"common_errors":    v.getCommonErrors(allErrors),
		"common_warnings":  v.getCommonWarnings(allWarnings),
	}
}

// getCommonErrors returns the most common errors
func (v *Validator) getCommonErrors(errors []string) map[string]int {
	errorCounts := make(map[string]int)
	for _, err := range errors {
		errorCounts[err]++
	}
	return errorCounts
}

// getCommonWarnings returns the most common warnings
func (v *Validator) getCommonWarnings(warnings []string) map[string]int {
	warningCounts := make(map[string]int)
	for _, warning := range warnings {
		warningCounts[warning]++
	}
	return warningCounts
}
