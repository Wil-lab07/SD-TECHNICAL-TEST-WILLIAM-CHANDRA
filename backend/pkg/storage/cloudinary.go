package storage

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var (
	ErrInvalidFileType = errors.New("only image files are allowed (jpg, jpeg, png, webp)")
	ErrFileTooLarge    = errors.New("file size exceeds 5MB limit")
)

type Storage interface {
	UploadImage(ctx context.Context, file multipart.File, filename string) (string, error)
}

type CloudinaryStorage struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cloudinaryURL string) (*CloudinaryStorage, error) {
	if cloudinaryURL == "" {
		return nil, nil
	}
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("cloudinary init error: %w", err)
	}
	return &CloudinaryStorage{cld: cld}, nil
}

func ValidateImage(header *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return ErrInvalidFileType
	}
	if header.Size > 5*1024*1024 {
		return ErrFileTooLarge
	}
	return nil
}

func (s *CloudinaryStorage) UploadImage(ctx context.Context, file multipart.File, filename string) (string, error) {
	if s == nil || s.cld == nil {
		return "", errors.New("cloudinary storage is not configured")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return "", ErrInvalidFileType
	}

	resp, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "team_logos",
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return resp.SecureURL, nil
}
