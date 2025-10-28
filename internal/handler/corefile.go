package handler

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type CorefileHandler struct {
	CorefilePath string
	CorednsPath  string
}

// 检查 CoreDNS 运行状态
func (h *CorefileHandler) GetStatus(c *gin.Context) {
	status := h.checkCoreDNSStatus()
	c.JSON(200, gin.H{
		"running": status,
		"status":  map[bool]string{true: "运行中", false: "未运行"}[status],
	})
}

func (h *CorefileHandler) checkCoreDNSStatus() bool {
	// 使用 systemctl 检查服务状态
	cmd := exec.Command("systemctl", "is-active", "coredns")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "active"
}

// 验证 Corefile 格式
func (h *CorefileHandler) validateCorefile(content string) error {
	// 创建临时文件
	tmpFile := h.CorefilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// 使用 timeout + coredns 验证配置
	// 正确的配置：coredns 会启动，timeout 1秒后退出，返回码 124
	// 错误的配置：coredns 启动失败，立即退出，返回码 1
	logFile := "/tmp/coredns_check.log"
	cmd := exec.Command("timeout", "1s", h.CorednsPath, "-conf", tmpFile, "-dns.port", "0")

	// 重定向输出到日志文件
	logF, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("无法创建日志文件: %v", err)
	}
	defer logF.Close()
	defer os.Remove(logFile)

	cmd.Stdout = logF
	cmd.Stderr = logF

	err = cmd.Run()

	// 读取日志内容
	logContent, _ := os.ReadFile(logFile)
	logStr := string(logContent)

	if err != nil {
		// 检查退出码
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()

			// 124 表示 timeout 正常退出（配置正确，coredns 启动成功）
			if exitCode == 124 {
				return nil
			}

			// 其他退出码表示配置错误
			if logStr != "" {
				return fmt.Errorf("配置验证失败:\n%s", logStr)
			}
			return fmt.Errorf("配置验证失败，退出码: %d", exitCode)
		}
		return fmt.Errorf("验证命令执行失败: %v", err)
	}

	// 如果没有错误，说明配置正确
	return nil
}

// 验证 API 端点
func (h *CorefileHandler) ValidateCorefile(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求格式"})
		return
	}

	if err := h.validateCorefile(req.Content); err != nil {
		c.JSON(200, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"valid":   true,
		"message": "配置文件验证通过",
	})
}

func (h *CorefileHandler) GetCorefile(c *gin.Context) {
	content, err := h.ReadCorefile()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"content": content})
}

func (h *CorefileHandler) UpdateCorefile(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求格式"})
		return
	}

	// 备份原配置
	oldContent, err := h.ReadCorefile()
	if err != nil {
		c.JSON(500, gin.H{"error": "无法读取原配置文件"})
		return
	}

	// 在写入之前验证配置
	if err := h.validateCorefile(req.Content); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 验证通过后再写入文件
	if err := os.WriteFile(h.CorefilePath, []byte(req.Content), 0644); err != nil {
		c.JSON(500, gin.H{"error": "写入配置文件失败: " + err.Error()})
		return
	}

	// 重启 CoreDNS
	if err := h.restartCoreDNS(); err != nil {
		// 如果重启失败，恢复原配置
		os.WriteFile(h.CorefilePath, []byte(oldContent), 0644)
		c.JSON(500, gin.H{"error": "重启 CoreDNS 失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "配置已保存并重启成功"})
}

func (h *CorefileHandler) restartCoreDNS() error {
	// 使用 systemctl restart 重启服务
	cmd := exec.Command("systemctl", "restart", "coredns")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("重启失败: %v, 输出: %s", err, string(output))
	}
	return nil
}

func (h *CorefileHandler) ReadCorefile() (string, error) {
	content, err := os.ReadFile(h.CorefilePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
