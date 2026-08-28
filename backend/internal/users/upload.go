package users

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func saveAvatar(
	file io.Reader,
	header []byte,
	contentType string,
) (string, error) {
	var extension string

	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		return "", os.ErrInvalid
	}

	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	filename := hex.EncodeToString(randomBytes) + extension

	avatarDir := "uploads/avatars"

	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(avatarDir, filename)

	output, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer output.Close()

	if _, err := output.Write(header); err != nil {
		os.Remove(filePath)
		return "", err
	}

	if _, err := io.Copy(output, file); err != nil {
		os.Remove(filePath)
		return "", err
	}

	return "/uploads/avatars/" + filename, nil
}

func detectImageType(file io.Reader) (string, []byte, error) {
	header := make([]byte, 512)

	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	if n == 0 {
		return "", nil, io.ErrUnexpectedEOF
	}

	contentType := http.DetectContentType(header[:n])

	return contentType, header[:n], nil
}


func deleteAvatarFile(avatarURL string) {
	if avatarURL == "" {
		return
	}

	const prefix = "/uploads/avatars/"

	if !strings.HasPrefix(avatarURL, prefix) {
		return
	}

	filename := strings.TrimPrefix(avatarURL, prefix)

	if strings.Contains(filename, "/") ||
		strings.Contains(filename, "\\") ||
		filename == "" {
		return
	}

	filePath := filepath.Join(
		"uploads",
		"avatars",
		filename,
	)

	_ = os.Remove(filePath)
}
