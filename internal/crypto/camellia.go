package crypto

import (
	"crypto/cipher"
	"fmt"

	"github.com/emmansun/gmsm/sm4"
)

// CamelliaCipher реализует шифрование Camellia
type CamelliaCipher struct {
	keySize int
	mode    Mode
}

// createCamelliaCipher создает шифр Camellia
func createCamelliaCipher(keySize int, mode Mode) (Cipher, error) {
	return &CamelliaCipher{
		keySize: keySize,
		mode:    mode,
	}, nil
}

// Encrypt шифрует данные с использованием Camellia
func (c *CamelliaCipher) Encrypt(data, key, iv []byte) ([]byte, error) {
	if len(key) != c.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", c.keySize, len(key))
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create camellia block: %v", err)
	}

	return encryptBlockCipher(block, Mode(c.mode), data, iv)
}

// Decrypt расшифровывает данные с использованием Camellia
func (c *CamelliaCipher) Decrypt(data, key, iv []byte) ([]byte, error) {
	if len(key) != c.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", c.keySize, len(key))
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create camellia block: %v", err)
	}

	return decryptBlockCipher(block, Mode(c.mode), data, iv)
}

// BlockSize возвращает размер блока
func (c *CamelliaCipher) BlockSize() int {
	return 16 // Размер блока Camellia
}

// KeySize возвращает размер ключа
func (c *CamelliaCipher) KeySize() int {
	return c.keySize
}

// encryptBlockCipher шифрует данные с использованием блочного шифра
func encryptBlockCipher(block cipher.Block, mode Mode, data []byte, iv []byte) ([]byte, error) {
	switch mode {
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
	default:
		return nil, fmt.Errorf("unsupported mode: %v", mode)
	}
}

// decryptBlockCipher расшифровывает данные с использованием блочного шифра
func decryptBlockCipher(block cipher.Block, mode Mode, data []byte, iv []byte) ([]byte, error) {
	switch mode {
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
	default:
		return nil, fmt.Errorf("unsupported mode: %v", mode)
	}
}
