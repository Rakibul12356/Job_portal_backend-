package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rakib/job-portal-api/internal/config"
)

type StorageService interface {
	SaveUploadedFile(file *multipart.FileHeader, category string, ownerID string) (string, error)
	DeleteFile(fileURL string) error
}

type localStorageService struct {
	baseDir string
	baseURL string
}

func NewStorageService() StorageService {
	cfg := config.AppConfig
	return &localStorageService{
		baseDir: cfg.UploadDir, // e.g. "./uploads"
		baseURL: cfg.AppBaseURL, // e.g. "http://localhost:8080"
	}
}

func (s *localStorageService) SaveUploadedFile(file *multipart.FileHeader, category string, ownerID string) (string, error) {
	// 1. Validate size and type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if err := s.validateFile(file.Size, ext, category); err != nil {
		return "", err
	}

	// 2. Set directory layout according to category
	// resumes -> uploads/resumes/{userId}/{uuid}.pdf
	// avatars -> uploads/avatars/{userId}/{uuid}.jpg
	// logos   -> uploads/logos/{companyId}/{uuid}.png
	var subFolder string
	switch category {
	case "resume":
		subFolder = filepath.Join("resumes", ownerID)
	case "avatar":
		subFolder = filepath.Join("avatars", ownerID)
	case "logo":
		subFolder = filepath.Join("logos", ownerID)
	default:
		return "", errors.New("invalid upload category")
	}

	targetDir := filepath.Join(s.baseDir, subFolder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directories: %v", err)
	}

	// Generate clean filename
	newFilename := uuid.New().String() + ext
	targetPath := filepath.Join(targetDir, newFilename)

	// Save file locally
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return "", err
	}

	// Return URL path: e.g. "http://localhost:8080/uploads/resumes/123/uuid.pdf"
	// Replace backslashes for clean URL paths
	urlPath := fmt.Sprintf("/uploads/%s/%s", strings.ReplaceAll(subFolder, "\\", "/"), newFilename)
	return s.baseURL + urlPath, nil
}

func (s *localStorageService) DeleteFile(fileURL string) error {
	if fileURL == "" || !strings.Contains(fileURL, "/uploads/") {
		return nil
	}

	parts := strings.Split(fileURL, "/uploads/")
	if len(parts) < 2 {
		return nil
	}

	relPath := parts[1]
	localPath := filepath.Join(s.baseDir, relPath)

	// Check if file exists before deleting
	if _, err := os.Stat(localPath); err == nil {
		return os.Remove(localPath)
	}

	return nil
}

func (s *localStorageService) validateFile(size int64, ext string, category string) error {
	switch category {
	case "resume":
		// Max 5MB, PDF, DOC, DOCX
		if size > 5*1024*1024 {
			return errors.New("resume file size exceeds 5MB limit")
		}
		if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
			return errors.New("unsupported resume file format; only PDF, DOC, DOCX allowed")
		}
	case "avatar":
		// Max 5MB, JPG, JPEG, PNG, GIF
		if size > 5*1024*1024 {
			return errors.New("avatar file size exceeds 5MB limit")
		}
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			return errors.New("unsupported avatar image format; only JPG, PNG, GIF allowed")
		}
	case "logo":
		// Max 2MB, JPG, JPEG, PNG, SVG
		if size > 2*1024*1024 {
			return errors.New("logo file size exceeds 2MB limit")
		}
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".svg" {
			return errors.New("unsupported logo image format; only JPG, PNG, SVG allowed")
		}
	default:
		return errors.New("unknown upload validation category")
	}

	return nil
}
