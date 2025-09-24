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

- **`go.mod`** - Go module definition and dependencies
- **`go.sum`** - Go module checksums for dependency verification
- **`README.md`** - This system documentation
- **`sample_links.txt`** - Sample input file containing document URLs for testing

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

### Step 1: Prerequisites

Install the required dependencies:

#### Required Software

- **Go 1.22 or higher**: [Download from golang.org](https://golang.org/dl/)
- **Poppler tools**: For PDF processing
  - **Windows**: Download from [poppler-windows](https://github.com/oschwartz10612/poppler-windows) or use `choco install poppler`
  - **macOS**: `brew install poppler`
  - **Linux**: `sudo apt-get install poppler-utils` (Ubuntu/Debian) or `sudo yum install poppler-utils` (CentOS/RHEL)

#### Optional Software

- **Tesseract OCR**: For fallback OCR processing
  - **Windows**: Download from [UB Mannheim](https://github.com/UB-Mannheim/tesseract/wiki)
  - **macOS**: `brew install tesseract`
  - **Linux**: `sudo apt-get install tesseract-ocr` (Ubuntu/Debian)

### Step 2: Clone and Setup Project

```bash
# Clone the repository
git clone <your-repository-url>
cd extraction

# Install Go dependencies
go mod download
```

### Step 3: API Keys Setup

You need to obtain API keys from Google:

#### Google Cloud Vision API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the "Vision API" service
4. Create credentials (API Key)
5. Copy your API key

#### Google Gemini API Key

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Copy your API key

### Step 4: Environment Configuration

Create a `.env` file in the project root:

```env
GOOGLE_VISION_API_KEY=your_google_cloud_vision_api_key_here
GEMINI_API_KEY=your_gemini_api_key_here
```

**Important**:

- Replace the placeholder values with your actual API keys
- Never commit the `.env` file to version control
- Add `.env` to your `.gitignore` file

### Step 5: Build the Application

```bash
# Build the main extractor application
go build -o extract.exe ./cmd/extractor
```

### Step 6: Test Your Installation

```bash
# Test with a Google Drive link
./extract.exe --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out test_output.xlsx --source business_license

# Or test in text-only mode (no AI analysis)
./extract.exe --input "https://drive.google.com/file/d/YOUR_FILE_ID/view" --out test_output.xlsx --skip-analysis
```

**Note**:

- Replace `YOUR_FILE_ID` with the actual file ID from your Google Drive link
- Google Drive links should be in format: `https://drive.google.com/file/d/FILE_ID/view`
- Make sure the Google Drive files are publicly accessible or you have the proper permissions

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
  - Set up database for processing history and user management
  - Configure API rate limiting and security
  - Implement logging and monitoring

- [ ] **Frontend Deployment**
  - Deploy frontend to web hosting platform
  - Set up CI/CD pipeline for automatic deployments
  - Configure domain and SSL certificates
  - Implement user authentication and access control

### Phase 4: Team Integration

- [ ] **Company Team Onboarding**

  - Create user training materials and documentation
  - Set up team access and permissions
  - Implement audit trails for document processing
  - Create support and troubleshooting guides

- [ ] **Production Optimization**
  - Performance monitoring and optimization
  - User feedback collection and feature improvements
  - Scaling infrastructure based on usage
  - Regular maintenance and updates

### Technical Implementation Notes

#### Google Sheets Integration

```go
// Example field mapping structure needed
type SheetFieldMapping struct {
    CustomerCheckField string `json:"customer_check_field"`
    GoogleSheetColumn  string `json:"google_sheet_column"` // e.g., "A1", "B2", "Client Name"
    FieldType         string `json:"field_type"`          // text, number, date, currency
    Required          bool   `json:"required"`
    SheetTemplate     string `json:"sheet_template"`       // Template sheet ID
}

// Example mapping configuration
type SheetMappingConfig struct {
    TemplateSheetID string            `json:"template_sheet_id"`
    FieldMappings   []SheetFieldMapping `json:"field_mappings"`
}
```
