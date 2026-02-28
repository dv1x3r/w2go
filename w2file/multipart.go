package w2file

import (
	"errors"
	"mime/multipart"
	"net/http"
)

const defaultMemory = 32 << 20        // 32 MB
const defaultMaxUploadSize = 32 << 20 // 32 MB

type ParseMultipartFilesOptions struct {
	Memory        int64
	MaxUploadSize int64
}

func ParseMultipartFiles(r *http.Request) ([]*multipart.FileHeader, error) {
	return ParseMultipartFilesWithOptions(r, ParseMultipartFilesOptions{})
}

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
