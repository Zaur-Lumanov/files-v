package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FormatSize форматирует размер файла в человекочитаемый формат
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// IsHidden проверяет, является ли файл скрытым
func IsHidden(path string) bool {
	name := filepath.Base(path)
	return strings.HasPrefix(name, ".")
}

// EnsureDir убеждается, что директория существует
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists проверяет существование файла
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
