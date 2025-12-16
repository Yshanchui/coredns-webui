package main

import (
	"coredns-webui/internal/handler"
	"coredns-webui/internal/middleware"
	"embed"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
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
	AuthUser     = getEnv("AUTH_USER", "admin")
	AuthPass     = getEnv("AUTH_PASS", "admin")
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

	// 使用 Auth 中间件
	r.Use(middleware.AuthMiddleware())

	// 创建 handler
	corefileHandler := &handler.CorefileHandler{
		CorefilePath: CorefilePath,
		CorednsPath:  CorednsPath,
	}

	// 登录页面
	r.GET("/login", func(c *gin.Context) {
		// 如果已经登录，直接跳转首页
		if cookie, err := c.Cookie(middleware.AuthCookieName); err == nil && cookie == middleware.AuthCookieValue {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// 登录 API
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == AuthUser && password == AuthPass {
			c.SetCookie(middleware.AuthCookieName, middleware.AuthCookieValue, 3600*24, "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		} else {
			c.HTML(http.StatusOK, "login.html", gin.H{"error": "Invalid username or password"})
		}
	})

	// Metrics 代理
	r.GET("/api/metrics", func(c *gin.Context) {
		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "coredns:9153"
			req.URL.Path = "/metrics"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// API 路由
	api := r.Group("/api")
	{
		api.GET("/corefile", corefileHandler.GetCorefile)
		api.POST("/corefile", corefileHandler.UpdateCorefile)
		api.POST("/corefile/validate", corefileHandler.ValidateCorefile)
		api.GET("/status", corefileHandler.GetStatus)
	}

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
