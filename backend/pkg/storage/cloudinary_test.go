package storage_test

import (
	"mime/multipart"
	"testing"

	"football-api/pkg/storage"
)

func TestValidateImage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		size     int64
		wantErr  error
	}{
		{
			name:     "valid jpg image",
			filename: "logo.jpg",
			size:     1024 * 1024,
			wantErr:  nil,
		},
		{
			name:     "valid png image",
			filename: "logo.png",
			size:     2 * 1024 * 1024,
			wantErr:  nil,
		},
		{
			name:     "invalid file extension .exe",
			filename: "malicious.exe",
			size:     1024,
			wantErr:  storage.ErrInvalidFileType,
		},
		{
			name:     "invalid file extension .pdf",
			filename: "doc.pdf",
			size:     1024,
			wantErr:  storage.ErrInvalidFileType,
		},
		{
			name:     "file size exceeding 5MB limit",
			filename: "large.png",
			size:     6 * 1024 * 1024,
			wantErr:  storage.ErrFileTooLarge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			header := &multipart.FileHeader{
				Filename: tc.filename,
				Size:     tc.size,
			}
			err := storage.ValidateImage(header)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Fatalf("ValidateImage() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateImage() unexpected error: %v", err)
			}
		})
	}
}
