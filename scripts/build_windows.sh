#!/bin/bash

# Устанавливаем переменные среды для кросс-компиляции
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=0

# Создаем директорию для билдов, если она не существует
mkdir -p bin/windows

# Выводим информацию о сборке
echo "Сборка Windows-версии (amd64) с поддержкой логирования..."

# Компилируем приложение
go build -o bin/windows/files-v.exe -ldflags "-s -w" ./cmd/files-v

# Проверяем результат сборки
if [ $? -eq 0 ]; then
    echo "Сборка x64 успешно завершена: bin/windows/files-v.exe"
else
    echo "Ошибка при сборке x64 версии"
    exit 1
fi

# Собираем 32-битную версию для более старых систем
export GOARCH=386

echo "Сборка Windows-версии (x86) с поддержкой логирования..."

# Компилируем 32-битную версию
go build -o bin/windows/files-v_x86.exe -ldflags "-s -w" ./cmd/files-v

# Проверяем результат сборки
if [ $? -eq 0 ]; then
    echo "Сборка x86 успешно завершена: bin/windows/files-v_x86.exe"
else
    echo "Ошибка при сборке x86 версии"
    exit 1
fi

# Создаем README.txt с инструкциями
cat > bin/windows/README.txt << EOF
Files-V - утилита для шифрования и расшифровки файлов

ТЕСТОВАЯ ВЕРСИЯ С ЛОГИРОВАНИЕМ ДЛЯ WINDOWS

Эта версия автоматически создает файл лога рядом с исполняемым файлом.
Имя файла лога: files-v_log_YYYY-MM-DD_HH-MM-SS.txt

При возникновении проблем, пожалуйста, отправьте этот файл разработчику.

Использование:
  files-v.exe <директория_или_файл> [опции]

Опции:
  -help       Показать справку
  -dir        Директория для работы
  -d, -decrypt Режим расшифровки
  -p, -password Пароль для шифрования/расшифровки
  -z, -zip    Использовать ZIP-архивацию

Примеры:
  Шифрование директории (запросит пароль):
    files-v.exe C:\Documents\MyFiles
  
  Шифрование с указанием пароля:
    files-v.exe C:\Documents\MyFiles -p my_password
  
  Расшифровка директории:
    files-v.exe C:\Documents\__fv__MyFiles
EOF

echo "Создан файл README.txt с инструкциями"

# Упаковываем все в ZIP-архив
cd bin
zip -r windows_debug_build.zip windows/
cd ..

echo "Готово! Архив с отладочной сборкой создан: bin/windows_debug_build.zip"
echo "Отправьте этот архив пользователю для тестирования на Windows." 