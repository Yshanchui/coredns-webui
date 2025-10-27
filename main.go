package main

import (
	"coredns-webui/internal/handler"
	"embed"
	"html/template"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed web/templates/*
var templatesFS embed.FS

// 配置参数
var (
	ServerHost   = getEnv("SERVER_HOST", "0.0.0.0")
	ServerPort   = getEnv("SERVER_PORT", "80")
	CorefilePath = getEnv("COREFILE_PATH", "/etc/coredns/Corefile")
	CorednsPath  = getEnv("COREDNS_PATH", "coredns")
)

func main() {
	log.Printf("配置信息:")
	log.Printf("  监听地址: %s:%s", ServerHost, ServerPort)
	log.Printf("  Corefile 路径: %s", CorefilePath)
	log.Printf("  CoreDNS 二进制: %s", CorednsPath)

	r := gin.Default()

	// 加载嵌入的模板文件
	tmpl := template.Must(template.New("").ParseFS(templatesFS, "web/templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	// 创建 handler
	corefileHandler := &handler.CorefileHandler{
		CorefilePath: CorefilePath,
		CorednsPath:  CorednsPath,
	}

	// API 路由
	r.GET("/api/corefile", corefileHandler.GetCorefile)
	r.POST("/api/corefile", corefileHandler.UpdateCorefile)
	r.POST("/api/corefile/validate", corefileHandler.ValidateCorefile)
	r.GET("/api/status", corefileHandler.GetStatus)

	// 页面路由
	r.GET("/", func(c *gin.Context) {
		content, err := corefileHandler.ReadCorefile()
		if err != nil {
			c.HTML(500, "index.html", gin.H{"error": err.Error()})
			return
		}
		c.HTML(200, "index.html", gin.H{"content": content})
	})

	// 启动服务器
	addr := ServerHost + ":" + ServerPort
	log.Printf("启动 Web 服务器: http://%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
