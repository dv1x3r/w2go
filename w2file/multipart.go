package w2file

import (
	"errors"
	"mime/multipart"
	"net/http"
)

const defaultMemory = 32 << 20        // 32 MB
const defaultMaxUploadSize = 32 << 20 // 32 MB

// ParseMultipartFilesOptions customizes multipart parsing limits.
type ParseMultipartFilesOptions struct {
	// Memory is the maximum memory buffer passed to Request.ParseMultipartForm.
	// It defaults to 32 MiB.
	Memory int64

	// MaxUploadSize is the maximum allowed size for each uploaded file.
	// It defaults to 32 MiB.
	MaxUploadSize int64
}

// ParseMultipartFiles parses files from the "files[]" multipart field using default limits.
func ParseMultipartFiles(r *http.Request) ([]*multipart.FileHeader, error) {
	return ParseMultipartFilesWithOptions(r, ParseMultipartFilesOptions{})
}

// ParseMultipartFilesWithOptions parses files from the "files[]" multipart
// field and checks each file against options.MaxUploadSize.
func ParseMultipartFilesWithOptions(r *http.Request, options ParseMultipartFilesOptions) ([]*multipart.FileHeader, error) {
	memory := options.Memory
	if memory == 0 {
		memory = defaultMemory
	}
	maxUploadSize := options.MaxUploadSize
	if maxUploadSize == 0 {
		maxUploadSize = defaultMaxUploadSize
	}

	if err := r.ParseMultipartForm(memory); err != nil {
		return nil, err
	}

	headers := r.MultipartForm.File["files[]"]
	for _, header := range headers {
		if header.Size > maxUploadSize {
			return nil, errors.New(http.StatusText(http.StatusRequestEntityTooLarge))
		}
	}

	return headers, nil
}
