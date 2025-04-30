package file

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yourusername/files-v/internal/crypto"
)

// Magic signature для зашифрованных файлов
var magicSignature = []byte{0xFE, 0xED, 0xFA, 0xCE, 0xCA, 0xFE, 0xDE, 0xAD}

// encryptionKey - хранит текущий ключ шифрования
var encryptionKey = ""

// currentPrinciple - хранит текущий принцип шифрования
var currentPrinciple *crypto.EncryptionPrinciple

// EncryptError отслеживает ошибки при шифровании
var EncryptError struct {
	Count int
	Files []string
}

// ResetEncryptError сбрасывает счетчик ошибок шифрования
func ResetEncryptError() {
	EncryptError.Count = 0
	EncryptError.Files = make([]string, 0)
}

// AddEncryptError добавляет информацию об ошибке шифрования
func AddEncryptError(filePath string, err error) {
	EncryptError.Count++
	EncryptError.Files = append(EncryptError.Files, fmt.Sprintf("%s: %v", filePath, err))
}

// SetEncryptionKey устанавливает ключ для шифрования/расшифровки
func SetEncryptionKey(key string) {
	encryptionKey = key
}

// SetEncryptionPrinciple устанавливает принцип шифрования
func SetEncryptionPrinciple(principle *crypto.EncryptionPrinciple) {
	currentPrinciple = principle
}

// GetPaddedKey возвращает ключ расширенный до 32 байт для AES-256
func GetPaddedKey() []byte {
	key := []byte(encryptionKey)
	paddedKey := make([]byte, 32)
	copy(paddedKey, key)
	return paddedKey
}

// EncryptData шифрует содержимое файла с использованием указанного принципа
func EncryptData(data []byte) ([]byte, error) {
	if currentPrinciple == nil {
		currentPrinciple = crypto.NewDefaultPrinciple()
	}

	// Создаем заголовок файла
	header := make([]byte, 64)
	copy(header[:8], magicSignature)

	// Записываем принцип шифрования
	principleBytes := currentPrinciple.Bytes()
	if len(principleBytes) > 4 {
		return nil, fmt.Errorf("размер принципа шифрования слишком большой: %d > 4", len(principleBytes))
	}
	copy(header[8:12], principleBytes)

	// Генерируем хеш ключа
	keyHash := sha256.Sum256([]byte(encryptionKey))
	copy(header[12:44], keyHash[:])

	// Заполняем размер данных
	binary.LittleEndian.PutUint32(header[44:48], uint32(len(data)))

	// Генерируем IV
	if _, err := io.ReadFull(rand.Reader, header[48:64]); err != nil {
		return nil, err
	}

	// Шифруем данные
	encryptedData := data
	factory := &crypto.DefaultCipherFactory{}

	for _, algo := range currentPrinciple.Algorithms {
		cipher, err := factory.Create(algo)
		if err != nil {
			return nil, err
		}

		for i := uint8(0); i < currentPrinciple.Passes; i++ {
			encryptedData, err = cipher.Encrypt(encryptedData, GetPaddedKey(), header[48:64])
			if err != nil {
				return nil, err
			}
		}
	}

	// Объединяем заголовок и зашифрованные данные
	result := make([]byte, 0, 64+len(encryptedData))
	result = append(result, header...)
	result = append(result, encryptedData...)

	return result, nil
}

// DecryptData дешифрует содержимое файла
func DecryptData(data []byte) ([]byte, error) {
	LogDebug("Расшифровка данных размером %d байт", len(data))

	// Проверяем минимальный размер
	if len(data) < 64 {
		err := errors.New("зашифрованные данные слишком короткие")
		LogError("Ошибка расшифровки: %v", err)
		return nil, err
	}

	// Проверяем магическую подпись
	for i := 0; i < 8; i++ {
		if data[i] != magicSignature[i] {
			err := errors.New("файл не имеет корректной подписи шифрования")
			LogError("Ошибка расшифровки: %v", err)
			return nil, err
		}
	}

	// Извлекаем принцип шифрования
	principleBytes := data[8:12]
	principle, err := crypto.ParsePrincipleBytes(principleBytes)
	if err != nil {
		LogError("Ошибка парсинга принципа шифрования: %v", err)
		return nil, err
	}

	// Проверяем хеш ключа
	storedKeyHash := data[12:44]
	currentKeyHash := sha256.Sum256([]byte(encryptionKey))

	// Если хеш ключа не совпадает, значит ключ неверный
	for i := 0; i < 32; i++ {
		if storedKeyHash[i] != currentKeyHash[i] {
			err := errors.New("неверный ключ расшифровки")
			LogError("Ошибка расшифровки: %v", err)
			return nil, err
		}
	}

	// Получаем оригинальный размер данных
	originalSize := binary.LittleEndian.Uint32(data[44:48])
	LogDebug("Ожидаемый оригинальный размер данных: %d байт", originalSize)

	// Извлекаем IV и зашифрованные данные
	iv := data[48:64]
	ciphertext := data[64:]

	// Расшифровываем данные
	factory := &crypto.DefaultCipherFactory{}
	decryptedData := ciphertext

	for _, algo := range principle.Algorithms {
		cipher, err := factory.Create(algo)
		if err != nil {
			return nil, err
		}

		for i := uint8(0); i < principle.Passes; i++ {
			decryptedData, err = cipher.Decrypt(decryptedData, GetPaddedKey(), iv)
			if err != nil {
				return nil, err
			}
		}
	}

	// Обрезаем до оригинального размера
	if uint32(len(decryptedData)) >= originalSize {
		decryptedData = decryptedData[:originalSize]
	}

	LogDebug("Данные успешно расшифрованы, размер: %d байт", len(decryptedData))
	return decryptedData, nil
}

// EncryptFile шифрует содержимое файла и перезаписывает его
func EncryptFile(filePath string) error {
	LogDebug("Шифрование файла: %s", filePath)

	// Читаем содержимое файла
	data, err := Read(filePath)
	if err != nil {
		// Логируем ошибку и возвращаем nil, чтобы продолжить шифрование других файлов
		LogError("Ошибка чтения файла для шифрования: %v", err)
		AddEncryptError(filePath, err)
		return nil
	}

	// Если файл пустой, пропускаем его
	if len(data) == 0 {
		LogDebug("Файл пуст, пропускаем: %s", filePath)
		return nil
	}

	// Шифруем данные
	encryptedData, err := EncryptData(data)
	if err != nil {
		// Логируем ошибку и возвращаем nil, чтобы продолжить шифрование других файлов
		LogError("Ошибка шифрования данных: %v", err)
		AddEncryptError(filePath, err)
		return nil
	}

	LogDebug("Данные успешно зашифрованы: %s (исходный размер: %d, зашифрованный размер: %d)",
		filePath, len(data), len(encryptedData))

	// Перезаписываем файл зашифрованными данными
	err = Create(filePath, encryptedData)
	if err != nil {
		// Логируем ошибку и возвращаем nil, чтобы продолжить шифрование других файлов
		LogError("Ошибка записи зашифрованных данных: %v", err)
		AddEncryptError(filePath, err)
		return nil
	}

	LogDebug("Файл успешно зашифрован: %s", filePath)
	return nil
}

// IsFileEncrypted проверяет, является ли файл зашифрованным
func IsFileEncrypted(data []byte) bool {
	if len(data) < len(magicSignature) {
		return false
	}

	// Проверяем магическую подпись
	for i := 0; i < len(magicSignature); i++ {
		if data[i] != magicSignature[i] {
			return false
		}
	}

	return true
}

// EncryptDirectory шифрует все файлы в директории (без рекурсии)
func EncryptDirectory(dirPath string) error {
	// Получаем список файлов в директории
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Логируем ошибку и возвращаем nil, чтобы продолжить шифрование других директорий
		AddEncryptError(dirPath, err)
		return nil
	}

	// Шифруем каждый файл
	for _, entry := range entries {
		// Пропускаем директории
		if entry.IsDir() {
			continue
		}

		// Формируем полный путь к файлу
		filePath := filepath.Join(dirPath, entry.Name())

		// Шифруем файл - EncryptFile уже обрабатывает ошибки
		_ = EncryptFile(filePath)
	}

	return nil
}

// EncryptDirectoryRecursive шифрует все файлы в директории с рекурсивным обходом поддиректорий
func EncryptDirectoryRecursive(dirPath string) error {
	// Шифруем файлы в текущей директории - EncryptDirectory уже обрабатывает ошибки
	_ = EncryptDirectory(dirPath)

	// Получаем список поддиректорий
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Логируем ошибку и возвращаем nil, чтобы продолжить шифрование других директорий
		AddEncryptError(dirPath, err)
		return nil
	}

	// Рекурсивно вызываем функцию для каждой поддиректории
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			// Вызываем функцию для поддиректории - игнорируем ошибки, так как они уже обработаны
			_ = EncryptDirectoryRecursive(subDirPath)
		}
	}

	return nil
}
