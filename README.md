# 语雀知识库管理器（yuque-manager-gui）

<div align="center">

![语雀](https://img.shields.io/badge/语雀-知识库管理-blue)
![Wails](https://img.shields.io/badge/Wails-v2-green)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-orange)

语雀知识库 **下载** 与 **本地上传**（`.lake` / `.lakesheet`）桌面端工具；下载支持 `md` / `lake` 两种导出模式，多任务管理。打包产物文件名为 **yuque-manager-gui**（各平台可执行文件 / `.app`）。

</div>

## ✨ 功能特性

- 📚 **批量下载** - 同时管理多个下载任务
- 📊 **独立进度** - 每个任务独立的实时进度与状态
- 🎯 **任务管理** - 添加、删除、开始、取消；支持「开始全部」「清除完成」
- 📝 **按行导入链接** - 在新建任务中打开弹窗，按行粘贴 URL 批量生成知识库输入行
- 🧭 **文档类型精准识别** - 通过 `api/docs?book_id=...` 按 `slug` 匹配真实 `format/type`，区分 `lake` 与 `lakesheet`
- 🖼️ **Markdown 图片** - `md` 模式下可选择是否将文中图片下载到本地并替换为相对路径
- ⚠️ **图片失败策略** - `md` 模式下可选「继续保存文档」或「整篇标记失败」
- 🔐 **私有知识库** - 支持 Cookie
- ⚙️ **灵活配置** - 延迟、请求超时、图片超时、重试、并发、下载模式
- 📄 **双导出模式** - `md`（Markdown 接口）与 `lake`（Lake 导出）
- 🌓 **主题** - 浅色 / 深色；深色下表单输入框与整体面板一致
- 🚀 **跨平台** - Windows、macOS、Linux（Wails 构建）
- 📤 **本地上传** - 切换为「上传」模式：选择本地目录、填写知识库 URL / Book ID、Login、CToken、Cookie；可选创建目录节点、是否移动文档；上传进度与日志在页面展示（上传模式下顶部不展示下载任务统计）
- 🔄 **双模式界面** - 顶部在「下载 / 上传」间切换；下载页展示任务数量等统计并居中，上传页隐藏该项

## 🚀 快速开始

### 下载使用

1. 从 [Releases](https://github.com/DiCraft-Jerry/yuque-spider-gui/releases) 下载对应平台产物（包内可执行文件 / 应用名为 **yuque-manager-gui**）
2. 解压并运行
3. **下载**：先通过侧栏或「新建任务」里的 **选择目录** 指定保存根目录（路径为只读展示，需用系统目录选择器）；在「新建任务」中填写知识库 URL；私有库填写 Cookie；**添加** 任务后在列表中开始或 **开始全部**
4. **上传**：顶部切换到「上传」→ 选择本地上传目录，填写知识库 URL、Book ID、Referer、Login、CToken、Cookie 等 → **开始上传**（详见 [USAGE.md](USAGE.md)）

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
# 产物默认输出到 build/bin/yuque-manager-gui（或 yuque-manager-gui.app）
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

### 下载行为说明

- `lake` 模式下会先获取 `book_id` 对应的文档元信息（`/api/docs?book_id=...`），再按 `slug` 匹配每篇文档。  
- 当匹配到 `format=lakesheet`（或 `type=Sheet`）时，下载地址使用 `.../lakesheet?attachment=true`。  
- 普通文档（`format=lake` / `type=Doc`）使用 `.../lake?attachment=true`。  
- 当 TOC 节点既是文档又有子文档时：父文档文件与同名子目录并存（例如 `A.md` 与 `A/`）。

## 🛠️ 技术栈

- **后端**: Go 1.23
- **前端**: Svelte + Vite
- **桌面**: Wails v2

## 📁 项目结构（摘要）

```
yuque-spider-gui/          # 仓库目录名（Go module 仍为 yuque-spider-gui）
├── internal/spider/       # 下载：fetcher、downloader、spider、types
├── internal/uploader/     # 上传：Lake / Lakesheet 导入与目录节点
├── frontend/src/
│   ├── App.svelte         # 下载 / 上传双模式 UI
│   ├── components/ConfigDropdown.svelte
│   └── appApi.js
├── app.go                 # 任务、上传与 Wails 绑定
├── main.go
├── wails.json             # 应用名 / 输出文件名：yuque-manager-gui
└── .github/workflows/build.yml
```

### 基于原项目

- [yuque-crawl](https://github.com/burpheart/yuque-crawl)

## 🤝 贡献

欢迎 Issue 与 Pull Request。

## 📄 许可证

MIT License
