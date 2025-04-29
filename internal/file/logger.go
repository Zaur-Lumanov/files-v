package file

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogFile хранит информацию о файле логов
var LogFile struct {
	Enabled bool
	Path    string
	File    *os.File
}

// InitLogger инициализирует логгер и создает лог-файл рядом с исполняемым файлом
func InitLogger() error {
	// Включаем логирование только для Windows
	if runtime.GOOS != "windows" {
		LogFile.Enabled = false
		return nil
	}

	LogFile.Enabled = true

	// Получаем путь к исполняемому файлу
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ошибка получения пути к исполняемому файлу: %w", err)
	}

	// Создаем имя лог-файла с текущей датой и временем
	now := time.Now().Format("2006-01-02_15-04-05")
	logName := fmt.Sprintf("files-v_log_%s.txt", now)
	LogFile.Path = filepath.Join(filepath.Dir(exePath), logName)

	// Создаем лог-файл
	file, err := os.OpenFile(LogFile.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("ошибка создания лог-файла: %w", err)
	}

	LogFile.File = file

	// Записываем информацию о системе
	LogInfo("=== Лог начат %s ===", time.Now().Format("2006-01-02 15:04:05"))
	LogInfo("ОС: %s, Архитектура: %s", runtime.GOOS, runtime.GOARCH)
	LogInfo("Путь к исполняемому файлу: %s", exePath)
	LogInfo("Версия Go: %s", runtime.Version())
	LogInfo("Количество процессоров: %d", runtime.NumCPU())
	LogInfo("======================================")

	return nil
}

// CloseLogger закрывает лог-файл
func CloseLogger() {
	if LogFile.Enabled && LogFile.File != nil {
		LogInfo("=== Лог завершен %s ===", time.Now().Format("2006-01-02 15:04:05"))
		LogFile.File.Close()
	}
}

// LogInfo записывает информационное сообщение в лог
func LogInfo(format string, args ...interface{}) {
	if !LogFile.Enabled || LogFile.File == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[INFO] %s: %s\n", timestamp, message)

	LogFile.File.WriteString(logLine)
}

// LogError записывает сообщение об ошибке в лог
func LogError(format string, args ...interface{}) {
	if !LogFile.Enabled || LogFile.File == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[ERROR] %s: %s\n", timestamp, message)

	LogFile.File.WriteString(logLine)
}

// LogDebug записывает отладочное сообщение в лог
func LogDebug(format string, args ...interface{}) {
	if !LogFile.Enabled || LogFile.File == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Получаем информацию о вызывающей функции
	pc, file, line, ok := runtime.Caller(1)
	caller := "неизвестно"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			caller = fn.Name()
		}
		file = filepath.Base(file)
		caller = fmt.Sprintf("%s:%d [%s]", file, line, caller)
	}

	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[DEBUG] %s: (%s) %s\n", timestamp, caller, message)

	LogFile.File.WriteString(logLine)
}

// LogAccess записывает информацию о доступе к файлу
func LogAccess(operation, path string, err error) {
	if !LogFile.Enabled || LogFile.File == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	status := "успех"
	if err != nil {
		status = fmt.Sprintf("ошибка: %v", err)
	}

	logLine := fmt.Sprintf("[ACCESS] %s: %s [%s] - %s\n", timestamp, operation, path, status)
	LogFile.File.WriteString(logLine)
}
