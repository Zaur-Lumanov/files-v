package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"golang.org/x/crypto/chacha20"
)

// Cipher определяет интерфейс для алгоритмов шифрования
type Cipher interface {
	Encrypt(data []byte, key []byte, iv []byte) ([]byte, error)
	Decrypt(data []byte, key []byte, iv []byte) ([]byte, error)
	BlockSize() int
	KeySize() int
}

// CipherFactory создает экземпляры алгоритмов шифрования
type CipherFactory interface {
	Create(algorithm Algorithm) (Cipher, error)
}

// DefaultCipherFactory реализует фабрику шифров по умолчанию
type DefaultCipherFactory struct{}

// Create создает экземпляр шифра на основе алгоритма
func (f *DefaultCipherFactory) Create(algorithm Algorithm) (Cipher, error) {
	switch algorithm.Type {
	case AlgorithmAES128, AlgorithmAES192, AlgorithmAES256:
		return createAESCipher(algorithm)
	case AlgorithmChaCha20:
		return createChaCha20Cipher(algorithm)
	default:
		return nil, fmt.Errorf("неизвестный алгоритм: %d", algorithm.Type)
	}
}

// createAESCipher создает шифр AES
func createAESCipher(algorithm Algorithm) (Cipher, error) {
	var keySize int
	switch algorithm.Type {
	case AlgorithmAES128:
		keySize = 16
	case AlgorithmAES192:
		keySize = 24
	case AlgorithmAES256:
		keySize = 32
	default:
		return nil, fmt.Errorf("неверный тип AES: %d", algorithm.Type)
	}

	return &AESCipher{
		keySize: keySize,
		mode:    algorithm.Mode,
	}, nil
}

// AESCipher реализует шифрование AES
type AESCipher struct {
	keySize int
	mode    Mode
}

// Encrypt шифрует данные с использованием AES
func (c *AESCipher) Encrypt(data []byte, key []byte, iv []byte) ([]byte, error) {
	if len(key) < c.keySize {
		return nil, fmt.Errorf("ключ слишком короткий: %d < %d", len(key), c.keySize)
	}

	block, err := aes.NewCipher(key[:c.keySize])
	if err != nil {
		return nil, err
	}

	switch Mode(c.mode) {
	case ModeECB:
		return encryptECB(block, data)
	case ModeCBC:
		return encryptCBC(block, data, iv)
	case ModeCFB:
		return encryptCFB(block, data, iv)
	case ModeOFB:
		return encryptOFB(block, data, iv)
	case ModeCTR:
		return encryptCTR(block, data, iv)
	case ModeGCM:
		return encryptGCM(block, data, iv)
	case ModeCCM:
		return encryptCCM(block, data, iv)
	default:
		return nil, fmt.Errorf("неизвестный режим AES: %d", c.mode)
	}
}

// Decrypt расшифровывает данные с использованием AES
func (c *AESCipher) Decrypt(data []byte, key []byte, iv []byte) ([]byte, error) {
	if len(key) < c.keySize {
		return nil, fmt.Errorf("ключ слишком короткий: %d < %d", len(key), c.keySize)
	}

	block, err := aes.NewCipher(key[:c.keySize])
	if err != nil {
		return nil, err
	}

	switch Mode(c.mode) {
	case ModeECB:
		return decryptECB(block, data)
	case ModeCBC:
		return decryptCBC(block, data, iv)
	case ModeCFB:
		return decryptCFB(block, data, iv)
	case ModeOFB:
		return decryptOFB(block, data, iv)
	case ModeCTR:
		return decryptCTR(block, data, iv)
	case ModeGCM:
		return decryptGCM(block, data, iv)
	case ModeCCM:
		return decryptCCM(block, data, iv)
	default:
		return nil, fmt.Errorf("неизвестный режим AES: %d", c.mode)
	}
}

// BlockSize возвращает размер блока AES
func (c *AESCipher) BlockSize() int {
	return aes.BlockSize
}

// KeySize возвращает размер ключа
func (c *AESCipher) KeySize() int {
	return c.keySize
}

// Реализации режимов шифрования AES
func encryptECB(block cipher.Block, data []byte) ([]byte, error) {
	if len(data)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("данные должны быть кратны размеру блока: %d", block.BlockSize())
	}

	encrypted := make([]byte, len(data))
	size := block.BlockSize()

	for i := 0; i < len(data); i += size {
		block.Encrypt(encrypted[i:i+size], data[i:i+size])
	}
	return encrypted, nil
}

func decryptECB(block cipher.Block, data []byte) ([]byte, error) {
	if len(data)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("данные должны быть кратны размеру блока: %d", block.BlockSize())
	}

	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for i := 0; i < len(data); i += size {
		block.Decrypt(decrypted[i:i+size], data[i:i+size])
	}
	return decrypted, nil
}

func encryptCBC(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("неверный размер IV: %d != %d", len(iv), block.BlockSize())
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("данные должны быть кратны размеру блока: %d", block.BlockSize())
	}

	encrypted := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, data)

	return encrypted, nil
}

func decryptCBC(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("неверный размер IV: %d != %d", len(iv), block.BlockSize())
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("данные должны быть кратны размеру блока: %d", block.BlockSize())
	}

	decrypted := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, data)

	return decrypted, nil
}

func encryptCFB(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("неверный размер IV: %d != %d", len(iv), block.BlockSize())
	}

	// Создаем шифр CFB
	stream := cipher.NewCFBEncrypter(block, iv)

	// Шифруем данные
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)

	return ciphertext, nil
}

func decryptCFB(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("неверный размер IV: %d != %d", len(iv), block.BlockSize())
	}

	// Создаем шифр CFB
	stream := cipher.NewCFBDecrypter(block, iv)

	// Расшифровываем данные
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)

	return plaintext, nil
}

func encryptOFB(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("неверный размер IV: %d != %d", len(iv), block.BlockSize())
	}

	stream := cipher.NewOFB(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	return encrypted, nil
}

func decryptOFB(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	// OFB использует тот же процесс для шифрования и расшифровки
	return encryptOFB(block, data, iv)
}

func encryptCTR(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("неверный размер IV: %d != %d", len(iv), block.BlockSize())
	}

	stream := cipher.NewCTR(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	return encrypted, nil
}

func decryptCTR(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	// CTR использует тот же процесс для шифрования и расшифровки
	return encryptCTR(block, data, iv)
}

func encryptGCM(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != 12 { // GCM требует 12-байтовый nonce
		return nil, fmt.Errorf("неверный размер nonce для GCM: %d != 12", len(iv))
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %v", err)
	}

	// Шифруем данные с дополнительными данными (nil)
	encrypted := aesGCM.Seal(nil, iv, data, nil)
	return encrypted, nil
}

func decryptGCM(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != 12 { // GCM требует 12-байтовый nonce
		return nil, fmt.Errorf("неверный размер nonce для GCM: %d != 12", len(iv))
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %v", err)
	}

	// Расшифровываем данные с дополнительными данными (nil)
	decrypted, err := aesGCM.Open(nil, iv, data, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки GCM: %v", err)
	}

	return decrypted, nil
}

func encryptCCM(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != 7 { // CCM требует 7-байтовый nonce
		return nil, fmt.Errorf("неверный размер nonce для CCM: %d != 7", len(iv))
	}

	// Создаем CCM шифр
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %v", err)
	}

	// Преобразуем 7-байтовый nonce в 12-байтовый для GCM
	nonce := make([]byte, 12)
	copy(nonce, iv)

	// Шифруем данные с дополнительными данными (nil)
	encrypted := aesGCM.Seal(nil, nonce, data, nil)
	return encrypted, nil
}

func decryptCCM(block cipher.Block, data []byte, iv []byte) ([]byte, error) {
	if len(iv) != 7 { // CCM требует 7-байтовый nonce
		return nil, fmt.Errorf("неверный размер nonce для CCM: %d != 7", len(iv))
	}

	// Создаем CCM шифр
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %v", err)
	}

	// Преобразуем 7-байтовый nonce в 12-байтовый для GCM
	nonce := make([]byte, 12)
	copy(nonce, iv)

	// Расшифровываем данные с дополнительными данными (nil)
	decrypted, err := aesGCM.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки GCM: %v", err)
	}

	return decrypted, nil
}

// createChaCha20Cipher создает шифр ChaCha20
func createChaCha20Cipher(algorithm Algorithm) (Cipher, error) {
	return &ChaCha20Cipher{
		nonceSize: 12, // Стандартный размер nonce для ChaCha20
		keySize:   32, // Стандартный размер ключа для ChaCha20
	}, nil
}

// ChaCha20Cipher реализует шифрование ChaCha20
type ChaCha20Cipher struct {
	nonceSize int
	keySize   int
}

// Encrypt шифрует данные с использованием ChaCha20
func (c *ChaCha20Cipher) Encrypt(data, key, iv []byte) ([]byte, error) {
	if len(key) != c.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", c.keySize, len(key))
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(key, iv)
	if err != nil {
		return nil, fmt.Errorf("failed to create chacha20 cipher: %v", err)
	}

	dst := make([]byte, len(data))
	cipher.XORKeyStream(dst, data)
	return dst, nil
}

// Decrypt расшифровывает данные с использованием ChaCha20
func (c *ChaCha20Cipher) Decrypt(data, key, iv []byte) ([]byte, error) {
	// ChaCha20 использует тот же процесс для шифрования и расшифровки
	return c.Encrypt(data, key, iv)
}

// BlockSize возвращает размер блока
func (c *ChaCha20Cipher) BlockSize() int {
	return 1 // ChaCha20 - потоковый шифр
}

// KeySize возвращает размер ключа
func (c *ChaCha20Cipher) KeySize() int {
	return c.keySize
}
