package main

import (
	"fmt"
	"os"

	"github.com/Zaur-Lumanov/files-v/internal/file"
)

func main() {
	// Инициализируем логгер
	err := file.InitLogger()
	if err != nil {
		fmt.Printf("Предупреждение: не удалось инициализировать логгер: %v\n", err)
	}
	defer file.CloseLogger()

	// Если переданы аргументы, используем первый как целевую директорию
	targetDir := ""
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	}

	// Запускаем приложение
	exitCode := Run(targetDir)

	// Добавляем информацию о Windows-специфичных ошибках в лог
	if file.LogFile.Enabled {
		file.LogInfo("Программа завершила работу с кодом: %d", exitCode)
		if file.EncryptError.Count > 0 {
			file.LogError("Во время шифрования возникли проблемы с %d файлами", file.EncryptError.Count)
			for _, errFile := range file.EncryptError.Files {
				file.LogError("Пропущенный файл: %s", errFile)
			}
		}
	}

	os.Exit(exitCode)
}
