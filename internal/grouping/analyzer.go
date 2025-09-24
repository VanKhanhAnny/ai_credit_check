package grouping

import (
	"regexp"
	"strings"
	"time"

	"extraction/internal/types"
)

// GroupingAnalyzer analyzes files and groups them based on various criteria
type GroupingAnalyzer struct {
	GroupByDocumentType bool
	GroupByDate         bool
	GroupByClient       bool
	GroupBySource       bool
}

// NewGroupingAnalyzer creates a new grouping analyzer
func NewGroupingAnalyzer(groupByDocumentType, groupByDate, groupByClient, groupBySource bool) *GroupingAnalyzer {
	return &GroupingAnalyzer{
		GroupByDocumentType: groupByDocumentType,
		GroupByDate:         groupByDate,
		GroupByClient:       groupByClient,
		GroupBySource:       groupBySource,
	}
}

// AnalyzeAndGroup analyzes file results and groups them
func (ga *GroupingAnalyzer) AnalyzeAndGroup(results []types.FileResult) []types.FileGroup {
	groups := make(map[string]*types.FileGroup)
	
	for _, result := range results {
		groupKey := ga.generateGroupKey(result)
		
		if group, exists := groups[groupKey]; exists {
			group.Files = append(group.Files, result)
		} else {
			groups[groupKey] = &types.FileGroup{
				ID:          groupKey,
				Name:        ga.generateGroupName(result),
				Description: ga.generateGroupDescription(result),
				Files:       []types.FileResult{result},
				CreatedAt:   time.Now(),
			}
		}
	}
	
	// Convert map to slice
	var groupSlice []types.FileGroup
	for _, group := range groups {
		groupSlice = append(groupSlice, *group)
	}
	
	return groupSlice
}

// generateGroupKey creates a unique key for grouping files
func (ga *GroupingAnalyzer) generateGroupKey(result types.FileResult) string {
	var keyParts []string
	
	if ga.GroupByDocumentType {
		keyParts = append(keyParts, result.DocumentSource)
	}
	
	if ga.GroupBySource {
		// Group by source domain or local path
		if result.SourceURL != "" {
			keyParts = append(keyParts, extractDomain(result.SourceURL))
		} else {
			keyParts = append(keyParts, "local")
		}
	}
	
	if ga.GroupByDate {
		// Group by processing date (day level)
		keyParts = append(keyParts, result.ProcessedAt.Format("2006-01-02"))
	}
	
	if ga.GroupByClient {
		// Try to extract client name from filename or content
		clientName := extractClientName(result)
		keyParts = append(keyParts, clientName)
	}
	
	// If no grouping criteria specified, group by file type
	if len(keyParts) == 0 {
		keyParts = append(keyParts, result.FileType)
	}
	
	return strings.Join(keyParts, "_")
}

// generateGroupName creates a human-readable name for the group
func (ga *GroupingAnalyzer) generateGroupName(result types.FileResult) string {
	var nameParts []string
	
	if ga.GroupByDocumentType {
		nameParts = append(nameParts, formatDocumentType(result.DocumentSource))
	}
	
	if ga.GroupBySource {
		if result.SourceURL != "" {
			nameParts = append(nameParts, extractDomain(result.SourceURL))
		} else {
			nameParts = append(nameParts, "Local Files")
		}
	}
	
	if ga.GroupByDate {
		nameParts = append(nameParts, result.ProcessedAt.Format("Jan 2, 2006"))
	}
	
	if ga.GroupByClient {
		clientName := extractClientName(result)
		if clientName != "unknown" {
			nameParts = append(nameParts, clientName)
		}
	}
	
	if len(nameParts) == 0 {
		nameParts = append(nameParts, formatFileType(result.FileType))
	}
	
	return strings.Join(nameParts, " - ")
}

// generateGroupDescription creates a description for the group
func (ga *GroupingAnalyzer) generateGroupDescription(result types.FileResult) string {
	var descParts []string
	
	if ga.GroupByDocumentType {
		descParts = append(descParts, "Document Type: "+formatDocumentType(result.DocumentSource))
	}
	
	if ga.GroupBySource {
		if result.SourceURL != "" {
			descParts = append(descParts, "Source: "+extractDomain(result.SourceURL))
		} else {
			descParts = append(descParts, "Source: Local Files")
		}
	}
	
	if ga.GroupByDate {
		descParts = append(descParts, "Date: "+result.ProcessedAt.Format("January 2, 2006"))
	}
	
	if ga.GroupByClient {
		clientName := extractClientName(result)
		if clientName != "unknown" {
			descParts = append(descParts, "Client: "+clientName)
		}
	}
	
	if len(descParts) == 0 {
		descParts = append(descParts, "File Type: "+formatFileType(result.FileType))
	}
	
	return strings.Join(descParts, " | ")
}

// extractDomain extracts the domain from a URL
func extractDomain(url string) string {
	// Simple domain extraction - in production you might want to use url.Parse
	re := regexp.MustCompile(`https?://([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

// extractClientName tries to extract client name from filename or content
func extractClientName(result types.FileResult) string {
	// Try to extract from filename first
	filename := strings.ToLower(result.FileName)
	
	// Common patterns for client names in filenames
	patterns := []string{
		`client[_-]?(\w+)`,
		`(\w+)[_-]?client`,
		`company[_-]?(\w+)`,
		`(\w+)[_-]?company`,
		`business[_-]?(\w+)`,
		`(\w+)[_-]?business`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(filename)
		if len(matches) > 1 {
			return strings.Title(matches[1])
		}
	}
	
	// Try to extract from extracted text (first few words)
	if result.ExtractedText != "" {
		words := strings.Fields(result.ExtractedText)
		if len(words) > 0 {
			// Look for company-like words in the first 50 words
			for i, word := range words {
				if i > 50 {
					break
				}
				word = strings.Trim(word, ".,!?;:")
				if len(word) > 3 && isLikelyCompanyName(word) {
					return word
				}
			}
		}
	}
	
	return "unknown"
}

// isLikelyCompanyName checks if a word looks like a company name
func isLikelyCompanyName(word string) bool {
	// Simple heuristics for company names
	companySuffixes := []string{"ltd", "inc", "corp", "llc", "co", "group", "company", "enterprise"}
	wordLower := strings.ToLower(word)
	
	for _, suffix := range companySuffixes {
		if strings.HasSuffix(wordLower, suffix) {
			return true
		}
	}
	
	// Check if it's capitalized (likely a proper noun)
	if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
		return true
	}
	
	return false
}

// formatDocumentType formats document type for display
func formatDocumentType(docType string) string {
	switch docType {
	case "business_license":
		return "Business License"
	case "evn_bill":
		return "EVN Bill"
	case "rental_agreement":
		return "Rental Agreement"
	case "land_certificate":
		return "Land Certificate"
	case "id_check":
		return "ID Check"
	case "bank_statement":
		return "Bank Statement"
	case "site_visit_photos":
		return "Site Visit Photos"
	case "cic_report":
		return "CIC Report"
	default:
		return strings.Title(strings.ReplaceAll(docType, "_", " "))
	}
}

// formatFileType formats file type for display
func formatFileType(fileType string) string {
	switch fileType {
	case "pdf":
		return "PDF Documents"
	case "image":
		return "Images"
	case "text":
		return "Text Files"
	case "word":
		return "Word Documents"
	case "excel":
		return "Excel Spreadsheets"
	case "powerpoint":
		return "PowerPoint Presentations"
	default:
		return strings.Title(fileType) + " Files"
	}
}

// GetGroupStatistics returns statistics about the groups
func (ga *GroupingAnalyzer) GetGroupStatistics(groups []types.FileGroup) map[string]interface{} {
	totalFiles := 0
	totalSize := int64(0)
	successfulFiles := 0
	failedFiles := 0
	
	for _, group := range groups {
		totalFiles += len(group.Files)
		for _, file := range group.Files {
			totalSize += file.FileSize
			if file.Error == "" {
				successfulFiles++
			} else {
				failedFiles++
			}
		}
	}
	
	return map[string]interface{}{
		"total_groups":     len(groups),
		"total_files":      totalFiles,
		"total_size":       totalSize,
		"successful_files": successfulFiles,
		"failed_files":     failedFiles,
		"average_files_per_group": float64(totalFiles) / float64(len(groups)),
	}
}
