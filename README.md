# GitHub Fast DNS

一个GitHub定制版DNS服务，自动扫描GitHub域名对应IP并测速，返回响应最快的IP地址。

## 功能特性

- **GitHub官方IP范围扫描**：从 `https://api.github.com/meta` 获取GitHub官方公开的IP地址范围，展开并测速
- **智能IP优选**：定期扫描GitHub常用域名，测试所有IP的延迟，自动选择最优IP
- **DNS代理拦截**：本地DNS代理，拦截GitHub域名查询并返回最快IP
- **定时检测**：每5分钟自动更新IP延迟数据
- **上游转发**：非GitHub域名自动转发到上游DNS服务器
- **跨平台支持**：支持Windows、Linux、macOS
- **GUI预留接口**：ServiceManager 提供 Start/Stop/Status/GetBestIP 等接口，便于后续集成Fyne/Walk等GUI框架

## 项目结构

```
helloGithub/
├── src/                    # 源代码目录
│   ├── cmd/
│   │   └── main.go         # 主程序入口（使用ServiceManager管理生命周期）
│   └── internal/
│       ├── app/            # 服务管理器（GUI集成入口）
│       │   └── manager.go  # 统一管理DNS/Scanner/Scheduler生命周期
│       ├── config/         # 配置模块
│       ├── logger/         # 日志模块
│       ├── scanner/        # IP扫描器（含GitHub官方IP范围获取）
│       ├── dns/            # DNS代理服务器
│       ├── scheduler/      # 定时任务调度器
│       └── ippool/         # IP池管理
├── scripts/                # 编译和运行脚本
│   ├── build.bat           # Windows编译
│   ├── build.sh            # Linux/Mac编译
│   ├── run.bat             # Windows一键运行
│   └── run.sh              # Linux/Mac一键运行
├── doc/                    # 项目文档
├── tmp/                    # 临时文件和日志
├── config.yaml             # 配置文件
└── go.mod                  # Go模块定义
```

## 架构设计

### ServiceManager（服务管理器）

`internal/app/manager.go` 提供统一的组件管理接口，为GUI集成预留以下方法：

- `Start()` / `Stop()` - 启动/停止所有服务
- `Status()` - 获取服务运行状态（stopped/starting/running/stopping）
- `GetBestIP(domain)` - 获取指定域名的最优IP
- `GetAllDomains()` - 获取所有监控域名
- `GetPool()` - 获取IP池（用于展示所有IP及延迟数据）
- `TriggerScan()` - 手动触发一次扫描
- `GetConfig()` / `GetLogger()` - 获取配置和日志实例

GUI集成时只需创建 `ServiceManager` 实例，调用上述方法即可，无需直接操作底层模块。

## 快速开始

### Windows

1. 双击运行 `scripts/run.bat`
2. 脚本会自动请求管理员权限（DNS服务需要绑定53端口）
3. 如果未编译，脚本会自动调用 `build.bat` 进行编译

### Linux / macOS

```bash
# 编译
bash scripts/build.sh

# 运行（需要root权限绑定53端口）
sudo bash scripts/run.sh
```

## 配置说明

编辑 `config.yaml` 文件：

```yaml
dns:
  listen_addr: "127.0.0.1"    # DNS监听地址
  listen_port: 53             # DNS监听端口（需要管理员权限）

scanner:
  interval: 5m                # 扫描间隔
  timeout: 3s                 # 单次测速超时
  concurrency: 100           # 并发测速数量
  test_port: 443             # 测速目标端口
  max_latency: 5s            # 最大可接受延迟

logger:
  level: info                # 日志级别：debug/info/warn/error
  format: json               # 日志格式：json/text
  output_path: tmp/logs/app.log

# 监控的GitHub域名
domains:
  - github.com
  - api.github.com
  - raw.githubusercontent.com
  - ...

# 上游DNS服务器（用于非GitHub域名）
upstream:
  - 8.8.8.8:53
  - 114.114.114.114:53
```

## 使用方法

启动服务后，将系统DNS设置为 `127.0.0.1` 即可使用。

### Windows设置DNS

1. 打开 设置 -> 网络和Internet -> 更改适配器选项
2. 右键当前网络连接 -> 属性
3. 选择 "Internet协议版本4(TCP/IPv4)" -> 属性
4. 选择 "使用下面的DNS服务器地址"，填入 `127.0.0.1`
5. 点击确定

### Linux设置DNS

编辑 `/etc/resolv.conf`：

```bash
nameserver 127.0.0.1
```

或使用 systemd-resolved：

```bash
sudo resolvectl dns eth0 127.0.0.1
```

## 日志查看

日志文件位于 `tmp/logs/app.log`，同时会输出到控制台。

## 技术栈

- **Go 1.24**
- **miekg/dns** - DNS协议库
- **log/slog** - 结构化日志
- **YAML** - 配置文件

## 后续扩展

- 集成Fyne GUI框架，提供图形化界面
- 支持系统托盘运行
- 支持自动设置/恢复系统DNS
- 支持更多CDN域名优化
