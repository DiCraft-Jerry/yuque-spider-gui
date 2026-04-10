# 语雀知识库下载器 GUI

<div align="center">

![语雀知识库下载器](https://img.shields.io/badge/语雀-知识库下载器-blue)
![Wails](https://img.shields.io/badge/Wails-v2-green)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-orange)

一个语雀知识库批量下载工具，支持 `md` / `lake` 两种导出模式，桌面端多任务管理。

</div>

## ✨ 功能特性

- 📚 **批量下载** - 同时管理多个下载任务
- 📊 **独立进度** - 每个任务独立的实时进度与状态
- 🎯 **任务管理** - 添加、删除、开始、取消；支持「开始全部」「清除完成」
- 📝 **批量导入** - 按行导入 `URL,Cookie(可选)`
- 🖼️ **Markdown 图片** - `md` 模式下可选择是否将文中图片下载到本地并替换为相对路径
- ⚠️ **图片失败策略** - `md` 模式下可选「继续保存文档」或「整篇标记失败」
- 🔐 **私有知识库** - 支持 Cookie
- ⚙️ **灵活配置** - 延迟、请求超时、图片超时、重试、并发、下载模式
- 📄 **双导出模式** - `md`（Markdown 接口）与 `lake`（Lake 导出）
- 🌓 **主题** - 浅色 / 深色；深色下表单输入框与整体面板一致
- 🚀 **跨平台** - Windows、macOS、Linux（Wails 构建）

## 🚀 快速开始

### 下载使用

1. 从 [Releases](https://github.com/DiCraft-Jerry/yuque-spider-gui/releases) 下载对应平台产物
2. 解压并运行
3. 先通过侧栏或「新建任务」里的 **选择目录** 指定保存根目录（路径为只读展示，需用系统目录选择器）
4. 填写知识库 URL；私有库填写 Cookie
5. 点击卡片右上角 **添加** 创建任务
6. 在任务列表点击 ▶️ 或使用 **开始全部** 下载

### 本地开发

#### 环境要求

- Go 1.23+
- Node.js 20+
- Wails CLI v2

#### 安装依赖

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
go mod tidy
cd frontend && npm install && cd ..
```

#### 运行与构建

```bash
wails dev
wails build
# 跨平台示例
wails build -platform windows/amd64,darwin/arm64,linux/amd64
```

### 配置说明（侧栏「下载配置」）

| 项 | 说明 |
| --- | --- |
| 延迟范围 | 文档之间的随机等待（秒） |
| 超时 | 单次请求超时（秒） |
| 图片超时 | 拉取图片时的超时（秒） |
| 文档类型 | `md` 或 `lake`；推荐一般场景优先 `lake` 更稳 |
| 文中图片 | **仅 `md`**：下载到本地并替换链接，或仅保留远程链接 |
| 图片下载失败策略 | **仅 `md`**：图片失败时仍保存文档，或整篇记为失败 |

## 🛠️ 技术栈

- **后端**: Go 1.23
- **前端**: Svelte + Vite
- **桌面**: Wails v2

## 📁 项目结构（摘要）

```
yuque-spider-gui/
├── internal/spider/       # 爬虫：fetcher、downloader、spider、types
├── frontend/src/
│   ├── App.svelte
│   ├── components/ConfigDropdown.svelte  # 文档类型等自定义下拉
│   └── appApi.js
├── app.go                 # 任务与 Wails 绑定
├── main.go
└── .github/workflows/build.yml
```

### 基于原项目

- [yuque-crawl](https://github.com/burpheart/yuque-crawl)

## 🤝 贡献

欢迎 Issue 与 Pull Request。

## 📄 许可证

MIT License
