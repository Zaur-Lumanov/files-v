package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Конфигурация сборки
const (
	version     = "1.0.0"
	mainPackage = "./cmd/files-v"
	outputDir   = "./bin"
)

// Платформы для сборки
var platforms = []struct {
	os   string
	arch string
}{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
	{"windows", "arm64"},
}

func main() {
	// Создаем директорию для бинарников, если она еще не существует
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Ошибка при создании директории %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	fmt.Printf("Начинаем сборку Files-V версии %s...\n", version)

	// Проходим по всем платформам и собираем бинарники
	for _, platform := range platforms {
		// Формируем имя бинарника с учетом ОС и архитектуры
		outputName := fmt.Sprintf("files-v_%s_%s_%s", version, platform.os, platform.arch)
		if platform.os == "windows" {
			outputName += ".exe"
		}
		outputPath := filepath.Join(outputDir, outputName)

		fmt.Printf("Сборка для %s/%s...\n", platform.os, platform.arch)

		// Подготавливаем команду для сборки
		cmd := exec.Command("go", "build", "-o", outputPath, mainPackage)
		cmd.Env = append(os.Environ(),
			"GOOS="+platform.os,
			"GOARCH="+platform.arch,
		)

		// Запускаем сборку
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  ✗ Ошибка при сборке для %s/%s: %v\n", platform.os, platform.arch, err)
			fmt.Printf("    Вывод: %s\n", output)
			continue
		}

		fmt.Printf("  ✓ Бинарник создан: %s\n", outputPath)
	}

	fmt.Println("Сборка завершена. Бинарники находятся в директории", outputDir)

	// Выводим список созданных файлов
	fmt.Println("\nСозданные бинарники:")
	files, err := os.ReadDir(outputDir)
	if err != nil {
		fmt.Printf("Ошибка при чтении директории %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "files-v_") {
			info, err := file.Info()
			if err != nil {
				fmt.Printf("  %s\n", file.Name())
			} else {
				size := formatSize(info.Size())
				fmt.Printf("  %s (%s)\n", file.Name(), size)
			}
		}
	}
}

// formatSize форматирует размер файла в человекочитаемый вид
func formatSize(size int64) string {
	const (
		B  int64 = 1
		KB       = B * 1024
		MB       = KB * 1024
		GB       = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
