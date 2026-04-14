# 项目总结 - 语雀知识库管理器

## 项目概况

| 项 | 内容 |
| --- | --- |
| 产品名 | 语雀知识库管理器（界面标题） |
| 打包名 | **yuque-manager-gui**（`wails.json` 的 `name` / `outputfilename`，CI 产物 zip/tar 前缀） |
| 技术栈 | Go 1.23、Wails v2、Svelte、Vite |
| 下载核心 | `internal/spider`（Fetcher / Downloader / Spider） |
| 上传核心 | `internal/uploader`（Lake / Lakesheet、`/api/import` 与 `/api/catalog_nodes`） |
| 前端 | `App.svelte`（下载/上传双模式）、`components/ConfigDropdown.svelte`、`appApi.js` |

## 功能清单（与当前代码一致）

### 下载与配置

- 语雀知识库整库下载；生成 `SUMMARY.md`
- `md` / `lake` 两种导出模式
- `md` 时可选是否本地化文中图片（`SkipMarkdownImages`）
- `md` 时图片失败策略：`FailOnImageError`
- 延迟、超时、图片超时、重试、并发、Cookie

### 上传

- 本地目录扫描 `.lake` / `.lakesheet`，按路径顺序上传
- 需配置：知识库 URL、Book ID、Referer、Login、CToken、Cookie、Base URL 等
- 可选：创建目录节点（TITLE）、不移动文档（`noMoveDocs`）、条数限制、上传间隔

### 任务与界面

- **下载模式**：多任务；新建任务支持多 URL、按行导入；侧栏输出目录与下载配置；顶部展示任务统计（总任务 / 运行中等）并居中
- **上传模式**：表单 + 任务列表（进度与日志）；**不展示**顶部下载任务统计；进度信息在卡片内居中展示
- 顶部 **下载 ⟷ 上传** 切换
- 浅色 / 深色主题

### 工程化

- `wails build` 多平台；GitHub Actions（见 `.github/workflows`）
- 文档：README、USAGE、CHANGELOG、本文件

## 目录结构（摘要）

```
yuque-spider-gui/
├── internal/spider/
├── internal/uploader/
├── frontend/src/
│   ├── App.svelte
│   ├── components/ConfigDropdown.svelte
│   ├── appApi.js
│   └── main.js
├── app.go
├── main.go
├── wails.json
└── README.md / USAGE.md / CHANGELOG.md
```

## 配置结构（后端）

- `internal/spider.Config`：`DelayMin`、`DelayMax`、`Timeout`、`ImageTimeout`、`MaxRetries`、`ConcurrentDownloads`、`DownloadMode`（`md`|`lake`）、`FailOnImageError`、`SkipMarkdownImages` 等
- 上传参数见 `internal/uploader.UploadConfig` 与 `app.go` 中 `RunUpload` 绑定

## 致谢

- 原项目：[yuque-crawl](https://github.com/burpheart/yuque-crawl)
- 框架：[Wails](https://wails.io)

---

文档随版本迭代更新；细节以源码与 README 为准。
