#!/bin/bash

# Создаем директорию для бинарников, если она еще не существует
mkdir -p bin

# Текущая версия программы
VERSION="1.0.0"

# Список операционных систем и архитектур для сборки
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Цвета для вывода в терминал
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Начинаем сборку Files-V версии ${VERSION}...${NC}"

# Проходим по всем платформам и собираем бинарники
for PLATFORM in "${PLATFORMS[@]}"; do
    # Разделяем строку платформы на ОС и архитектуру
    OS=$(echo $PLATFORM | cut -d/ -f1)
    ARCH=$(echo $PLATFORM | cut -d/ -f2)
    
    # Формируем имя бинарника с учетом ОС и архитектуры
    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME=bin/files-v_${VERSION}_${OS}_${ARCH}.exe
    else
        OUTPUT_NAME=bin/files-v_${VERSION}_${OS}_${ARCH}
    fi
    
    echo -e "Сборка для ${GREEN}$OS/$ARCH${NC}..."
    
    # Устанавливаем переменные окружения для кросс-компиляции
    GOOS=$OS GOARCH=$ARCH go build -o $OUTPUT_NAME ./cmd/files-v
    
    # Проверяем успешность сборки
    if [ $? -eq 0 ]; then
        echo -e "  ✓ Бинарник создан: ${GREEN}$OUTPUT_NAME${NC}"
    else
        echo -e "  ✗ Ошибка при сборке для $OS/$ARCH"
    fi
done

echo -e "${YELLOW}Сборка завершена. Бинарники находятся в директории bin/${NC}"

# Выводим список созданных файлов
echo -e "\nСозданные бинарники:"
ls -lh bin/ 