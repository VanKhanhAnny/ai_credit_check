package types

import (
	"time"
)

type FileResult struct {
	SourceURL     string
	LocalPath     string
	FileName      string
	FileType      string
	ExtractedText string
	Error         string
	ProcessedAt   time.Time
	ProcessingTime time.Duration
	FileSize      int64
	DocumentSource string // The type of document (business_license, evn_bill, etc.)
}

// BatchResult represents the result of processing multiple files
type BatchResult struct {
	TotalFiles     int
	ProcessedFiles int
	FailedFiles    int
	SkippedFiles   int
	Results        []FileResult
	StartTime      time.Time
	EndTime        time.Time
	TotalDuration  time.Duration
	CustomerCheck  interface{} // Will hold the aggregated customer check data
}

// FileGroup represents a group of related files
type FileGroup struct {
	ID          string
	Name        string
	Description string
	Files       []FileResult
	CreatedAt   time.Time
}

// ProcessingStats provides statistics about the processing operation
type ProcessingStats struct {
	TotalFiles       int
	SuccessfulFiles  int
	FailedFiles      int
	SkippedFiles     int
	TotalSize        int64
	AverageFileSize  int64
	ProcessingRate   float64 // files per second
	ErrorRate        float64 // percentage of failed files
}


