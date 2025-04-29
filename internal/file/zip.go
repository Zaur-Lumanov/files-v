package file

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	// EncryptedZipExtension расширение зашифрованного ZIP-архива
	EncryptedZipExtension = ".zip.efv"
)

// IsZipFile проверяет, является ли файл ZIP-архивом
func IsZipFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".zip")
}

// IsEncryptedZipFile проверяет, является ли файл зашифрованным ZIP-архивом
func IsEncryptedZipFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), EncryptedZipExtension)
}

// CreateZipArchive создает ZIP-архив из указанной директории
func CreateZipArchive(sourceDir, targetZip string) error {
	// Проверяем, существует ли исходная директория
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		return errors.Wrap(err, "ошибка доступа к исходной директории")
	}
	if !sourceInfo.IsDir() {
		return errors.New("указанный путь не является директорией")
	}

	// Создаем ZIP-файл
	zipFile, err := os.Create(targetZip)
	if err != nil {
		return errors.Wrap(err, "ошибка создания ZIP-файла")
	}
	defer zipFile.Close()

	// Создаем новый ZIP-архив
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// Обходим все файлы в директории и добавляем их в архив
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Получаем относительный путь для записи в архив
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return errors.Wrap(err, "ошибка получения относительного пути")
		}
		if relPath == "." {
			return nil // Пропускаем корневую директорию
		}

		// Подготавливаем запись в архиве
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errors.Wrap(err, "ошибка создания заголовка для файла")
		}

		// Устанавливаем относительный путь как имя в архиве
		header.Name = relPath

		// Устанавливаем метод сжатия
		if !info.IsDir() {
			header.Method = zip.Deflate
		}

		// Создаем запись в архиве
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return errors.Wrap(err, "ошибка создания записи в архиве")
		}

		// Если это директория, просто продолжаем
		if info.IsDir() {
			return nil
		}

		// Открываем файл для чтения
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "ошибка открытия файла для архивирования")
		}
		defer file.Close()

		// Копируем содержимое файла в архив
		_, err = io.Copy(writer, file)
		return errors.Wrap(err, "ошибка записи файла в архив")
	})

	return err
}

// EncryptZipFile шифрует ZIP-архив и сохраняет результат с расширением .zip.efv
func EncryptZipFile(zipPath string) (string, error) {
	// Проверяем, является ли файл ZIP-архивом
	if !IsZipFile(zipPath) {
		return "", errors.New("указанный файл не является ZIP-архивом")
	}

	// Формируем имя зашифрованного файла
	encryptedPath := zipPath + ".efv"

	// Читаем содержимое ZIP-архива
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return "", errors.Wrap(err, "ошибка чтения ZIP-архива")
	}

	// Шифруем данные
	encryptedData, err := EncryptData(zipData)
	if err != nil {
		return "", errors.Wrap(err, "ошибка шифрования ZIP-архива")
	}

	// Записываем зашифрованные данные в файл
	err = os.WriteFile(encryptedPath, encryptedData, 0644)
	if err != nil {
		return "", errors.Wrap(err, "ошибка записи зашифрованного ZIP-архива")
	}

	return encryptedPath, nil
}

// DecryptZipFile расшифровывает ZIP-архив с расширением .zip.efv
func DecryptZipFile(encryptedPath string) (string, error) {
	// Проверяем, является ли файл зашифрованным ZIP-архивом
	if !IsEncryptedZipFile(encryptedPath) {
		return "", errors.New("указанный файл не является зашифрованным ZIP-архивом")
	}

	// Формируем имя расшифрованного файла (удаляем .efv)
	decryptedPath := strings.TrimSuffix(encryptedPath, ".efv")

	// Читаем содержимое зашифрованного файла
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		return "", errors.Wrap(err, "ошибка чтения зашифрованного файла")
	}

	// Расшифровываем данные
	zipData, err := DecryptData(encryptedData)
	if err != nil {
		return "", errors.Wrap(err, "ошибка расшифровки данных")
	}

	// Записываем расшифрованные данные в ZIP-файл
	err = os.WriteFile(decryptedPath, zipData, 0644)
	if err != nil {
		return "", errors.Wrap(err, "ошибка записи расшифрованного ZIP-архива")
	}

	return decryptedPath, nil
}

// ExtractZipArchive распаковывает ZIP-архив в указанную директорию
func ExtractZipArchive(zipPath, destDir string) error {
	// Открываем ZIP-архив для чтения
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return errors.Wrap(err, "ошибка открытия ZIP-архива")
	}
	defer reader.Close()

	// Создаем целевую директорию, если она не существует
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Wrap(err, "ошибка создания целевой директории")
	}

	// Извлекаем каждый файл из архива
	for _, file := range reader.File {
		// Полный путь для извлечения файла
		path := filepath.Join(destDir, file.Name)

		// Проверяем пути (защита от Zip Slip)
		if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("некорректный путь в архиве: %s", file.Name)
		}

		// Если это директория, создаем ее и продолжаем
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return errors.Wrap(err, "ошибка создания директории из архива")
			}
			continue
		}

		// Создаем все родительские директории
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return errors.Wrap(err, "ошибка создания родительских директорий")
		}

		// Создаем файл
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return errors.Wrap(err, "ошибка создания файла")
		}

		// Открываем архивированный файл для чтения
		inFile, err := file.Open()
		if err != nil {
			outFile.Close()
			return errors.Wrap(err, "ошибка открытия файла в архиве")
		}

		// Копируем содержимое
		_, err = io.Copy(outFile, inFile)
		inFile.Close()
		outFile.Close()
		if err != nil {
			return errors.Wrap(err, "ошибка копирования содержимого файла")
		}
	}

	return nil
}

// ProcessDirectoryWithZip обрабатывает директорию с использованием ZIP
// Создает ZIP-архив, шифрует его и сохраняет с расширением .zip.efv
func ProcessDirectoryWithZip(sourceDir string) (string, error) {
	// Формируем имя временного ZIP-архива
	tempZipPath := sourceDir + ".zip"

	// Создаем ZIP-архив
	if err := CreateZipArchive(sourceDir, tempZipPath); err != nil {
		return "", errors.Wrap(err, "ошибка создания ZIP-архива")
	}

	// Шифруем ZIP-архив
	encryptedPath, err := EncryptZipFile(tempZipPath)
	if err != nil {
		// Удаляем временный ZIP-архив при ошибке
		os.Remove(tempZipPath)
		return "", errors.Wrap(err, "ошибка шифрования ZIP-архива")
	}

	// Удаляем временный ZIP-архив
	os.Remove(tempZipPath)

	return encryptedPath, nil
}

// ProcessEncryptedZipFile обрабатывает зашифрованный ZIP-архив
// Расшифровывает его и распаковывает в указанную директорию
func ProcessEncryptedZipFile(encryptedPath string) (string, error) {
	// Формируем имя директории для распаковки (удаляем .zip.efv)
	extractDir := strings.TrimSuffix(encryptedPath, EncryptedZipExtension)

	// Расшифровываем ZIP-архив
	decryptedZipPath, err := DecryptZipFile(encryptedPath)
	if err != nil {
		return "", errors.Wrap(err, "ошибка расшифровки ZIP-архива")
	}

	// Распаковываем ZIP-архив
	if err := ExtractZipArchive(decryptedZipPath, extractDir); err != nil {
		// Удаляем временный расшифрованный ZIP-архив при ошибке
		os.Remove(decryptedZipPath)
		return "", errors.Wrap(err, "ошибка распаковки ZIP-архива")
	}

	// Удаляем временный расшифрованный ZIP-архив
	os.Remove(decryptedZipPath)

	return extractDir, nil
}
