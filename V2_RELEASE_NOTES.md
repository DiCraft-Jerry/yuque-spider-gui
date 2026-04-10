# v2 批量管理版本说明

本文档描述 **语雀知识库下载器 GUI** 当前主线能力：以仓库内实现为准，与 README / USAGE 同步。

---

## 1. 产品定位

基于 **Go + Wails v2 + Svelte** 的桌面应用，在单窗口内管理**多个语雀知识库下载任务**，支持 `md` / `lake` 两种导出模式、Cookie、可配置延迟与超时，以及 Markdown 图片本地化相关选项。

---

## 2. 后端：多任务与 API

### 2.1 任务模型

- 任务状态：`pending` → `running` → `completed` | `failed` | `cancelled`
- 每项任务包含：URL、Cookie、输出目录、`spider.Config`、进度结构、错误信息、时间戳等
- `sync.RWMutex` 保护任务映射与顺序列表；每个运行中任务具备独立 `context` 与取消函数

### 2.2 Wails 暴露方法（`app.go`）

| 方法 | 作用 |
| --- | --- |
| `GetDefaultConfig` | 返回默认爬虫配置 |
| `SelectDirectory` | 系统目录选择对话框 |
| `AddTask` | 添加任务（URL、Cookie、输出路径、配置） |
| `RemoveTask` | 删除任务；运行中会先取消 |
| `StartTask` | 启动单任务下载 |
| `CancelTask` | 取消运行中任务 |
| `GetAllTasks` | 获取任务列表快照 |
| `ClearCompletedTasks` | 移除已完成 / 失败 / 已取消任务 |
| `StartAllPendingTasks` | 依次启动所有 `pending` 任务 |
| `ValidateURL` | 校验 URL 是否包含语雀域名 |

### 2.3 前端事件

- `tasks:update`：任务列表整体变更  
- `task:update`：单个任务进度或状态变更  

---

## 3. 爬虫配置（`internal/spider.Config`）

| JSON 字段 | 含义 |
| --- | --- |
| `delayMin` / `delayMax` | 文档间隔随机延迟（秒） |
| `timeout` | 普通请求超时（秒） |
| `imageTimeout` | 图片请求超时（秒） |
| `maxRetries` | 重试次数 |
| `concurrentDownloads` | 并发下载数 |
| `downloadMode` | `md` 或 `lake` |
| `failOnImageError` | 图片失败时是否令该篇 Markdown 保存失败 |
| `skipMarkdownImages` | 为 `true` 时保存 `.md` 不下载文中图、不替换链接 |

默认值：`downloadMode` 为 `lake`，`failOnImageError` 为 `false`；`skipMarkdownImages` 未在 `DefaultConfig` 中显式赋值时为零值 `false`（即默认仍尝试本地化图片，且仅在走图片处理逻辑时生效）。

---

## 4. 界面结构（当前实现）

### 4.1 顶栏

- 左侧：应用图标 + 标题「语雀下载器」
- 中部：指标——总任务、运行中、等待中、已完成
- 右侧：使用提示（灯泡）、浅色 / 深色主题切换

### 4.2 左侧边栏（宽约 280px，可滚动）

1. **输出目录**  
   - 只读路径展示（虚线框样式，非输入框）  
   - **选择目录**：描边主色按钮（`btn-outline`），打开系统对话框  

2. **下载配置**  
   - 延迟范围、超时、图片超时（数字输入）  
   - **文档类型**：`ConfigDropdown` 自定义下拉（`md` / `lake`），固定定位菜单，避免 WebView 原生下拉遮挡  
   - **当且仅当文档类型为 `md` 时显示**：  
     - **文中图片**：分段控件——「下载到本地」/「仅保留链接」↔ `skipMarkdownImages`  
     - **图片下载失败策略**：分段控件——「继续下载文档」/「文档下载失败」↔ `failOnImageError`  
   - 每项下可有简短说明文案（`config-micro-hint`）

### 4.3 右侧主内容区（可滚动）

1. **新建任务**（卡片）  
   - 标题行：左侧「新建任务」，右侧主按钮 **添加**（`btn-add-task` 扁平主色样式）  
   - 表单栅格：知识库 URL、Cookie（文本输入）  
   - **保存路径**：只读路径展示 + **选择目录**；路径不可手输，与侧栏一致  
   - **深色模式**：URL / Cookie 输入框使用深色背景与 `var(--text-main)`，避免默认亮底刺眼  

2. **任务列表**（卡片）  
   - 标题行：统计文案、**开始全部**、**清除完成**  
   - 每条任务卡片：URL、状态徽章、知识库名（若有）、保存路径、进度条与文档计数、当前文档名；`pending` 显示开始，`running` 显示取消，非运行中可删除  

3. **URL 输入增强**（新建任务内）  
   - 支持动态新增 URL 输入行（可自定义数量）  
   - 支持 **按行导入** 弹窗：每行一个 URL，导入后自动扩充输入行  
   - 同团队场景下 Cookie 共用一次填写  

---

## 5. 下载与文件行为（摘要）

- 知识库输出在用户选定根目录下以知识库名建子目录；生成 `SUMMARY.md` 等逻辑见 `internal/spider`  
- `md` 且未开启「仅保留链接」时：对 `.md` 执行文中图片下载与链接替换（`downloader.go`）；失败行为由 `failOnImageError` 控制  
- `lake` 或非 `.md` 结果：不按 Markdown 图片逻辑处理侧栏中两项配置  

---

## 6. 前端工程

- `frontend/src/App.svelte`：主界面与样式  
- `frontend/src/components/ConfigDropdown.svelte`：配置项自定义下拉  
- `frontend/src/appApi.js`：Wails 绑定封装与开发期兜底  

构建与图标：`wails.json` 中 `icon` 指向 `build/appicon.png`；前端资源中另有 `appicon.png` 供界面品牌展示。

---

## 7. 与早期单任务 / CLI 形态对比

| 维度 | 早期形态（参考） | 当前 v2 桌面版 |
| --- | --- | --- |
| 任务数 | 单任务为主 | 多任务队列 |
| 界面 | 简单表单或 CLI | 顶栏 + 左右分栏管理台 |
| 进度 | 单进度 | 每任务独立进度与状态 |
| 批量 | 无或弱 | 按行导入链接、开始全部、清除完成 |
| 配置 | 部分硬编码 | 全套 `Config` + 按模式显隐的 UI |

---

## 8. 后续可改进方向（非承诺）

任务持久化、代理、增量同步、更细队列调度、单测与可观测性增强等；以 Issue / Roadmap 为准。

---

## 9. 文档与仓库

- 使用说明：[USAGE.md](USAGE.md)  
- 变更摘要：[CHANGELOG.md](CHANGELOG.md)  
- 项目说明：[README.md](README.md)  

发布包与 CI 见 `.github/workflows` 与 Releases。

---

## 10. 致谢
**前作者：[Spritualkb](https://github.com/Spritualkb)**
<br/>
**参考项目： [yuque-crawl](https://github.com/burpheart/yuque-crawl)** 、 **[旧yuque-spider-gui](https://github.com/Spritualkb/yuque-spider-gui)**

---

*文档与实现对齐维护；若界面或 API 有变，请同步更新本节。*
