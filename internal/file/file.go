package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileInfo содержит информацию о файле
type FileInfo struct {
	Path  string
	Size  int64
	Mode  os.FileMode
	IsDir bool
	Depth int // Уровень вложенности от исходной директории
}

// EncryptionError отслеживает ошибки при расшифровке
var EncryptionError struct {
	Count int
	Files []string
}

// ResetEncryptionError сбрасывает счетчик ошибок
func ResetEncryptionError() {
	EncryptionError.Count = 0
	EncryptionError.Files = make([]string, 0)
}

// AddEncryptionError добавляет информацию об ошибке расшифровки
func AddEncryptionError(filePath string) {
	EncryptionError.Count++
	EncryptionError.Files = append(EncryptionError.Files, filePath)
}

// PrepareDirectory создает временную директорию __fv__ и копирует содержимое исходной директории
func PrepareDirectory(sourcePath string) (string, error) {
	// Сбрасываем счетчик ошибок шифрования перед началом
	ResetEncryptError()

	// Получаем абсолютный путь для корректной работы
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	// Получаем базовое имя директории
	dirName := filepath.Base(absSourcePath)

	// Создаем имя временной директории
	tempDirName := fmt.Sprintf("__fv__%s", dirName)

	var tempDirPath string
	var isCurrentDir bool

	// Проверяем, является ли источник текущей директорией
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("ошибка при определении текущей директории: %w", err)
	}

	// Если указана текущая директория (абсолютные пути совпадают)
	if absSourcePath == currentDir || sourcePath == "." {
		// Создаем временную директорию внутри текущей директории
		tempDirPath = filepath.Join(absSourcePath, tempDirName)
		isCurrentDir = true
	} else {
		// В остальных случаях создаем временную директорию рядом с исходной
		parentDir := filepath.Dir(absSourcePath)
		tempDirPath = filepath.Join(parentDir, tempDirName)
		isCurrentDir = false
	}

	// Проверяем, существует ли уже временная директория
	if _, err := os.Stat(tempDirPath); err == nil {
		// Если временная директория уже существует, удаляем ее
		err = os.RemoveAll(tempDirPath)
		if err != nil {
			return "", fmt.Errorf("ошибка при удалении существующей временной директории: %w", err)
		}
	}

	// Создаем временную директорию
	err = os.MkdirAll(tempDirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании временной директории: %w", err)
	}

	// Копируем содержимое исходной директории во временную
	if isCurrentDir {
		// Если это текущая директория, копируем файлы по одному
		err = CopyFilesFromCurrentDir(absSourcePath, tempDirPath, tempDirName)
	} else {
		// Для других директорий используем стандартное рекурсивное копирование
		err = CopyDirectory(absSourcePath, tempDirPath)
	}

	if err != nil {
		// В случае ошибки удаляем созданную временную директорию
		os.RemoveAll(tempDirPath)
		return "", fmt.Errorf("ошибка при копировании файлов: %w", err)
	}

	// Шифруем содержимое всех файлов во временной директории
	_ = EncryptDirectoryRecursive(tempDirPath)

	// Проверяем, были ли проблемы с шифрованием файлов
	if EncryptionError.Count > 0 {
		return tempDirPath, fmt.Errorf("при шифровании некоторые файлы были пропущены (%d файлов)", EncryptionError.Count)
	}

	return tempDirPath, nil
}

// CopyFilesFromCurrentDir копирует файлы из текущей директории во временную
// с пропуском самой временной директории
func CopyFilesFromCurrentDir(source, destination, tempDirName string) error {
	// Читаем содержимое исходной директории
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	// Обрабатываем каждый файл/директорию
	for _, entry := range entries {
		entryName := entry.Name()

		// Пропускаем временную директорию, чтобы избежать зацикливания
		if entryName == tempDirName {
			continue
		}

		sourcePath := filepath.Join(source, entryName)
		destPath := filepath.Join(destination, entryName)

		if entry.IsDir() {
			// Копируем директорию рекурсивно
			err = CopyDirectory(sourcePath, destPath)
			if err != nil {
				return err
			}
		} else {
			// Копируем файл
			err = CopyFile(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyDirectory рекурсивно копирует содержимое директории
func CopyDirectory(source, destination string) error {
	// Получаем информацию об исходной директории
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// Создаем директорию назначения с теми же правами доступа
	err = os.MkdirAll(destination, sourceInfo.Mode())
	if err != nil {
		return err
	}

	// Читаем содержимое исходной директории, включая скрытые файлы
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	// Копируем каждый файл или директорию
	for _, entry := range entries {
		// Пропускаем директории __fv__, чтобы избежать зацикливания
		if strings.HasPrefix(entry.Name(), "__fv__") {
			continue
		}

		sourcePath := filepath.Join(source, entry.Name())
		destPath := filepath.Join(destination, entry.Name())

		// Проверяем, не копируем ли мы в саму себя (во избежание бесконечной рекурсии)
		relPath, err := filepath.Rel(source, destPath)
		if err == nil && !strings.HasPrefix(relPath, "..") && relPath != "." {
			// Пропускаем, если путь назначения находится внутри исходного пути
			continue
		}

		if entry.IsDir() {
			// Рекурсивно копируем поддиректории
			err = CopyDirectory(sourcePath, destPath)
			if err != nil {
				return err
			}
		} else {
			// Копируем файлы
			err = CopyFile(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile копирует один файл из исходного пути в путь назначения
func CopyFile(source, destination string) error {
	// Логируем начало операции
	LogDebug("Копирование файла: %s -> %s", source, destination)

	// Открываем исходный файл
	sourceFile, err := os.Open(source)
	if err != nil {
		LogAccess("Чтение", source, err)
		return err
	}
	defer sourceFile.Close()

	// Получаем информацию о файле
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		LogAccess("Stat", source, err)
		return err
	}

	// Создаем файл назначения с теми же правами доступа
	destFile, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		LogAccess("Создание", destination, err)
		return err
	}
	defer destFile.Close()

	// Копируем содержимое
	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		LogAccess("Копирование", source, err)
		return err
	}

	LogDebug("Скопировано байт: %d", bytesWritten)
	LogAccess("Копирование", source, nil)
	return nil
}

// List возвращает рекурсивный список всех файлов в указанной директории
func List(path string) ([]FileInfo, error) {
	// Получаем абсолютный путь для предотвращения проблем с относительными путями
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	files := []FileInfo{}

	// Начинаем с глубины 0
	err = recursiveList(absPath, 0, &files)
	if err != nil {
		return nil, err
	}

	// Сортируем файлы: сначала по глубине, затем директории, затем файлы
	sort.Slice(files, func(i, j int) bool {
		// Сначала сортируем по глубине
		if files[i].Depth != files[j].Depth {
			return files[i].Depth < files[j].Depth
		}

		// Если одинаковая глубина, то директории идут перед файлами
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}

		// В противном случае сортируем по имени
		return filepath.Base(files[i].Path) < filepath.Base(files[j].Path)
	})

	return files, nil
}

// recursiveList рекурсивно обходит директории и добавляет все файлы в список
func recursiveList(path string, depth int, files *[]FileInfo) error {
	// Читаем содержимое директории, включая скрытые файлы
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("ошибка чтения директории '%s': %w", path, err)
	}

	// Обходим все файлы и директории
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			// Пропускаем файлы, к которым нет доступа
			continue
		}

		// Создаем информацию о текущем файле/директории
		fileInfo := FileInfo{
			Path:  fullPath,
			Size:  info.Size(),
			Mode:  info.Mode(),
			IsDir: entry.IsDir(),
			Depth: depth,
		}

		// Добавляем текущий файл/директорию в список
		*files = append(*files, fileInfo)

		// Если это директория, рекурсивно обходим ее
		if entry.IsDir() {
			err = recursiveList(fullPath, depth+1, files)
			if err != nil {
				// Пропускаем недоступные директории, но продолжаем работу
				continue
			}
		}
	}

	return nil
}

// Create создает новый файл с указанным содержимым
func Create(path string, content []byte) error {
	LogDebug("Создание файла: %s (размер: %d байт)", path, len(content))
	err := os.WriteFile(path, content, 0644)
	LogAccess("Запись", path, err)
	return err
}

// Read читает содержимое файла
func Read(path string) ([]byte, error) {
	LogDebug("Чтение файла: %s", path)
	data, err := os.ReadFile(path)
	LogAccess("Чтение", path, err)
	return data, err
}

// PrepareDecryptedDirectory создает директорию для расшифрованных файлов и копирует содержимое из зашифрованной директории
func PrepareDecryptedDirectory(sourcePath string) (string, error) {
	// Получаем абсолютный путь для корректной работы
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	// Получаем базовое имя директории
	dirName := filepath.Base(absSourcePath)

	// Создаем имя расшифрованной директории
	var decryptedDirName string
	var parentDir string

	// Если указана директория с префиксом __fv__, удаляем префикс
	if strings.HasPrefix(dirName, "__fv__") {
		// Удаляем префикс __fv__
		decryptedDirName = dirName[6:]
		parentDir = filepath.Dir(absSourcePath)
	} else {
		// Если директория без префикса, используем оригинальное имя
		decryptedDirName = dirName
		parentDir = filepath.Dir(absSourcePath)
	}

	// Формируем полный путь к директории для расшифрованных файлов
	decryptedDirPath := filepath.Join(parentDir, decryptedDirName)

	// Проверяем, существует ли уже директория для расшифрованных файлов
	if _, err := os.Stat(decryptedDirPath); err == nil {
		// Если директория уже существует, удаляем ее
		err = os.RemoveAll(decryptedDirPath)
		if err != nil {
			return "", fmt.Errorf("ошибка при удалении существующей директории для расшифрованных файлов: %w", err)
		}
	}

	// Создаем директорию для расшифрованных файлов
	err = os.MkdirAll(decryptedDirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании директории для расшифрованных файлов: %w", err)
	}

	// Копируем содержимое из зашифрованной директории
	err = CopyDirectory(absSourcePath, decryptedDirPath)
	if err != nil {
		// В случае ошибки удаляем созданную директорию
		os.RemoveAll(decryptedDirPath)
		return "", fmt.Errorf("ошибка при копировании файлов: %w", err)
	}

	// Расшифровываем содержимое всех файлов
	err = DecryptDirectoryRecursive(decryptedDirPath)
	if err != nil {
		// В случае ошибки удаляем созданную директорию
		os.RemoveAll(decryptedDirPath)
		return "", fmt.Errorf("ошибка при расшифровке файлов: %w", err)
	}

	return decryptedDirPath, nil
}

// DecryptFile расшифровывает содержимое файла и перезаписывает его
func DecryptFile(filePath string) error {
	// Читаем содержимое файла
	data, err := Read(filePath)
	if err != nil {
		return err
	}

	// Если файл пустой, пропускаем его
	if len(data) == 0 {
		return nil
	}

	// Проверяем, зашифрован ли файл
	if !IsFileEncrypted(data) {
		// Если файл не имеет признаков шифрования, пропускаем его
		return nil
	}

	// Расшифровываем данные
	decryptedData, err := DecryptData(data)
	if err != nil {
		// Возвращаем ошибку с неверным паролем, чтобы прервать расшифровку всех файлов
		return fmt.Errorf("ошибка расшифровки файла %s: %w", filePath, err)
	}

	// Перезаписываем файл расшифрованными данными
	err = Create(filePath, decryptedData)
	if err != nil {
		return err
	}

	return nil
}

// min возвращает минимальное из двух значений
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DecryptDirectory расшифровывает все файлы в директории (без рекурсии)
func DecryptDirectory(dirPath string) error {
	// Получаем список файлов в директории
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Расшифровываем каждый файл
	for _, entry := range entries {
		// Пропускаем директории
		if entry.IsDir() {
			continue
		}

		// Формируем полный путь к файлу
		filePath := filepath.Join(dirPath, entry.Name())

		// Расшифровываем файл
		err = DecryptFile(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// DecryptDirectoryRecursive расшифровывает все файлы в директории с рекурсивным обходом поддиректорий
func DecryptDirectoryRecursive(dirPath string) error {
	// Расшифровываем файлы в текущей директории
	err := DecryptDirectory(dirPath)
	if err != nil {
		return err
	}

	// Получаем список поддиректорий
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Рекурсивно вызываем функцию для каждой поддиректории
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			err = DecryptDirectoryRecursive(subDirPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
