# Changelog

本项目 notable 变更记录于此文件。

## [Unreleased]

### Added

- **md 模式**：配置项「文中图片」——是否将 Markdown 内图片下载到本地并替换为相对路径（`skipMarkdownImages`）。  
- **md 模式**：「图片下载失败策略」与「文中图片」同为分段切换，仅在文档类型为 `md` 时显示。  
- 侧栏 **文档类型** 使用自定义下拉（`ConfigDropdown`），避免系统原生下拉在 WebView 中遮挡与样式不一致。  
- 输出目录 / 保存路径 **只读路径展示**（非输入框样式）；**选择目录** 使用描边按钮样式。  
- **新建任务**：标题与 **添加** 同排；表单区与标题区分隔更清晰。  
- **深色模式**：主内容区 URL / Cookie 文本框背景与文字对比优化。

### Changed

- 应用图标资源更新（`frontend/src/assets/appicon.png`、`build/appicon.png`）。  
- 多任务管理、事件推送、Wails 绑定等保持与 v2 架构一致；文档与 README 与当前界面、配置项对齐。

### Fixed

- 图片下载与 Markdown 处理相关错误处理、重试等持续优化（见 `internal/spider` 提交记录）。

## [2.x] 批量任务与 GUI

- Go + Wails + Svelte 桌面应用。  
- 多任务、独立进度、批量导入、开始全部、清除完成。  
- 左右分栏：输出目录与下载配置 / 新建任务与任务列表。  
- 浅色与深色主题。

## [原项目] Python CLI

基于 [yuque-crawl](https://github.com/burpheart/yuque-crawl)。
