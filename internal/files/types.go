package files

import "path/filepath"

type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeImage
	FileTypeText
	FileTypePDF
	FileTypeWord
	FileTypeExcel
	FileTypePowerPoint
	FileTypeArchive
	FileTypeVideo
	FileTypeAudio
)

func (t FileType) String() string {
	switch t {
	case FileTypeImage:
		return "image"
	case FileTypeText:
		return "text"
	case FileTypePDF:
		return "pdf"
	case FileTypeWord:
		return "word"
	case FileTypeExcel:
		return "excel"
	case FileTypePowerPoint:
		return "powerpoint"
	case FileTypeArchive:
		return "archive"
	case FileTypeVideo:
		return "video"
	case FileTypeAudio:
		return "audio"
	default:
		return "unknown"
	}
}

func DetectFileType(filename string, mediaType string) FileType {
	if mediaType != "" {
		if hasPrefix(mediaType, "image/") {
			return FileTypeImage
		}
		// Handle common image MIME types that might not have image/ prefix
		if mediaType == "image/jpeg" || mediaType == "image/jpg" || mediaType == "image/png" || 
		   mediaType == "image/gif" || mediaType == "image/bmp" || mediaType == "image/webp" ||
		   mediaType == "image/tiff" || mediaType == "image/svg+xml" || mediaType == "image/heic" ||
		   mediaType == "image/heif" || mediaType == "image/avif" || mediaType == "image/jxl" ||
		   mediaType == "image/vnd.microsoft.icon" || mediaType == "image/x-icon" {
			return FileTypeImage
		}
		if mediaType == "text/plain" {
			return FileTypeText
		}
		if mediaType == "application/pdf" {
			return FileTypePDF
		}
		if hasPrefix(mediaType, "application/vnd.openxmlformats-officedocument.wordprocessingml") ||
			mediaType == "application/msword" {
			return FileTypeWord
		}
		if hasPrefix(mediaType, "application/vnd.openxmlformats-officedocument.spreadsheetml") ||
			mediaType == "application/vnd.ms-excel" {
			return FileTypeExcel
		}
		if hasPrefix(mediaType, "application/vnd.openxmlformats-officedocument.presentationml") ||
			mediaType == "application/vnd.ms-powerpoint" {
			return FileTypePowerPoint
		}
		if hasPrefix(mediaType, "video/") {
			return FileTypeVideo
		}
		if hasPrefix(mediaType, "audio/") {
			return FileTypeAudio
		}
		if hasPrefix(mediaType, "application/zip") ||
			hasPrefix(mediaType, "application/x-rar") ||
			hasPrefix(mediaType, "application/x-7z") {
			return FileTypeArchive
		}
	}
	
	ext := filepath.Ext(filename)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".tif", ".tiff", ".bmp", ".gif", ".webp", ".heic", ".svg", ".jfif", ".pjpeg", ".pjp", ".ico", ".cur", ".tga", ".psd", ".raw", ".cr2", ".nef", ".orf", ".sr2", ".dng", ".arw", ".rw2", ".pef", ".srw", ".x3f", ".mrw", ".raf", ".dcr", ".kdc", ".erf", ".mef", ".iiq", ".3fr", ".fff", ".hdr", ".exr", ".dds", ".ktx", ".pkm", ".pvr", ".astc":
		return FileTypeImage
	case ".txt", ".md", ".csv", ".json", ".xml", ".yaml", ".yml":
		return FileTypeText
	case ".pdf":
		return FileTypePDF
	case ".doc", ".docx":
		return FileTypeWord
	case ".xls", ".xlsx", ".xlsm":
		return FileTypeExcel
	case ".ppt", ".pptx", ".pptm":
		return FileTypePowerPoint
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2":
		return FileTypeArchive
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv":
		return FileTypeVideo
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return FileTypeAudio
	default:
		return FileTypeUnknown
	}
}

// IsProcessableFileType returns true if the file type can be processed for text extraction
func IsProcessableFileType(fileType FileType) bool {
	switch fileType {
	case FileTypeImage, FileTypeText, FileTypePDF, FileTypeWord, FileTypeExcel, FileTypePowerPoint:
		return true
	default:
		return false
	}
}

func hasPrefix(s, prefix string) bool { return len(s) >= len(prefix) && s[:len(prefix)] == prefix }



