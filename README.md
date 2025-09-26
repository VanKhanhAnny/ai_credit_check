# Document Extraction and Analysis System

## System Overview

This is a comprehensive document processing system that extracts text from various document types using OCR technology and analyzes the content using Gemini API to extract structured information for customer verification checks. The system is designed as a modular Go application with two main entry points for different use cases.

## Architecture

The system follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Command Layer                            │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   extractor     │    │     vision      │                │
│  │   (main.go)     │    │   (main.go)     │                │
│  └─────────────────┘    └─────────────────┘                │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  Business Logic Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Batch     │  │  Analysis   │  │      Export         │  │
│  │ Processing  │  │    (AI)     │  │    (XLSX/JSON)      │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │     OCR     │  │    Files    │  │      Transfer       │  │
│  │ (Vision/Tess)│  │ Detection  │  │   (Download)        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Data Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Models    │  │    Types    │  │    Validation       │  │
│  │ (Customer)  │  │ (File/Result)│  │    & Grouping      │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

### Root Directory

- **`go.mod`**
- **`go.sum`**
- **`README.md`**
- **`sample_links.txt`**

### Executables

- **`extract.exe`** - Main application with full AI analysis pipeline (can run in OCR-only mode with `--skip-analysis`)
- **`bin/tesseract-extract.exe`** - Tesseract OCR executable (optional fallback)

### Command Applications (`cmd/`)

#### `cmd/extractor/main.go`

**Purpose**: Main application entry point for full document processing pipeline
**Functionality**:

- Processes multiple document types with AI analysis
- Supports batch processing with concurrency control
- Implements file grouping and validation features
- Aggregates extracted data into structured customer check models
- Exports results in multiple formats (XLSX, JSON, CSV, RTF)

### Core Business Logic (`internal/`)

#### `internal/analysis/`

**Purpose**: AI-powered document analysis and data extraction

- **`gemini_client.go`** - Google Gemini AI client for document analysis
- **`prompts.go`** - AI prompts and templates for different document types
- **`customer_check_updater.go`** - Updates customer check models with extracted data
- **`types.go`** - Document source type definitions and constants

#### `internal/batch/`

**Purpose**: Concurrent batch processing management

- **`processor.go`** - Handles concurrent file processing with semaphores and progress tracking

#### `internal/export/`

**Purpose**: Data export functionality

- **`xlsx.go`** - Excel file generation and formatting
- **`customer_check.go`** - Structured customer check data export

#### `internal/files/`

**Purpose**: File type detection and processing utilities

- **`types.go`** - File type enumeration and detection logic

#### `internal/grouping/`

**Purpose**: File grouping and organization analysis

- **`analyzer.go`** - Groups related files by document type, client, or other criteria

#### `internal/models/`

**Purpose**: Data models and business entities

- **`customer_check.go`** - Comprehensive customer check data model with corporate, land, financial, and additional information structures

#### `internal/ocr/`

**Purpose**: Optical Character Recognition services

- **`pdf.go`** - PDF text extraction using Tesseract OCR
- **`vision.go`** - Google Cloud Vision API integration with Tesseract fallback

#### `internal/types/`

**Purpose**: Common data structures and types

- **`types.go`** - File results, batch results, and processing statistics structures

#### `internal/validation/`

**Purpose**: Data validation and quality assurance

- **`validator.go`** - Validates extracted data quality and completeness

#### `internal/xfer/`

**Purpose**: File transfer and download utilities

- **`download.go`** - Downloads files from URLs to temporary storage

## Document Types Supported

The system supports analysis of the following document types:

| Document Type         | Description                             | Key Fields Extracted                                                            |
| --------------------- | --------------------------------------- | ------------------------------------------------------------------------------- |
| `business_license`    | Business registration/license documents | Client name, tax code, business address, registered capital, incorporation date |
| `evn_bill`            | Electricity bills (EVN)                 | Billing address, billing amount                                                 |
| `rental_agreement`    | Property rental agreements              | Land ownership situation, signatory information, lease expiration               |
| `land_certificate`    | Land ownership certificates             | Land ownership situation, documentation completeness                            |
| `id_check`            | ID verification documents               | Company director name, key decision maker                                       |
| `financial_statement` | Financial statements                    | Revenue, costs, assets, debt information                                        |
| `site_visit_photos`   | Site visit documentation                | Company signboard status, account manager commentary                            |
| `cic_report`          | Credit Information Center reports       | Corporate history description                                                   |

## Data Model

The system extracts information into a structured `CustomerCheck` model with four main sections:

### Corporate Information

- **General**: Client details, business type, tax information, operations
- **History**: Incorporation date, corporate background
- **Relationship**: Source of client relationship
- **Ownership**: Owner names, ownership percentages, key personnel

### Land Information

- **EVN**: Electricity billing information and address verification
- **Ownership**: Land ownership status, lease agreements, documentation

### Financial Information

- **P&L**: Revenue, costs, and energy costs over time periods
- **Balance Sheet**: Assets and debt tracking
- **Loans**: Loan details, classifications, payment history

### Additional Information

- **Site Visit**: Company signboard verification and site observations

## Setup and Installation

#### Windows (PowerShell)

```powershell
# 0) Open PowerShell as Administrator

# 1) Install Go (winget) – skip if already installed
winget install -e --id GoLang.Go --silent

# 2) Install Poppler and Tesseract using winget (recommended)
winget install poppler
winget install tesseract

# 3) Add Poppler to PATH (winget installs to LOCALAPPDATA)
$popplerPath = "C:\Users\$env:USERNAME\AppData\Local\Microsoft\WinGet\Packages\oschwartz12.Poppler_Microsoft.Winget.Source_8wekyb3d8bbwe\poppler-25.07.0\Library\bin"
if (Test-Path $popplerPath) {
    [Environment]::SetEnvironmentVariable("PATH", "$([Environment]::GetEnvironmentVariable('PATH','User'));$popplerPath", "User")
    $env:PATH += ";$popplerPath"
    Write-Host "Added Poppler to PATH: $popplerPath"
} else {
    # Find the actual path if the default doesn't exist
    $popplerDir = Get-ChildItem -Path "$env:LOCALAPPDATA\Microsoft\WinGet\Packages" -Filter "*poppler*" -Directory | Select-Object -First 1 -ExpandProperty FullName
    if ($popplerDir) {
        $actualPath = Get-ChildItem -Path $popplerDir -Recurse -Filter "pdftoppm.exe" | Select-Object -First 1 -ExpandProperty DirectoryName
        if ($actualPath) {
            [Environment]::SetEnvironmentVariable("PATH", "$([Environment]::GetEnvironmentVariable('PATH','User'));$actualPath", "User")
            $env:PATH += ";$actualPath"
            Write-Host "Added Poppler to PATH: $actualPath"
        }
    }
}

# 4) Add Tesseract to PATH (winget installs to Program Files)
if (Test-Path "C:\Program Files\Tesseract-OCR\tesseract.exe") {
    [Environment]::SetEnvironmentVariable("PATH", "$([Environment]::GetEnvironmentVariable('PATH','User'));C:\Program Files\Tesseract-OCR", "User")
    $env:PATH += ";C:\Program Files\Tesseract-OCR"
    Write-Host "Added Tesseract to PATH"
}

# 5) Verify tools are accessible
Write-Host "Verifying tools..."
where pdftoppm
where tesseract

# 6) Build and run
cd D:\extraction
go mod download
go build -o extract.exe ./cmd/extractor

# 7) Test run (replace with your Drive file ID)
./extract.exe --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out test_output.xlsx --skip-analysis --progress
```

### Step 2: API Keys Setup

Create a `.env` file in the project root:

```env
GOOGLE_VISION_API_KEY=your_google_cloud_vision_api_key_here
GEMINI_API_KEY=your_gemini_api_key_here
```

**Get API Keys:**

- **Vision API**: [Google Cloud Console](https://console.cloud.google.com/) → Enable Vision API → Create API Key
- **Gemini API**: [Google AI Studio](https://makersuite.google.com/app/apikey) → Create API Key

### Step 3: Build and Test

```bash
# Build the application
go build -o extract.exe ./cmd/extractor

# Test with a Google Drive link (replace YOUR_FILE_ID)
./extract.exe --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out test_output.xlsx --skip-analysis --progress
```

## Usage Instructions

### Main Extractor (Full Pipeline)

```bash
# Basic usage with single Google Drive document
extract --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out results.xlsx --source business_license

# Process multiple Google Drive documents
extract --input "https://drive.google.com/file/d/FILE_ID_1/view" --input "https://drive.google.com/file/d/FILE_ID_2/view" --out results.xlsx

# Process from links file containing Google Drive URLs
extract --links-file documents.txt --out results.xlsx

# Advanced options with specific document sources
extract --file-source "https://drive.google.com/file/d/FILE_ID_1/view:business_license" --file-source "https://drive.google.com/file/d/FILE_ID_2/view:evn_bill" --out results.xlsx

# With AI analysis disabled (text extraction only)
extract --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out raw_text.xlsx --skip-analysis

# Enable grouping and validation
extract --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out results.xlsx --group --validate

# Export structured data as JSON
extract --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out results.xlsx --json data.json

# Control concurrency and show progress
extract --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out results.xlsx --concurrency 5 --progress
```

### Text-Only Mode (OCR without AI Analysis)

```bash
# Extract text without AI analysis
extract --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out extracted_text.xlsx --skip-analysis

# Process multiple Google Drive files in text-only mode
extract --input "https://drive.google.com/file/d/FILE_ID_1/view" --input "https://drive.google.com/file/d/FILE_ID_2/view" --out results.xlsx --skip-analysis

# From links file (text extraction only)
extract --links-file documents.txt --out results.xlsx --skip-analysis
```

### Command Line Options

#### Common Options

- `--input`: Google Drive URL or local path to document (repeatable)
- `--links-file`: Text file containing Google Drive URLs/paths (one per line)
- `--out`: Output file path (.xlsx format)
- `--lang`: OCR language codes (e.g., 'eng', 'eng+vie')
- `--dpi`: PDF rasterization DPI (default: 300)
- `--timeout`: Overall timeout in seconds

#### Extractor-Specific Options

- `--source`: Document source type (business_license, evn_bill, etc.)
- `--file-source`: File with specific document source format: 'file_path:source_type'
- `--skip-analysis`: Skip AI analysis (extract text only)
- `--concurrency`: Maximum concurrent files (default: 3)
- `--progress`: Show progress updates
- `--group`: Enable file grouping analysis
- `--validate`: Enable validation and quality checks
- `--json`: Export structured data as JSON

## Output Formats

### XLSX Files

- **Structured Data**: Multi-sheet Excel file with organized customer check data
- **Raw Data**: Single sheet with all extraction results and metadata

### JSON Export

- Complete structured customer check data in JSON format
- Includes all extracted fields with proper data types

### Processing Reports

- **Grouping Results**: JSON file with file grouping analysis
- **Validation Results**: JSON file with data quality assessment
- **Processing Statistics**: Console output with performance metrics

## Next Steps - Project Roadmap

### Phase 1: Data Integration (Current Priority)

- [ ] **Google Sheets Auto-Fill Integration**
  - Map extracted `CustomerCheck` model fields to Google Sheets columns
  - Implement Google Sheets API integration for automatic data population
  - Create field mapping configuration system
  - Test auto-fill with real Google Sheets templates

### Phase 2: Frontend Development

- [ ] **User Interface Design**

  - Create intuitive web interface for document upload
  - Design progress tracking dashboard
  - Implement drag-and-drop file upload for Google Drive links
  - Add document type selection interface
  - Create results preview and validation screens
  - Add Google Sheets template selection interface

- [ ] **Frontend Features**
  - Real-time processing status updates
  - Document preview and validation
  - Export options (XLSX, JSON, direct Google Sheets population)
  - Error handling and user feedback
  - Batch processing management

### Phase 3: Deployment & Production

- [ ] **Infrastructure Setup**

  - Deploy backend API to cloud platform (AWS/GCP/Azure)
  - Configure API rate limiting and security
  - Implement logging and monitoring

- [ ] **Frontend Deployment**
  - Deploy frontend to web hosting platform
  - Set up CI/CD pipeline for automatic deployments
