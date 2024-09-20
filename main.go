package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type RouteConfig struct {
	Target string `json:"target"`
	Port   int    `json:"port"`
}

type Config map[string]RouteConfig

var tmpl *template.Template

func main() {
	// 读取配置文件
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// 解析 HTML 模板
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// 创建 HTTP 处理函数
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(config, w, r)
	})

	// 启动服务器
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func readConfig(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func handleRequest(config Config, w http.ResponseWriter, r *http.Request) {
	for route, routeConfig := range config {
		if strings.HasPrefix(r.URL.Path, route) {
			proxyURL := fmt.Sprintf("%s:%d", routeConfig.Target, routeConfig.Port)
			originalPath := r.URL.Path
			trimmedPath := strings.TrimPrefix(r.URL.Path, route)
			if !strings.HasPrefix(trimmedPath, "/") {
				trimmedPath = "/" + trimmedPath
			}

			if r.Method == http.MethodGet {
				// 返回包含 iframe 的 HTML 页面
				data := struct {
					ProxyURL string
				}{
					ProxyURL: proxyURL + trimmedPath,
				}
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				err := tmpl.Execute(w, data)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				// 添加日志记录
				log.Printf("Serving iframe for request from %s to %s%s", originalPath, proxyURL, trimmedPath)
				return
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}
