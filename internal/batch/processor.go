package batch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"extraction/internal/analysis"
	"extraction/internal/files"
	"extraction/internal/models"
	"extraction/internal/ocr"
	"extraction/internal/types"
	"extraction/internal/xfer"
)

// Processor handles batch processing of multiple files
type Processor struct {
	MaxConcurrency int
	SkipAnalysis   bool
	Lang           string
	DPI            int
	Source         analysis.DocumentSource
	ProgressChan   chan ProgressUpdate
}

// ProgressUpdate provides progress information during batch processing
type ProgressUpdate struct {
	CurrentFile    int
	TotalFiles     int
	CurrentFileURL string
	Status         string
	Error          error
}

// NewProcessor creates a new batch processor
func NewProcessor(maxConcurrency int, skipAnalysis bool, lang string, dpi int, source analysis.DocumentSource) *Processor {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // Default to 3 concurrent files
	}
	
	return &Processor{
		MaxConcurrency: maxConcurrency,
		SkipAnalysis:   skipAnalysis,
		Lang:           lang,
		DPI:            dpi,
		Source:         source,
		ProgressChan:   make(chan ProgressUpdate, 100),
	}
}

// ProcessFiles processes multiple files concurrently
func (p *Processor) ProcessFiles(ctx context.Context, inputs []string) (*types.BatchResult, error) {
	return p.ProcessFilesWithSources(ctx, inputs, nil)
}

// ProcessFilesWithSources processes multiple files with specific document sources
func (p *Processor) ProcessFilesWithSources(ctx context.Context, inputs []string, fileSources map[string]analysis.DocumentSource) (*types.BatchResult, error) {
	startTime := time.Now()
	
	// Create a semaphore to limit concurrent processing
	semaphore := make(chan struct{}, p.MaxConcurrency)
	
	// Create channels for results and errors
	resultsChan := make(chan types.FileResult, len(inputs))
	errorChan := make(chan error, len(inputs))
	
	// Initialize customer check - this will be shared across all processing
	check := &models.CustomerCheck{}
	now := time.Now()
	check.CheckCompletedAt = &now
	
	// Use a mutex to protect the shared customer check
	var checkMutex sync.Mutex
	
	var wg sync.WaitGroup
	
	// Process files concurrently
	for i, input := range inputs {
		wg.Add(1)
		go func(index int, inputURL string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Send progress update
			if p.ProgressChan != nil {
				p.ProgressChan <- ProgressUpdate{
					CurrentFile:    index + 1,
					TotalFiles:     len(inputs),
					CurrentFileURL: inputURL,
					Status:         "processing",
				}
			}
			
			// Determine document source for this file
			fileSource := p.Source
			if fileSources != nil {
				if specificSource, exists := fileSources[inputURL]; exists {
					fileSource = specificSource
				}
			}
			
			// Process the file
			result := p.processOneFileWithSource(ctx, inputURL, check, fileSource, &checkMutex)
			resultsChan <- result
			
			// Send completion update
			if p.ProgressChan != nil {
				status := "completed"
				if result.Error != "" {
					status = "failed"
				}
				p.ProgressChan <- ProgressUpdate{
					CurrentFile:    index + 1,
					TotalFiles:     len(inputs),
					CurrentFileURL: inputURL,
					Status:         status,
					Error:          func() error {
						if result.Error != "" {
							return fmt.Errorf(result.Error)
						}
						return nil
					}(),
				}
			}
		}(i, input)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorChan)
	}()
	
	// Collect results
	var results []types.FileResult
	var errors []error
	
	for result := range resultsChan {
		results = append(results, result)
	}
	
	for err := range errorChan {
		errors = append(errors, err)
	}
	
	endTime := time.Now()
	
	// Calculate statistics
	processedFiles := 0
	failedFiles := 0
	skippedFiles := 0
	
	for _, result := range results {
		if result.Error != "" {
			failedFiles++
		} else if result.ExtractedText == "" {
			skippedFiles++
		} else {
			processedFiles++
		}
	}
	
	batchResult := &types.BatchResult{
		TotalFiles:     len(inputs),
		ProcessedFiles: processedFiles,
		FailedFiles:    failedFiles,
		SkippedFiles:   skippedFiles,
		Results:        results,
		StartTime:      startTime,
		EndTime:        endTime,
		TotalDuration:  endTime.Sub(startTime),
		CustomerCheck:  check, // Include the aggregated customer check
	}
	
	// Post-process address comparison after all documents are processed
	analysis.CompareAddresses(check)
	
	return batchResult, nil
}

// processOneFile processes a single file
func (p *Processor) processOneFile(ctx context.Context, input string, check *models.CustomerCheck) types.FileResult {
	return p.processOneFileWithSource(ctx, input, check, p.Source, nil)
}

// processOneFileWithSource processes a single file with a specific document source
func (p *Processor) processOneFileWithSource(ctx context.Context, input string, check *models.CustomerCheck, source analysis.DocumentSource, checkMutex *sync.Mutex) types.FileResult {
	startTime := time.Now()
	
	localPath, sourceURL, filename, mediaType, err := xfer.DownloadToTemp(ctx, input)
	if err != nil {
		return types.FileResult{
			SourceURL:     sourceURL,
			FileName:      filename,
			FileType:      mediaType,
			Error:         err.Error(),
			ProcessedAt:   time.Now(),
			ProcessingTime: time.Since(startTime),
		}
	}
	
	// Get file size
	fileInfo, err := os.Stat(localPath)
	fileSize := int64(0)
	if err == nil {
		fileSize = fileInfo.Size()
	}
	
	var text string
	var extractErr error
	
	ft := files.DetectFileType(filename, mediaType)
	
	// Special case: if it's a site visit photo with unknown file type, treat it as an image
	if ft == files.FileTypeUnknown && source == analysis.SourceSiteVisitPhotos {
		ft = files.FileTypeImage
	}
	
	// Check if file type is processable
	if !files.IsProcessableFileType(ft) {
		extractErr = fmt.Errorf("unsupported file type: %s", ft.String())
	} else {
		switch ft {
		case files.FileTypeImage:
			text, extractErr = ocr.ExtractTextFromImageVision(ctx, localPath, p.Lang)
			// If vision processing fails due to bad image data, try alternative approaches
			if extractErr != nil && strings.Contains(extractErr.Error(), "Bad image data") {
				fmt.Printf("Image appears corrupted, trying alternative processing methods...\n")
				
				// Try to convert the image to a more standard format first
				convertedPath, convertErr := p.convertImageToStandardFormat(ctx, localPath)
				if convertErr == nil && convertedPath != "" {
					fmt.Printf("Successfully converted image, retrying OCR...\n")
					text, extractErr = ocr.ExtractTextFromImageVision(ctx, convertedPath, p.Lang)
					// Clean up converted file
					os.Remove(convertedPath)
				}
				
				// If still failing, try with Tesseract as fallback
				if extractErr != nil {
					fmt.Printf("Vision API still failing, trying Tesseract fallback...\n")
					tesseractText, tesseractErr := ocr.ExtractTextFromImageTesseract(ctx, localPath, p.Lang)
					if tesseractErr == nil && strings.TrimSpace(tesseractText) != "" {
						text = tesseractText
						extractErr = nil
						fmt.Printf("Successfully extracted text using Tesseract fallback\n")
					}
				}
				
				// If all methods fail, provide a helpful error message
				if extractErr != nil {
					// For site visit photos, provide a default value instead of failing completely
					if source == analysis.SourceSiteVisitPhotos {
						fmt.Printf("Site visit photo processing failed, using default value for company signboard\n")
						text = "No signboard visible or signboard unclear in site visit photos"
						extractErr = nil // Clear the error so processing can continue
					} else {
						extractErr = fmt.Errorf("image file appears to be corrupted or in an unsupported format, tried multiple processing methods: %w", extractErr)
					}
				}
			}
		case files.FileTypeText:
			b, err := os.ReadFile(localPath)
			if err != nil {
				extractErr = err
			} else {
				text = string(b)
			}
		case files.FileTypePDF:
			text, extractErr = ocr.ExtractTextFromPDF(ctx, localPath, p.Lang, p.DPI)
		case files.FileTypeWord, files.FileTypeExcel, files.FileTypePowerPoint:
			// For now, these are not supported but we can add support later
			extractErr = fmt.Errorf("office document processing not yet implemented for %s", ft.String())
		default:
			extractErr = fmt.Errorf("unsupported file type: %s", ft.String())
		}
	}
	
	res := types.FileResult{
		SourceURL:      sourceURL,
		LocalPath:      localPath,
		FileName:       filename,
		FileType:       ft.String(),
		ExtractedText:  text,
		ProcessedAt:    time.Now(),
		ProcessingTime: time.Since(startTime),
		FileSize:       fileSize,
		DocumentSource: string(source),
	}
	
	if extractErr != nil {
		res.Error = extractErr.Error()
	}
	
	// Analyze with AI if text was extracted successfully and analysis is not skipped
	if text != "" && !p.SkipAnalysis {
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
		
		// Update customer check with extracted data (thread-safe)
		if checkMutex != nil {
			checkMutex.Lock()
		}
		analysis.UpdateCustomerCheck(check, extractedData, source)
		if checkMutex != nil {
			checkMutex.Unlock()
		}
	}
	
	return res
}

// GetProcessingStats calculates processing statistics
func (p *Processor) GetProcessingStats(batchResult *types.BatchResult) types.ProcessingStats {
	totalSize := int64(0)
	for _, result := range batchResult.Results {
		totalSize += result.FileSize
	}
	
	var averageFileSize int64
	if len(batchResult.Results) > 0 {
		averageFileSize = totalSize / int64(len(batchResult.Results))
	}
	
	var processingRate float64
	if batchResult.TotalDuration.Seconds() > 0 {
		processingRate = float64(batchResult.ProcessedFiles) / batchResult.TotalDuration.Seconds()
	}
	
	var errorRate float64
	if batchResult.TotalFiles > 0 {
		errorRate = float64(batchResult.FailedFiles) / float64(batchResult.TotalFiles) * 100
	}
	
	return types.ProcessingStats{
		TotalFiles:      batchResult.TotalFiles,
		SuccessfulFiles: batchResult.ProcessedFiles,
		FailedFiles:     batchResult.FailedFiles,
		SkippedFiles:    batchResult.SkippedFiles,
		TotalSize:       totalSize,
		AverageFileSize: averageFileSize,
		ProcessingRate:  processingRate,
		ErrorRate:       errorRate,
	}
}

// Close closes the progress channel
func (p *Processor) Close() {
	if p.ProgressChan != nil {
		close(p.ProgressChan)
	}
}

// convertImageToStandardFormat attempts to convert a corrupted image to a standard PNG format
func (p *Processor) convertImageToStandardFormat(ctx context.Context, imagePath string) (string, error) {
	// Create a temporary file for the converted image
	tmpDir := filepath.Dir(imagePath)
	convertedPath := filepath.Join(tmpDir, "converted_"+filepath.Base(imagePath)+".png")
	
	// Try using ImageMagick's convert command if available
	cmd := exec.CommandContext(ctx, "convert", imagePath, "-quality", "95", convertedPath)
	if err := cmd.Run(); err != nil {
		// If ImageMagick is not available, try using ffmpeg
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", imagePath, "-y", "-frames:v", "1", convertedPath)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("image conversion failed: %v", err)
		}
	}
	
	// Check if the converted file exists and has content
	if info, err := os.Stat(convertedPath); err != nil || info.Size() == 0 {
		return "", fmt.Errorf("converted image file is empty or doesn't exist")
	}
	
	return convertedPath, nil
}
