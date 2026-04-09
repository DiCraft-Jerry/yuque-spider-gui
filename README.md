# 语雀知识库下载器 GUI

<div align="center">

![语雀知识库下载器](https://img.shields.io/badge/语雀-知识库下载器-blue)
![Wails](https://img.shields.io/badge/Wails-v2-green)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-orange)

一个语雀知识库批量下载工具,支持 `md` / `lake` 两种导出模式。

</div>

## ✨ 功能特性

- 📚 **批量下载** - 支持同时管理多个下载任务
- 📊 **独立进度** - 每个任务显示独立的实时进度条
- 🎯 **任务管理** - 添加、删除、开始、暂停任务
- 📝 **批量导入** - 从文本快速导入多个下载任务
- 🖼️ **图片处理** - 自动下载图片并支持失败策略
- 🔐 **私有知识库** - 支持使用 Cookie 访问私有知识库
- ⚙️ **灵活配置** - 延迟、超时、重试、下载模式
- 📄 **双下载模式** - 支持 `md` 导出和 `lake` 导出
- 🌓 **主题切换** - 支持白天 / 黑夜两种界面主题
- 🚀 **跨平台** - 支持 Windows、macOS、Linux

## 🚀 快速开始

### 下载使用

1. 从 [Releases](https://github.com/DiCraft-Jerry/yuque-spider-gui/releases) 页面下载适合你系统的版本
2. 解压并运行程序
3. 在右侧表单输入语雀知识库 URL
4. 选择保存路径
5. 点击"➕ 添加任务"
6. 点击任务卡片上的 ▶️ 按钮开始下载

### 本地开发

#### 环境要求

- Go 1.23+
- Node.js 20+
- Wails CLI v2

#### 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装项目依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install
cd ..
```

#### 运行开发模式

```bash
wails dev
```

#### 构建生产版本

```bash
# 构建当前平台
wails build

# 跨平台构建
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform darwin/arm64
wails build -platform linux/amd64
```

### 高级设置

- **最小/最大延迟**: 控制下载文档之间的等待时间
- **请求超时**: 网络请求的最长等待时间
- **图片超时**: 图片资源下载的超时时间
- **文档类型**:
  - `md`: 使用 `/markdown?attachment=true...` 导出
  - `lake`: 使用 `/lake?attachment=true` 导出
- **图片失败策略**:
  - 继续下载文档
  - 文档下载失败

## 🛠️ 技术栈

- **后端**: Go 1.23
- **前端**: Svelte
- **框架**: Wails v2
- **依赖**:
  - goquery - HTML 解析
  - Wails Runtime - 桌面应用框架

## 📁 项目结构

```
yuque-spider-gui/
├── internal/
│   └── spider/          # 爬虫核心逻辑
│       ├── types.go     # 数据结构定义
│       ├── fetcher.go   # 网络请求处理
│       ├── downloader.go # 文档和图片下载
│       └── spider.go    # 主爬虫逻辑
├── frontend/
│   └── src/
│       └── App.svelte   # 管理后台界面
├── app.go               # 应用后端接口(任务管理)
├── main.go              # 应用入口
└── .github/
    └── workflows/
        └── build.yml    # 自动构建配置
```

### 基于原项目
- 原项目: [yuque-crawl](https://github.com/burpheart/yuque-crawl)


## 🤝 贡献

欢迎提交 Issue 和 Pull Request!

## 📄 许可证

MIT License
