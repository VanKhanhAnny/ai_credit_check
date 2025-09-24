package xfer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadToTemp downloads a local path or URL into a temp file.
// Returns: localPath, sourceURL, filename (best-effort), mediaType, error
func DownloadToTemp(ctx context.Context, input string) (string, string, string, string, error) {
	if input == "" {
		return "", "", "", "", errors.New("empty input")
	}
	// Local file path
	if fileExists(input) {
		fi, _ := os.Stat(input)
		mt := mime.TypeByExtension(strings.ToLower(filepath.Ext(input)))
		return input, input, fi.Name(), mt, nil
	}

	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" {
		return "", input, "", "", fmt.Errorf("not a file and not a valid URL: %s", input)
	}

	resolvedURL := normalizeGoogleDrive(u)
	client := &http.Client{Timeout: 60 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, resolvedURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", resolvedURL, "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resolvedURL, "", "", fmt.Errorf("http %d", resp.StatusCode)
	}

	mediaType := resp.Header.Get("Content-Type")
	filename := filenameFromHeaders(resp.Header.Get("Content-Disposition"))
	if filename == "" {
		filename = filepath.Base(u.Path)
	}
	if filename == "" {
		filename = "downloaded"
	}

	f, err := os.CreateTemp("", "xfer-*"+filepath.Ext(filename))
	if err != nil {
		return "", resolvedURL, filename, mediaType, err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", resolvedURL, filename, mediaType, err
	}
	return f.Name(), resolvedURL, filename, mediaType, nil
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

func filenameFromHeaders(contentDisposition string) string {
	if contentDisposition == "" {
		return ""
	}
	// naive parsing for filename=...
	parts := strings.Split(contentDisposition, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "filename=") {
			v := strings.TrimPrefix(p, "filename=")
			v = strings.Trim(v, "\"'")
			return v
		}
	}
	return ""
}

func normalizeGoogleDrive(u *url.URL) string {
	// Handle common share links:
	// https://drive.google.com/file/d/<id>/view?usp=sharing
	// https://drive.google.com/uc?id=<id>&export=download
	if u.Host == "drive.google.com" {
		path := strings.Trim(u.Path, "/")
		parts := strings.Split(path, "/")
		for i := 0; i < len(parts)-1; i++ {
			if parts[i] == "file" && parts[i+1] == "d" && i+2 < len(parts) {
				id := parts[i+2]
				return "https://drive.google.com/uc?id=" + id + "&export=download"
			}
		}
		if u.Path == "/uc" {
			q := u.Query()
			if q.Get("id") != "" {
				return u.String()
			}
		}
	}
	return u.String()
}



