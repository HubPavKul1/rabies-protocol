package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
)

//go:embed all:web
var assets embed.FS

func main() {
	// 1. Инициализация базы данных
	initDB()

	// 2. Настройка роутера
	mux := http.NewServeMux()

	// Статические файлы (HTML, CSS, JS)
	webFS, _ := fs.Sub(assets, "web")
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	// 3. Регистрация API обработчиков
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/vets", handleVets)
	mux.HandleFunc("/api/owners", handleOwners)
	mux.HandleFunc("/api/animals", handleAnimals)
	
	// Протоколы
	mux.HandleFunc("/api/protocols", handleCreateProtocol)
	mux.HandleFunc("/api/protocols/list", handleListProtocols)
	mux.HandleFunc("/api/protocol/", handleGetProtocol)
	mux.HandleFunc("/api/protocol/edit/", handleEditProtocol)
	mux.HandleFunc("/api/protocol/update/", handleUpdateProtocol)
	mux.HandleFunc("/api/protocol/delete/", handleDeleteProtocol)
	
	// Экспорт документов
	mux.HandleFunc("/api/protocol-word/", handleGetProtocolWord)
	mux.HandleFunc("/api/protocol-docx/", handleGetProtocolDocx)

	// 4. Запуск сервера
	port := "8080"
	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("🚀 Сервер запущен: %s\n", url)
	
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	log.Fatal(http.ListenAndServe(":"+port, mux))
}