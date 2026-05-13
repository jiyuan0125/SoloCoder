package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"go-api-doc-gen/internal/handlers"
	"go-api-doc-gen/internal/storage"
)

func main() {
	port := flag.Int("port", 8200, "服务端口")
	dbPath := flag.String("db", "./api_docs.db", "SQLite 数据库路径")
	flag.Parse()

	store, err := storage.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	h := handlers.New(store)

	mux := http.NewServeMux()

	h.RegisterRoutes(mux)

	staticDir := getStaticDir()
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		} else {
			http.NotFound(w, r)
		}
	})

	fmt.Printf("API 文档管理系统启动中...\n")
	fmt.Printf("数据库: %s\n", *dbPath)
	fmt.Printf("静态资源目录: %s\n", staticDir)
	fmt.Printf("服务地址: http://localhost:%d\n", *port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)
		os.Exit(1)
	}
}

func getStaticDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "./web/static"
	}

	exeDir := filepath.Dir(exePath)

	staticDir := filepath.Join(exeDir, "web", "static")
	if _, err := os.Stat(staticDir); err == nil {
		return staticDir
	}

	workDir, err := os.Getwd()
	if err == nil {
		staticDir = filepath.Join(workDir, "web", "static")
		if _, err := os.Stat(staticDir); err == nil {
			return staticDir
		}
	}

	return "./web/static"
}
