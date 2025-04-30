package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Zaur-Lumanov/files-v/internal/file"
	"github.com/Zaur-Lumanov/files-v/pkg/utils"
	"golang.org/x/term"
)

const (
	version = "1.0.0"
)

// readPassword запрашивает ввод пароля без отображения символов на экране
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Переключаем терминал в режим без отображения ввода
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Добавляем перевод строки после ввода пароля

	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}

// Run запускает приложение и возвращает код выхода
func Run(targetDir string) int {
	var (
		showHelp    bool
		showVersion bool
		directory   string
		decrypt     bool
		key         string
		useZip      bool
	)

	// Создаем новый FlagSet для корректной обработки флагов
	flagSet := flag.NewFlagSet("files-v", flag.ExitOnError)

	// Парсинг флагов
	flagSet.BoolVar(&showHelp, "help", false, "Показать справку")
	flagSet.BoolVar(&showVersion, "v", false, "Показать версию")
	flagSet.BoolVar(&showVersion, "version", false, "Показать версию")
	flagSet.StringVar(&directory, "dir", targetDir, "Директория для работы (перезаписывает аргумент командной строки)")
	flagSet.BoolVar(&decrypt, "d", false, "Режим расшифровки")
	flagSet.BoolVar(&decrypt, "decrypt", false, "Режим расшифровки")
	flagSet.StringVar(&key, "p", "", "Пароль для шифрования/расшифровки")
	flagSet.StringVar(&key, "password", "", "Пароль для шифрования/расшифровки")
	flagSet.BoolVar(&useZip, "z", false, "Использовать ZIP-архивацию")
	flagSet.BoolVar(&useZip, "zip", false, "Использовать ZIP-архивацию")

	// Парсим все флаги кроме имени программы и первого аргумента (директории)
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Если первый аргумент не флаг, а директория, удаляем его из списка аргументов для парсинга
		args = args[1:]
	}
	flagSet.Parse(args)

	if showHelp {
		printHelp()
		return 0
	}

	if showVersion {
		fmt.Printf("Files-V версия %s\n", version)
		return 0
	}

	// Используем директорию из флага, если она была указана через --dir
	if directory == "" || directory != targetDir {
		directory = targetDir
	}

	// Проверяем, что директория указана
	if directory == "" {
		fmt.Println("Ошибка: не указана директория")
		fmt.Println("\nПравильное использование:")
		fmt.Println("  ./files-v <путь_к_директории> [опции]")
		return 1
	}

	// Логируем информацию о запуске (только для Windows)
	file.LogInfo("Запуск с параметрами: директория='%s', режим='%s', ZIP=%v",
		directory,
		map[bool]string{true: "расшифровка", false: "шифрование"}[decrypt],
		useZip)

	// Проверяем режим работы (шифрование или расшифровка)
	isDecryptMode := decrypt || strings.HasPrefix(filepath.Base(directory), "__fv__") ||
		file.IsEncryptedZipFile(directory)

	// Используем ZIP архивацию если указан флаг или если файл имеет расширение .zip.efv
	useZipMode := useZip || file.IsEncryptedZipFile(directory)

	// Проверяем, существует ли директория или файл
	if !utils.FileExists(directory) {
		err := fmt.Errorf("директория или файл '%s' не существует", directory)
		file.LogError("Ошибка: %v", err)
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("\nУкажите существующую директорию или файл:")
		fmt.Println("  ./files-v <существующая_директория_или_файл> [опции]")
		return 1
	}

	// Если ключ не был указан явно, запрашиваем его
	if key == "" {
		var err error
		var promptText string

		if isDecryptMode {
			promptText = "Введите пароль для расшифровки: "
		} else {
			promptText = "Введите пароль для шифрования: "
		}

		// Windows-специфичная проверка для логирования
		if file.LogFile.Enabled {
			file.LogInfo("Запрос пароля у пользователя")
		}

		key, err = readPassword(promptText)
		if err != nil {
			file.LogError("Ошибка при чтении пароля: %v", err)
			fmt.Printf("Ошибка при чтении пароля: %v\n", err)
			return 1
		}

		// Проверяем, что ключ не пустой
		if key == "" {
			err := fmt.Errorf("пароль не может быть пустым")
			file.LogError("Ошибка: %v", err)
			fmt.Printf("Ошибка: %v\n", err)
			return 1
		}
	}

	// Устанавливаем ключ шифрования
	file.SetEncryptionKey(key)

	// Обработка в соответствии с режимом
	if isDecryptMode {
		// Режим расшифровки
		if file.IsEncryptedZipFile(directory) {
			// Расшифровка и распаковка ZIP-архива
			extractDir, err := file.ProcessEncryptedZipFile(directory)
			if err != nil {
				file.LogError("Ошибка при расшифровке и распаковке архива: %v", err)
				fmt.Printf("Ошибка при расшифровке и распаковке архива: %v\nВозможно, введен неверный пароль.\n", err)
				return 1
			}
			fmt.Printf("Архив успешно расшифрован и распакован в: %s\n", extractDir)
			return 0
		} else {
			// Стандартная расшифровка директории
			_, err := file.PrepareDecryptedDirectory(directory)
			if err != nil {
				file.LogError("Ошибка при расшифровке директории: %v", err)
				fmt.Printf("Ошибка при расшифровке директории: %v\nВозможно, введен неверный пароль.\n", err)
				return 1
			}
			return 0
		}
	} else {
		// Режим шифрования
		if useZipMode {
			// Проверяем, является ли указанный путь директорией
			fileInfo, err := os.Stat(directory)
			if err != nil {
				file.LogError("Ошибка при получении информации о файле: %v", err)
				fmt.Printf("Ошибка при получении информации о файле: %v\n", err)
				return 1
			}

			if !fileInfo.IsDir() {
				err := fmt.Errorf("для ZIP-архивации необходимо указать директорию")
				file.LogError("Ошибка: %v", err)
				fmt.Printf("Ошибка: %v\n", err)
				return 1
			}

			// Создаем ZIP-архив, шифруем его и сохраняем с расширением .zip.efv
			encryptedPath, err := file.ProcessDirectoryWithZip(directory)
			if err != nil {
				file.LogError("Ошибка при создании и шифровании ZIP-архива: %v", err)
				fmt.Printf("Ошибка при создании и шифровании ZIP-архива: %v\n", err)
				return 1
			}
			fmt.Printf("Директория успешно заархивирована и зашифрована: %s\n", encryptedPath)
			return 0
		} else {
			// Стандартное шифрование директории
			_, err := file.PrepareDirectory(directory)
			if err != nil {
				fmt.Printf("Предупреждение: %v\n", err)
				fmt.Println("Процесс шифрования завершен, но некоторые файлы были пропущены.")
				if file.EncryptError.Count > 0 {
					fmt.Printf("Список пропущенных файлов (%d):\n", file.EncryptError.Count)
					for _, errFile := range file.EncryptError.Files {
						fmt.Printf("  - %s\n", errFile)
					}
				}
				return 0
			}
			return 0
		}
	}
}

// printHelp выводит справку по использованию
func printHelp() {
	fmt.Println("Files-V - утилита для работы с файлами")
	fmt.Println("\nИспользование:")
	fmt.Println("  files-v <директория_или_файл> [опции]")
	fmt.Println("\nОпции:")
	fmt.Println("  -help       Показать справку")
	fmt.Println("  -v, -version Показать версию")
	fmt.Println("  -dir        Директория для работы (перезаписывает аргумент командной строки)")
	fmt.Println("  -d, -decrypt Режим расшифровки")
	fmt.Println("  -p, -password Пароль для шифрования/расшифровки (если не указан, будет запрошен)")
	fmt.Println("  -z, -zip    Использовать ZIP-архивацию")
	fmt.Println("\nПримеры:")
	fmt.Println("  Шифрование директории (запросит пароль):")
	fmt.Println("    files-v /путь/к/директории")
	fmt.Println("  Шифрование директории с указанием пароля:")
	fmt.Println("    files-v /путь/к/директории -p мой_пароль")
	fmt.Println("  Шифрование директории с созданием ZIP-архива:")
	fmt.Println("    files-v /путь/к/директории -z")
	fmt.Println("  Расшифровка директории (запросит пароль):")
	fmt.Println("    files-v /путь/к/__fv__директории")
	fmt.Println("  Расшифровка архива .zip.efv с указанием пароля:")
	fmt.Println("    files-v /путь/к/архиву.zip.efv -p мой_пароль")
}
