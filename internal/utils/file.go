package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".tiff": true,
	".ico":  true,
	".webp": true,
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/bmp":  true,
	"image/tiff": true,
	"image/ico":  true,
	"image/webp": true,
}

func ValidateAndSaveFile(fileHeader *multipart.FileHeader, uploadDir string) (string, error) {
	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// Check if the file extension is allowed
	if !allowedExtensions[extension] {
		return "", fmt.Errorf("Invalid file extension: %s", extension)
	}

	// Check size of the file
	if fileHeader.Size > 10<<20 {
		return "", fmt.Errorf("File size must be less than 10MB")
	}

	// Check content type of the file
	file, err := fileHeader.Open()

	if err != nil {
		return "", fmt.Errorf("Failed to open file: %w", err)
	}

	defer file.Close()

	// Read the file to detect the mime type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)

	if err != nil {
		return "", fmt.Errorf("Failed to read file: %w", err)
	}

	// Detect the mime type of the file
	mimeType := http.DetectContentType(buffer)

	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("Invalid mime type: %s", mimeType)
	}

	// Generate a unique filename
	fileNameWithoutExtension := strings.TrimSuffix(filepath.Base(fileHeader.Filename), filepath.Ext(fileHeader.Filename))
	uniqueFilename := fmt.Sprintf("%s-%s", fileNameWithoutExtension, uuid.New().String())
	uniqueFilenameWithExtension := fmt.Sprintf("%s%s", uniqueFilename, extension)

	// Replace all spaces with hyphens
	uniqueFilenameWithExtension = strings.ReplaceAll(uniqueFilenameWithExtension, " ", "_")

	// os.ModePerm is 0777
	// 0777 means that the directory and all files in it can be read, written, and executed by anyone
	err = os.MkdirAll(uploadDir, os.ModePerm)

	if err != nil {
		return "", fmt.Errorf("Failed to create uploads directory: %w", err)
	}

	// Save file to local storage
	destinationPath := fmt.Sprintf("./uploads/%s", uniqueFilenameWithExtension)

	if err := SaveFile(fileHeader, destinationPath); err != nil {
		return "", fmt.Errorf("Failed to save file: %w", err)
	}

	return destinationPath, nil
}

func SaveFile(fileHeader *multipart.FileHeader, destinationPath string) error {
	// Open the source file
	src, err := fileHeader.Open()

	if err != nil {
		return err
	}

	defer src.Close()

	// Create the destination file
	out, err := os.Create(destinationPath)

	if err != nil {
		return err
	}

	defer out.Close()

	// Copy the file to the destination file
	_, err = io.Copy(out, src)

	return err
}
