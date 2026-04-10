# 项目总结 - 语雀知识库下载器 GUI

## 项目概况

| 项 | 内容 |
| --- | --- |
| 名称 | 语雀知识库下载器 GUI |
| 技术栈 | Go 1.23、Wails v2、Svelte、Vite |
| 后端核心 | `internal/spider`（Fetcher / Downloader / Spider） |
| 前端 | `App.svelte`、`components/ConfigDropdown.svelte`、`appApi.js` |

## 功能清单（与当前代码一致）

### 下载与配置

- 语雀知识库整库下载；生成 `SUMMARY.md`  
- `md` / `lake` 两种导出模式  
- `md` 时可选是否本地化文中图片（`SkipMarkdownImages`）  
- `md` 时图片失败策略：`FailOnImageError`  
- 延迟、超时、图片超时、重试、并发、Cookie  

### 任务与界面

- 多任务：添加、删除、开始、取消、开始全部、清除完成  
- 新建任务支持多 URL 输入行（首行固定保留，可按数量新增）  
- 支持按行导入弹窗（每行一个 URL），导入后自动扩充 URL 输入行  
- 侧栏：输出目录（只读路径 + 选择目录）、下载配置  
- 主区：新建任务（添加在标题栏右侧）、任务列表与进度  
- 浅色 / 深色主题；深色下表单输入样式与面板统一  

### 工程化

- `wails build` 多平台；GitHub Actions（见 `.github/workflows`）  
- 文档：README、USAGE、CHANGELOG、本文件  

## 目录结构（摘要）

```
yuque-spider-gui/
├── internal/spider/
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

`internal/spider.Config` 主要字段包括：`DelayMin`、`DelayMax`、`Timeout`、`ImageTimeout`、`MaxRetries`、`ConcurrentDownloads`、`DownloadMode`（`md`|`lake`）、`FailOnImageError`、`SkipMarkdownImages`。

## 致谢

- 原项目：[yuque-crawl](https://github.com/burpheart/yuque-crawl)  
- 框架：[Wails](https://wails.io)  

---

文档随版本迭代更新；细节以源码与 README 为准。
