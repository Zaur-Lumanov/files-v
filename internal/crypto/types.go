package crypto

import (
	"fmt"
	"strings"
)

// AlgorithmType представляет тип алгоритма шифрования
type AlgorithmType uint8

const (
	AlgorithmAES128 AlgorithmType = iota
	AlgorithmAES192
	AlgorithmAES256
	AlgorithmChaCha20
)

// Mode представляет режим работы блочного шифра
type Mode uint8

const (
	ModeECB Mode = iota // Electronic Codebook
	ModeCBC             // Cipher Block Chaining
	ModeCFB             // Cipher Feedback
	ModeOFB             // Output Feedback
	ModeCTR             // Counter
	ModeGCM             // Galois/Counter Mode
	ModeCCM             // Counter with CBC-MAC
)

// Algorithm представляет алгоритм шифрования
type Algorithm struct {
	Type AlgorithmType // Тип алгоритма
	Mode Mode          // Режим работы
}

// EncryptionPrinciple описывает принцип шифрования
type EncryptionPrinciple struct {
	Algorithms []Algorithm // Последовательность алгоритмов
	Passes     uint8       // Количество проходов
}

// NewDefaultPrinciple создает принцип шифрования по умолчанию
func NewDefaultPrinciple() *EncryptionPrinciple {
	return &EncryptionPrinciple{
		Algorithms: []Algorithm{
			{
				Type: AlgorithmAES256,
				Mode: ModeCFB,
			},
		},
		Passes: 1,
	}
}

// Bytes возвращает байтовое представление принципа шифрования
func (p *EncryptionPrinciple) Bytes() []byte {
	// Формат: [количество алгоритмов(1)] [алгоритм1(2)] [алгоритм2(2)] ... [количество проходов(1)]
	result := make([]byte, 1+len(p.Algorithms)*2+1)
	result[0] = byte(len(p.Algorithms))
	for i, alg := range p.Algorithms {
		result[1+i*2] = byte(alg.Type)
		result[1+i*2+1] = byte(alg.Mode)
	}
	result[len(result)-1] = p.Passes
	return result
}

// ParsePrinciple парсит строковое представление принципа шифрования
func ParsePrinciple(s string) (*EncryptionPrinciple, error) {
	parts := strings.Split(s, "+")
	if len(parts) == 0 {
		return nil, fmt.Errorf("пустой принцип шифрования")
	}

	// Парсим количество проходов
	passes := uint8(1)
	if strings.HasPrefix(parts[len(parts)-1], "x") {
		if _, err := fmt.Sscanf(parts[len(parts)-1], "x%d", &passes); err != nil {
			return nil, fmt.Errorf("неверный формат количества проходов: %v", err)
		}
		parts = parts[:len(parts)-1]
	}

	// Парсим алгоритмы
	algorithms := make([]Algorithm, 0, len(parts))
	for _, part := range parts {
		algParts := strings.Split(part, "-")
		if len(algParts) != 2 {
			return nil, fmt.Errorf("неверный формат алгоритма: %s", part)
		}

		var algType AlgorithmType
		switch strings.ToUpper(algParts[0]) {
		case "AES128":
			algType = AlgorithmAES128
		case "AES192":
			algType = AlgorithmAES192
		case "AES256":
			algType = AlgorithmAES256
		case "CHACHA20":
			algType = AlgorithmChaCha20
		default:
			return nil, fmt.Errorf("неизвестный тип алгоритма: %s", algParts[0])
		}

		var mode Mode
		switch strings.ToUpper(algParts[1]) {
		case "ECB":
			mode = ModeECB
		case "CBC":
			mode = ModeCBC
		case "CFB":
			mode = ModeCFB
		case "OFB":
			mode = ModeOFB
		case "CTR":
			mode = ModeCTR
		case "GCM":
			mode = ModeGCM
		case "CCM":
			mode = ModeCCM
		default:
			return nil, fmt.Errorf("неизвестный режим работы: %s", algParts[1])
		}

		algorithms = append(algorithms, Algorithm{
			Type: algType,
			Mode: mode,
		})
	}

	return &EncryptionPrinciple{
		Algorithms: algorithms,
		Passes:     passes,
	}, nil
}

// ParsePrincipleBytes парсит байтовое представление принципа шифрования
func ParsePrincipleBytes(data []byte) (*EncryptionPrinciple, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("недостаточно данных для парсинга принципа шифрования")
	}

	// Формат: [количество алгоритмов(1)] [алгоритм1(2)] [алгоритм2(2)] ... [количество проходов(1)]
	numAlgorithms := int(data[0])
	if numAlgorithms == 0 {
		return nil, fmt.Errorf("количество алгоритмов не может быть нулевым")
	}

	algorithms := make([]Algorithm, 0, numAlgorithms)
	for i := 0; i < numAlgorithms; i++ {
		if 1+i*2+1 >= len(data) {
			return nil, fmt.Errorf("недостаточно данных для парсинга алгоритма %d", i)
		}
		algorithms = append(algorithms, Algorithm{
			Type: AlgorithmType(data[1+i*2]),
			Mode: Mode(data[1+i*2+1]),
		})
	}

	passes := uint8(data[len(data)-1])
	if passes == 0 {
		return nil, fmt.Errorf("количество проходов не может быть нулевым")
	}

	return &EncryptionPrinciple{
		Algorithms: algorithms,
		Passes:     passes,
	}, nil
}

// String возвращает строковое представление принципа шифрования
func (p *EncryptionPrinciple) String() string {
	var sb strings.Builder
	for i, alg := range p.Algorithms {
		if i > 0 {
			sb.WriteString("+")
		}
		sb.WriteString(alg.String())
	}
	if p.Passes > 1 {
		fmt.Fprintf(&sb, "+x%d", p.Passes)
	}
	return sb.String()
}

// String возвращает строковое представление алгоритма
func (a Algorithm) String() string {
	var typeStr string
	switch a.Type {
	case AlgorithmAES128:
		typeStr = "AES128"
	case AlgorithmAES192:
		typeStr = "AES192"
	case AlgorithmAES256:
		typeStr = "AES256"
	case AlgorithmChaCha20:
		typeStr = "ChaCha20"
	default:
		typeStr = "Unknown"
	}

	var modeStr string
	switch a.Mode {
	case ModeECB:
		modeStr = "ECB"
	case ModeCBC:
		modeStr = "CBC"
	case ModeCFB:
		modeStr = "CFB"
	case ModeOFB:
		modeStr = "OFB"
	case ModeCTR:
		modeStr = "CTR"
	case ModeGCM:
		modeStr = "GCM"
	case ModeCCM:
		modeStr = "CCM"
	default:
		modeStr = "Unknown"
	}

	return fmt.Sprintf("%s-%s", typeStr, modeStr)
}

// String returns string representation of Mode
func (m Mode) String() string {
	switch m {
	case ModeECB:
		return "ECB"
	case ModeCBC:
		return "CBC"
	case ModeCFB:
		return "CFB"
	case ModeOFB:
		return "OFB"
	case ModeCTR:
		return "CTR"
	case ModeGCM:
		return "GCM"
	case ModeCCM:
		return "CCM"
	default:
		return "Unknown"
	}
}

// Version представляет версию формата файла
type Version struct {
	Major uint8
	Minor uint8
	Patch uint8
}

// CurrentVersion текущая версия формата файла
var CurrentVersion = Version{
	Major: 1,
	Minor: 0,
	Patch: 0,
}

// String возвращает строковое представление версии
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsCompatible проверяет совместимость версий
func (v Version) IsCompatible(other Version) bool {
	// Основная версия должна совпадать
	if v.Major != other.Major {
		return false
	}

	// Младшая версия должна быть не меньше
	if v.Minor < other.Minor {
		return false
	}

	// Если младшие версии совпадают, проверяем патч
	if v.Minor == other.Minor && v.Patch < other.Patch {
		return false
	}

	return true
}
