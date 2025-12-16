# CoreDNS WebUI - 可视化管理面板

![CoreDNS WebUI](image.png)

一个现代化的 CoreDNS Web 管理界面。支持 Corefile 在线编辑、实时指标监控、配置验证以及自动化部署。

## ✨ 主要特性

- 🎨 **现代化 UI** - 采用玻璃拟态 (Glassmorphism) 设计的 Light 主题，清晰美观。
- 📊 **实时监控** - 集成 Prometheus 指标，展示 QPS、查询总量、错误数及各协议/IP版本流量统计。
- � **智能编辑器** - 自适应布局的 Corefile 编辑器。
- �️ **安全验证** - 保存配置前自动通过 `coredns -dryrun` 进行语法检查，防止错误配置导致服务崩溃。
- 🔄 **平滑重载** - 基于 shared volume 和 CoreDNS `reload` 插件，在 Docker 环境下实现无中断配置热更新。
- 🔐 **访问控制** - 内置简单的 Cookie 登录认证。
- 🐳 **Docker 部署** - 提供完整的 `docker-compose` 方案，一键启动。

## 🚀 快速开始

### 使用 Docker Compose (推荐)

项目已配置好完整的 Docker 环境，包含 CoreDNS 和 WebUI。

1. **启动服务**
   ```bash
   docker-compose up -d
   ```

2. **访问面板**
   打开浏览器访问 [http://localhost:8080](http://localhost:8080)

3. **默认账号**
   - 用户名: `admin`
   - 密码: `admin`

### 自定义配置

你可以通过修改 `docker-compose.yml` 中的环境变量来自定义认证信息：

```yaml
    environment:
      - AUTH_USER=customUser
      - AUTH_PASS=customPass
```

## 🛠️ 功能说明

### 1. 仪表盘 (Dashboard)
首页直观展示 CoreDNS 的运行状况：
- **查询速率 (QPS)**: 实时计算每秒查询量。
- **协议统计**: 展示 IPv4/IPv6 及 UDP/TCP 的请求分布。
- **资源监控**: 显示 Go 运行时的内存占用。

### 2. 配置编辑器
- 支持全屏自适应高度。
- 点击 "保存配置" 时：
    1. 后端调用内置 `coredns` 二进制文件进行语法验证。
    2. 验证通过后写入共享卷。
    3. CoreDNS 进程自动感知文件变化并平滑重载 (Reload)，无需重启容器。

## 📝 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `SERVER_PORT` | `80` | WebUI 监听端口 |
| `COREFILE_PATH` | `/etc/coredns/Corefile` | Corefile 路径 |
| `AUTH_USER` | `admin` | 登录用户名 |
| `AUTH_PASS` | `admin` | 登录密码 |

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。
