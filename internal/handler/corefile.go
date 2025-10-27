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

	// 使用 coredns -conf <file> -plugins 验证配置文件
	cmd := exec.Command(h.CorednsPath, "-conf", tmpFile, "-plugins")
	output, err := cmd.CombinedOutput()

	// 如果命令执行失败，检查输出中是否有错误信息
	if err != nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "Error") || strings.Contains(outputStr, "error") {
			return fmt.Errorf("配置验证失败: %s", outputStr)
		}
	}

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
