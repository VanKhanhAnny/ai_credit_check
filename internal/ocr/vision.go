package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ExtractTextFromImageVision performs OCR using Google Cloud Vision's DOCUMENT_TEXT_DETECTION.
// Requires the environment variable GOOGLE_VISION_API_KEY to be set.
func ExtractTextFromImageVision(ctx context.Context, imagePath string, lang string) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("GOOGLE_VISION_API_KEY"))
	if apiKey == "" {
		return "", errors.New("GOOGLE_VISION_API_KEY is not set; set it in your environment or .env")
	}
	if imagePath == "" {
		return "", errors.New("image path is empty")
	}

	content, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("read image: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(content)

	langHints := tesseractLangToBCP47Hints(lang)

	req := visionAnnotateRequest{
		Requests: []visionSingleRequest{
			{
				Image: visionImage{Content: b64},
				Features: []visionFeature{{Type: "DOCUMENT_TEXT_DETECTION"}},
				ImageContext: &visionImageContext{LanguageHints: langHints},
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := "https://vision.googleapis.com/v1/images:annotate?key=" + apiKey
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("vision request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("vision http error: %s", resp.Status)
	}

	var vr visionAnnotateResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&vr); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(vr.Responses) == 0 {
		return "", errors.New("vision: empty response")
	}
	res := vr.Responses[0]
	if res.Error.Message != "" {
		return "", fmt.Errorf("vision error: %s", res.Error.Message)
	}
	if res.FullTextAnnotation.Text != "" {
		return res.FullTextAnnotation.Text, nil
	}
	if len(res.TextAnnotations) > 0 && strings.TrimSpace(res.TextAnnotations[0].Description) != "" {
		return res.TextAnnotations[0].Description, nil
	}
	return "", nil
}

func tesseractLangToBCP47Hints(lang string) []string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return nil
	}
	parts := strings.FieldsFunc(lang, func(r rune) bool { return r == '+' || r == ',' || r == ';' || r == ' ' })
	var hints []string
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		switch p {
		case "eng":
			hints = append(hints, "en")
		case "vie", "vin":
			hints = append(hints, "vi")
		case "jpn":
			hints = append(hints, "ja")
		case "zho", "chi_sim", "chi_tra":
			hints = append(hints, "zh")
		case "spa":
			hints = append(hints, "es")
		case "fra", "fre":
			hints = append(hints, "fr")
		case "deu", "ger":
			hints = append(hints, "de")
		case "ita":
			hints = append(hints, "it")
		case "rus":
			hints = append(hints, "ru")
		case "ara":
			hints = append(hints, "ar")
		case "hin":
			hints = append(hints, "hi")
		case "tha":
			hints = append(hints, "th")
		case "kor":
			hints = append(hints, "ko")
		case "por":
			hints = append(hints, "pt")
		default:
			// If already a plausible BCP-47 code (2-3 letters), pass through
			if len(p) == 2 || len(p) == 3 {
				hints = append(hints, p)
			}
		}
	}
	return hints
}

// --- Minimal response/request structs ---

type visionAnnotateRequest struct {
	Requests []visionSingleRequest `json:"requests"`
}

type visionSingleRequest struct {
	Image        visionImage        `json:"image"`
	Features     []visionFeature    `json:"features"`
	ImageContext *visionImageContext `json:"imageContext,omitempty"`
}

type visionImage struct {
	Content string `json:"content"`
}

type visionFeature struct {
	Type       string `json:"type"`
	MaxResults int    `json:"maxResults,omitempty"`
}

type visionImageContext struct {
	LanguageHints []string `json:"languageHints,omitempty"`
}

type visionAnnotateResponse struct {
	Responses []visionSingleResponse `json:"responses"`
}

type visionSingleResponse struct {
	FullTextAnnotation struct {
		Text string `json:"text"`
	} `json:"fullTextAnnotation"`
	TextAnnotations []struct {
		Description string `json:"description"`
	} `json:"textAnnotations"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ExtractTextFromPDFVision renders PDF pages to images and OCRs them via Vision.
// Requires Poppler's pdftoppm on PATH.
func ExtractTextFromPDFVision(ctx context.Context, pdfPath string, lang string, dpi int) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("GOOGLE_VISION_API_KEY"))
	if apiKey == "" {
		return "", errors.New("GOOGLE_VISION_API_KEY is not set; set it in your environment or .env")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-ocr-vision-*")
	if err != nil {
		return "", fmt.Errorf("ocr fallback: mkdir temp: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if dpi <= 0 {
		dpi = 300
	}
	prefix := filepath.Join(tmpDir, "page")
	cmd := exec.CommandContext(ctx, "pdftoppm", "-r", fmt.Sprintf("%d", dpi), "-png", pdfPath, prefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftoppm error: %v: %s", err, strings.TrimSpace(stderr.String()))
	}

	images, err := filepath.Glob(prefix + "-*.png")
	if err != nil {
		return "", fmt.Errorf("ocr fallback: glob images: %w", err)
	}
	if len(images) == 0 {
		return "", fmt.Errorf("ocr fallback: no images produced from PDF")
	}
	sort.Strings(images)

	var b strings.Builder
	for _, img := range images {
		text, err := ExtractTextFromImageVision(ctx, img, lang)
		if err != nil {
			continue
		}
		if s := strings.TrimSpace(text); s != "" {
			if b.Len() > 0 {
				b.WriteString("\n\n")
			}
			b.WriteString(s)
		}
	}
	return b.String(), nil
}

// ExtractTextFromImageTesseract performs OCR using Tesseract as a fallback when Vision API fails
func ExtractTextFromImageTesseract(ctx context.Context, imagePath string, lang string) (string, error) {
	if imagePath == "" {
		return "", errors.New("image path is empty")
	}

	// Convert Tesseract language codes
	tesseractLang := "eng" // default
	if lang != "" {
		parts := strings.FieldsFunc(lang, func(r rune) bool { return r == '+' || r == ',' || r == ';' || r == ' ' })
		if len(parts) > 0 {
			switch strings.ToLower(parts[0]) {
			case "vie", "vin":
				tesseractLang = "vie"
			case "jpn":
				tesseractLang = "jpn"
			case "chi_sim":
				tesseractLang = "chi_sim"
			case "chi_tra":
				tesseractLang = "chi_tra"
			case "spa":
				tesseractLang = "spa"
			case "fra", "fre":
				tesseractLang = "fra"
			case "deu", "ger":
				tesseractLang = "deu"
			case "ita":
				tesseractLang = "ita"
			case "rus":
				tesseractLang = "rus"
			case "ara":
				tesseractLang = "ara"
			case "hin":
				tesseractLang = "hin"
			case "tha":
				tesseractLang = "tha"
			case "kor":
				tesseractLang = "kor"
			case "por":
				tesseractLang = "por"
			default:
				tesseractLang = "eng"
			}
		}
	}

	// Run Tesseract command
	cmd := exec.CommandContext(ctx, "tesseract", imagePath, "stdout", "-l", tesseractLang)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %v, stderr: %s", err, stderr.String())
	}

	text := strings.TrimSpace(stdout.String())
	if text == "" {
		return "", errors.New("no text extracted from image")
	}

	return text, nil
}


