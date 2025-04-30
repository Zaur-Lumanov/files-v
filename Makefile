.PHONY: all build build-all clean help

# Переменные
VERSION := 1.0.0
BINARY_NAME := files-v
BUILD_DIR := ./bin

# Команды
all: clean build-all

# Сборка для текущей ОС и архитектуры
build:
	@echo "Сборка для текущей платформы..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Готово! Бинарник создан: $(BUILD_DIR)/$(BINARY_NAME)"

# Сборка для всех поддерживаемых платформ
build-all:
	@echo "Сборка для всех платформ..."
	@./build.sh

# Сборка с использованием Go скрипта
build-go:
	@echo "Сборка с использованием Go скрипта..."
	@go run ./scripts/build.go

# Очистка директории сборки
clean:
	@echo "Очистка директории сборки..."
	@rm -rf $(BUILD_DIR)/*
	@echo "Директория очищена!"

# Показать справку
help:
	@echo "Доступные команды:"
	@echo "  make build      - собрать бинарник для текущей платформы"
	@echo "  make build-all  - собрать бинарники для всех платформ (bash скрипт)"
	@echo "  make build-go   - собрать бинарники для всех платформ (Go скрипт)"
	@echo "  make clean      - очистить директорию сборки"
	@echo "  make help       - показать эту справку"
	@echo "  make all        - очистить и пересобрать все бинарники" 