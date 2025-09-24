package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"extraction/internal/analysis"
	"extraction/internal/batch"
	"extraction/internal/export"
	"extraction/internal/files"
	"extraction/internal/grouping"
	"extraction/internal/models"
	"extraction/internal/ocr"
	"extraction/internal/types"
	"extraction/internal/validation"
	"extraction/internal/xfer"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, ",") }
func (s *stringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// FileSourcePair represents a file with its specific document source
type FileSourcePair struct {
	FilePath string
	Source   analysis.DocumentSource
}

type fileSourcePairFlag []FileSourcePair

func (f *fileSourcePairFlag) String() string {
	var parts []string
	for _, pair := range *f {
		parts = append(parts, fmt.Sprintf("%s:%s", pair.FilePath, pair.Source))
	}
	return strings.Join(parts, ",")
}

func (f *fileSourcePairFlag) Set(v string) error {
	// Parse format: "file_path:source_type"
	// Need to handle URLs which contain colons (like https://)
	// Find the last colon that's not part of a URL scheme
	var lastColonIndex int = -1
	
	// If it starts with a protocol (http:// or https://), skip the first colon
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		// Find the last colon after the protocol
		for i := len(v) - 1; i >= 0; i-- {
			if v[i] == ':' && i > 7 { // Skip the : in http:// or https://
				lastColonIndex = i
				break
			}
		}
	} else {
		// For non-URLs, just find the last colon
		lastColonIndex = strings.LastIndex(v, ":")
	}
	
	if lastColonIndex == -1 {
		return fmt.Errorf("invalid format, expected 'file_path:source_type', got: %s", v)
	}
	
	filePath := strings.TrimSpace(v[:lastColonIndex])
	sourceType := strings.TrimSpace(v[lastColonIndex+1:])
	
	*f = append(*f, FileSourcePair{
		FilePath: filePath,
		Source:   analysis.DocumentSource(sourceType),
	})
	return nil
}

func readLinesFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func main() {
	var inputs stringSliceFlag
	var fileSources fileSourcePairFlag
	var linksFile string
	var outputPath string
	var jsonOutputPath string
	var lang string
	var timeoutSec int
	var dpi int
	var docSource string
	var skipAnalysis bool
	var maxConcurrency int
	var showProgress bool
	var enableGrouping bool
	var enableValidation bool
	var groupByDocumentType bool
	var groupByClient bool

	flag.Var(&inputs, "input", "Input URL or local path (repeatable)")
	flag.Var(&fileSources, "file-source", "File with specific document source: 'file_path:source_type' (repeatable)")
	flag.StringVar(&linksFile, "links-file", "", "Path to a text file containing URLs/paths (one per line)")
	flag.StringVar(&outputPath, "out", "output.xlsx", "Path to the output file (.xlsx)")
	flag.StringVar(&jsonOutputPath, "json", "", "Path to save extracted JSON data (optional)")
	flag.StringVar(&lang, "lang", "eng", "Language(s), e.g. 'eng' or 'eng+vie'")
	flag.StringVar(&docSource, "source", "unknown", "Document source type (business_license, evn_bill, rental_agreement, etc.)")
	flag.IntVar(&timeoutSec, "timeout", 1200, "Overall timeout in seconds")
	flag.IntVar(&dpi, "dpi", 300, "PDF rasterization DPI for OCR")
	flag.BoolVar(&skipAnalysis, "skip-analysis", false, "Skip AI analysis (extract text only)")
	flag.IntVar(&maxConcurrency, "concurrency", 3, "Maximum number of files to process concurrently")
	flag.BoolVar(&showProgress, "progress", false, "Show progress updates during processing")
	flag.BoolVar(&enableGrouping, "group", false, "Enable file grouping analysis")
	flag.BoolVar(&enableValidation, "validate", false, "Enable validation and quality checks")
	flag.BoolVar(&groupByDocumentType, "group-by-type", false, "Group files by document type")
	flag.BoolVar(&groupByClient, "group-by-client", false, "Group files by client name")
	flag.Parse()

	if linksFile != "" {
		lines, err := readLinesFile(linksFile)
		if err != nil {
			log.Fatalf("failed to read links file: %v", err)
		}
		inputs = append(inputs, lines...)
	}
	inputs = append(inputs, flag.Args()...)

	// Combine inputs and file-source pairs
	var allInputs []string
	fileSourceMap := make(map[string]analysis.DocumentSource)
	
	// Add regular inputs
	for _, input := range inputs {
		allInputs = append(allInputs, input)
	}
	
	// Add file-source pairs
	for _, pair := range fileSources {
		allInputs = append(allInputs, pair.FilePath)
		fileSourceMap[pair.FilePath] = pair.Source
	}

	if len(allInputs) == 0 {
		fmt.Println("Usage: extract --input <url|path> [--input <url|path> ...] [--file-source 'file_path:source_type'] [--links-file file] --out output.xlsx [--json data.json] [--lang eng] [--source document_type] [--dpi 300] [--skip-analysis] [--concurrency 3] [--progress] [--group] [--validate] [--group-by-type] [--group-by-client]")
		fmt.Println("\nDocument source types: business_license, evn_bill, rental_agreement, land_certificate, id_check, financial_statement, site_visit_photos, cic_report")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Load .env if present
	_ = loadDotEnvIfPresent()

	// Parse document source
	source := analysis.DocumentSource(docSource)

	// Create batch processor
	processor := batch.NewProcessor(maxConcurrency, skipAnalysis, lang, dpi, source)
	defer processor.Close()

	// Start progress monitoring if requested
	if showProgress {
		go monitorProgress(processor.ProgressChan)
	}

	// Process inputs using batch processor
	var batchResult *types.BatchResult
	var err error
	if len(fileSourceMap) > 0 {
		// Use specific document sources for files
		batchResult, err = processor.ProcessFilesWithSources(ctx, allInputs, fileSourceMap)
	} else {
		// Use default document source for all files
		batchResult, err = processor.ProcessFiles(ctx, allInputs)
	}
	if err != nil {
		log.Fatalf("failed to process files: %v", err)
	}
	
	// Get the aggregated customer check from the processor
	// Note: The customer check is now properly aggregated in the batch processor

	// Get processing statistics
	stats := processor.GetProcessingStats(batchResult)
	
	// Print processing summary
	fmt.Printf("\n=== Processing Summary ===\n")
	fmt.Printf("Total files: %d\n", stats.TotalFiles)
	fmt.Printf("Successfully processed: %d\n", stats.SuccessfulFiles)
	fmt.Printf("Failed: %d\n", stats.FailedFiles)
	fmt.Printf("Skipped: %d\n", stats.SkippedFiles)
	fmt.Printf("Total processing time: %v\n", batchResult.TotalDuration)
	fmt.Printf("Processing rate: %.2f files/second\n", stats.ProcessingRate)
	fmt.Printf("Error rate: %.1f%%\n", stats.ErrorRate)
	fmt.Printf("Total data processed: %.2f MB\n", float64(stats.TotalSize)/(1024*1024))
	fmt.Printf("========================\n\n")

	results := batchResult.Results

	// Perform file grouping if enabled
	if enableGrouping {
		groupingAnalyzer := grouping.NewGroupingAnalyzer(groupByDocumentType, true, groupByClient, true)
		groups := groupingAnalyzer.AnalyzeAndGroup(results)
		
		fmt.Printf("\n=== File Grouping Analysis ===\n")
		fmt.Printf("Created %d file groups\n", len(groups))
		
		stats := groupingAnalyzer.GetGroupStatistics(groups)
		fmt.Printf("Total files: %v\n", stats["total_files"])
		fmt.Printf("Total size: %.2f MB\n", float64(stats["total_size"].(int64))/(1024*1024))
		fmt.Printf("Average files per group: %.1f\n", stats["average_files_per_group"])
		fmt.Printf("=============================\n\n")
		
		// Save grouping results to a separate file
		groupingOutputPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_groups.json"
		if err := saveGroupingResults(groups, groupingOutputPath); err != nil {
			log.Printf("Warning: failed to save grouping results: %v", err)
		} else {
			fmt.Printf("Saved grouping results to %s\n", groupingOutputPath)
		}
	}

	// Perform validation if enabled
	if enableValidation {
		validator := validation.NewValidator()
		validationResult := validator.ValidateBatchResult(batchResult)
		
		fmt.Printf("\n=== Validation Results ===\n")
		fmt.Printf("Overall valid: %t\n", validationResult.IsValid)
		fmt.Printf("Quality score: %.2f/1.0\n", validationResult.Score)
		fmt.Printf("Errors: %d\n", len(validationResult.Errors))
		fmt.Printf("Warnings: %d\n", len(validationResult.Warnings))
		
		if len(validationResult.Errors) > 0 {
			fmt.Printf("\nErrors:\n")
			for _, err := range validationResult.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}
		
		if len(validationResult.Warnings) > 0 {
			fmt.Printf("\nWarnings:\n")
			for _, warning := range validationResult.Warnings {
				fmt.Printf("  - %s\n", warning)
			}
		}
		
		// Get detailed validation summary
		summary := validator.GetValidationSummary(results)
		fmt.Printf("\nDetailed Summary:\n")
		fmt.Printf("  Success rate: %.1f%%\n", summary["success_rate"])
		fmt.Printf("  Average score: %.2f\n", summary["average_score"])
		fmt.Printf("  Common errors: %v\n", summary["common_errors"])
		fmt.Printf("=======================\n\n")
		
		// Save validation results
		validationOutputPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_validation.json"
		if err := saveValidationResults(validationResult, summary, validationOutputPath); err != nil {
			log.Printf("Warning: failed to save validation results: %v", err)
		} else {
			fmt.Printf("Saved validation results to %s\n", validationOutputPath)
		}
	}

	// Export raw extraction results and structured customer check
	rawOutputPath := strings.TrimSuffix(outputPath, ".xlsx") + "_raw.xlsx"
	if err := export.WriteResults(results, rawOutputPath); err != nil {
		log.Fatalf("failed to write raw results: %v", err)
	}
	fmt.Printf("Wrote %d extraction results to %s\n", len(results), rawOutputPath)
	
	// Write structured customer check data
	if customerCheck, ok := batchResult.CustomerCheck.(*models.CustomerCheck); ok {
		if err := export.WriteCustomerCheck(customerCheck, outputPath); err != nil {
			log.Fatalf("failed to write customer check: %v", err)
		}
		fmt.Printf("Wrote structured customer check data to %s\n", outputPath)
	}

	// Export JSON data if requested
	if jsonOutputPath != "" {
		// Use the aggregated customer check from batch processing
		if customerCheck, ok := batchResult.CustomerCheck.(*models.CustomerCheck); ok {
			jsonData, err := json.MarshalIndent(customerCheck, "", "  ")
			if err != nil {
				log.Fatalf("failed to marshal JSON: %v", err)
			}
			if err := os.WriteFile(jsonOutputPath, jsonData, 0644); err != nil {
				log.Fatalf("failed to write JSON file: %v", err)
			}
			fmt.Printf("Wrote analyzed data to %s\n", jsonOutputPath)
		} else {
			log.Printf("Warning: No customer check data available for JSON export")
		}
	}
}

// monitorProgress monitors and displays progress updates
func monitorProgress(progressChan <-chan batch.ProgressUpdate) {
	for update := range progressChan {
		if update.Error != nil {
			fmt.Printf("[%d/%d] ❌ %s - %s\n", update.CurrentFile, update.TotalFiles, update.CurrentFileURL, update.Error.Error())
		} else {
			fmt.Printf("[%d/%d] ✅ %s - %s\n", update.CurrentFile, update.TotalFiles, update.CurrentFileURL, update.Status)
		}
	}
}

// saveGroupingResults saves grouping results to a JSON file
func saveGroupingResults(groups []types.FileGroup, outputPath string) error {
	jsonData, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, jsonData, 0644)
}

// saveValidationResults saves validation results to a JSON file
func saveValidationResults(validationResult validation.ValidationResult, summary map[string]interface{}, outputPath string) error {
	result := map[string]interface{}{
		"validation_result": validationResult,
		"summary":          summary,
		"timestamp":        time.Now().Format(time.RFC3339),
	}
	
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, jsonData, 0644)
}

func processOne(ctx context.Context, input string, lang string, dpi int, source analysis.DocumentSource, skipAnalysis bool, useGemini bool, check *models.CustomerCheck) types.FileResult {
	localPath, sourceURL, filename, mediaType, err := xfer.DownloadToTemp(ctx, input)
	if err != nil {
		return types.FileResult{SourceURL: sourceURL, FileName: filename, FileType: mediaType, Error: err.Error()}
	}

	var text string
	var extractErr error

	ft := files.DetectFileType(filename, mediaType)
	switch ft {
	case files.FileTypeImage:
		text, extractErr = ocr.ExtractTextFromImageVision(ctx, localPath, lang)
	case files.FileTypeText:
		b, err := os.ReadFile(localPath)
		if err != nil {
			extractErr = err
		} else {
			text = string(b)
		}
	case files.FileTypePDF:
		text, extractErr = ocr.ExtractTextFromPDF(ctx, localPath, lang, dpi)
	default:
		extractErr = errors.New("unsupported file type for now; supported: image, text, pdf")
	}

	res := types.FileResult{
		SourceURL:     sourceURL,
		LocalPath:     localPath,
		FileName:      filename,
		FileType:      ft.String(),
		ExtractedText: text,
	}
	if extractErr != nil {
		res.Error = extractErr.Error()
	}

	// Analyze with AI if text was extracted successfully and analysis is not skipped
	if text != "" && !skipAnalysis {
		var extractedData map[string]interface{}
		var err error
		
		// Use Gemini API
		client, clientErr := analysis.NewGeminiClient()
		if clientErr != nil {
			res.Error = fmt.Sprintf("Gemini client initialization error: %v", clientErr)
			return res
		}
		
		extractedData, err = client.AnalyzeDocument(ctx, text, source)
		if err != nil {
			res.Error = fmt.Sprintf("Gemini analysis error: %v", err)
			return res
		}

		// Update customer check with extracted data
		analysis.UpdateCustomerCheck(check, extractedData, source)
	}

	return res
}

// Minimal .env loader: reads key=value per line and sets env if not already set
func loadDotEnvIfPresent() error {
	file := ".env"
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return nil
}

