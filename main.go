package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Получаем путь к текущему исполняемому файлу
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("Ошибка получения пути к исполняемому файлу: %v\n", err)
		os.Exit(1)
	}

	// Получаем директорию проекта
	projectDir := filepath.Dir(exe)

	// Формируем путь к cmd/files-v
	cmdPath := filepath.Join(projectDir, "cmd", "files-v")

	// Создаем команду для запуска cmd/files-v
	cmd := exec.Command("go", "run", cmdPath)
	cmd.Args = append(cmd.Args, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Запускаем команду
	if err := cmd.Run(); err != nil {
		fmt.Printf("Ошибка запуска cmd/files-v: %v\n", err)
		os.Exit(1)
	}
}
