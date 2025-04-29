package file

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Magic signature для зашифрованных файлов
var magicSignature = []byte{0xFE, 0xED, 0xFA, 0xCE, 0xCA, 0xFE, 0xDE, 0xAD}

// encryptionKey - хранит текущий ключ шифрования
var encryptionKey = ""

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

// GetPaddedKey возвращает ключ расширенный до 32 байт для AES-256
func GetPaddedKey() []byte {
	key := []byte(encryptionKey)
	paddedKey := make([]byte, 32)
	copy(paddedKey, key)
	return paddedKey
}

// EncryptData шифрует содержимое файла с использованием AES-256
func EncryptData(data []byte) ([]byte, error) {
	// Получаем расширенный ключ для AES-256
	paddedKey := GetPaddedKey()

	// Создаем шифр
	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return nil, err
	}

	// Добавляем магическую подпись и хеш ключа для проверки правильности расшифровки
	signatureSize := len(magicSignature)
	keyHashSize := sha256.Size

	// Создаем вектор инициализации
	ciphertext := make([]byte, signatureSize+keyHashSize+4+aes.BlockSize+len(data))

	// Добавляем магическую подпись
	copy(ciphertext[:signatureSize], magicSignature)

	// Добавляем хеш ключа (чтобы потом проверить правильность расшифровки)
	keyHash := sha256.Sum256([]byte(encryptionKey))
	copy(ciphertext[signatureSize:signatureSize+keyHashSize], keyHash[:])

	// Добавляем размер оригинальных данных
	binary.LittleEndian.PutUint32(ciphertext[signatureSize+keyHashSize:signatureSize+keyHashSize+4], uint32(len(data)))

	// Подготавливаем IV
	iv := ciphertext[signatureSize+keyHashSize+4 : signatureSize+keyHashSize+4+aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Шифруем данные
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[signatureSize+keyHashSize+4+aes.BlockSize:], data)

	return ciphertext, nil
}

// DecryptData дешифрует содержимое файла
func DecryptData(data []byte) ([]byte, error) {
	LogDebug("Расшифровка данных размером %d байт", len(data))

	signatureSize := len(magicSignature)
	keyHashSize := sha256.Size
	headerSize := signatureSize + keyHashSize + 4

	// Проверяем длину данных
	if len(data) < headerSize+aes.BlockSize {
		err := errors.New("зашифрованные данные слишком короткие")
		LogError("Ошибка расшифровки: %v", err)
		return nil, err
	}

	// Проверяем магическую подпись
	for i := 0; i < signatureSize; i++ {
		if data[i] != magicSignature[i] {
			err := errors.New("файл не имеет корректной подписи шифрования")
			LogError("Ошибка расшифровки: %v", err)
			return nil, err
		}
	}

	// Проверяем хеш ключа
	storedKeyHash := data[signatureSize : signatureSize+keyHashSize]
	currentKeyHash := sha256.Sum256([]byte(encryptionKey))

	// Если хеш ключа не совпадает, значит ключ неверный
	for i := 0; i < keyHashSize; i++ {
		if storedKeyHash[i] != currentKeyHash[i] {
			err := errors.New("неверный ключ расшифровки")
			LogError("Ошибка расшифровки: %v", err)
			return nil, err
		}
	}

	// Получаем оригинальный размер данных
	originalSize := binary.LittleEndian.Uint32(data[signatureSize+keyHashSize : signatureSize+keyHashSize+4])
	LogDebug("Ожидаемый оригинальный размер данных: %d байт", originalSize)

	// Получаем расширенный ключ для AES-256
	paddedKey := GetPaddedKey()

	// Создаем шифр
	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		LogError("Ошибка создания шифра AES: %v", err)
		return nil, err
	}

	// Извлекаем вектор инициализации
	iv := data[headerSize : headerSize+aes.BlockSize]
	ciphertext := data[headerSize+aes.BlockSize:]

	// Дешифруем данные
	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	// Обрезаем до оригинального размера
	if uint32(len(plaintext)) >= originalSize {
		plaintext = plaintext[:originalSize]
	}

	LogDebug("Данные успешно расшифрованы, размер: %d байт", len(plaintext))
	return plaintext, nil
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
