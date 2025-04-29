package main

import (
	"fmt"
	"os"
)

func main() {
	// Если запущен с аргументами, перенаправляем их в cmd/files-v
	if len(os.Args) > 1 {
		// В продакшн версии здесь будет прямой вызов кода из cmd/files-v
		// Для разработки просто выводим сообщение
	} else {
		fmt.Println("Ошибка: необходимо указать директорию!")
		fmt.Println("\nПравильное использование:")
		fmt.Println("  go run ./cmd/files-v <путь_к_директории>")
	}
}
