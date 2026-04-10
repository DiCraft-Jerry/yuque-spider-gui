# 使用指南

## 安装与运行

### Windows

1. 下载 `yuque-spider-gui-vX.X.X-windows-amd64.zip`
2. 解压后运行 `yuque-spider-gui.exe`

### macOS

1. 按芯片选择包：Intel `darwin-amd64` / Apple Silicon `darwin-arm64`
2. 解压：`tar -xzf yuque-spider-gui-vX.X.X-darwin-*.tar.gz`
3. 运行 `yuque-spider-gui.app` 或 `open yuque-spider-gui.app`
4. 若提示无法验证开发者：`xattr -cr yuque-spider-gui.app`

### Linux

```bash
tar -xzf yuque-spider-gui-vX.X.X-linux-amd64.tar.gz
chmod +x yuque-spider-gui
./yuque-spider-gui
```

## 基础使用

### 1. 知识库 URL

打开目标语雀知识库，复制浏览器地址栏完整 URL。  
示例：`https://terminuscloud.yuque.com/uoaf0k/eyer97`

### 2. Cookie（仅私有库）

1. Chrome/Edge 登录语雀 → F12 → **Application** → **Cookies** → 对应域名  
2. 复制为请求头风格字符串，例如：`_yuque_session=...; yuque_ctoken=...`

### 3. 输出目录与保存路径

- **侧栏「输出目录」** 与 **新建任务里的「保存路径」** 均为 **只读展示**，不能直接键入路径。  
- 点击 **选择目录** 打开系统文件夹对话框选定目录。  
- 每个知识库会在该目录下生成同名子文件夹。

### 4. 新建任务

1. 填写 **知识库 URL**、**Cookie（可选）**  
2. 确认 **保存路径**（未选时先点 **选择目录**）  
3. 点击卡片标题栏右侧 **添加**（勿与「选择目录」混淆）  
4. 在任务列表对任务点击 **开始**，或使用 **开始全部**

### 5. 批量导入

1. **批量导入** → 每行一条：`URL` 或 `URL,Cookie`  
2. **导入** 后同样需在列表中启动任务

### 6. 进度与状态

任务卡片展示：知识库名、当前文档、进度条、完成数/总数、错误信息等。

## 高级设置说明

### 延迟范围

文档之间的随机等待，减轻限流风险；私有库可适当加大。

### 请求超时 / 图片超时

分别作用于正文请求与图片下载；网络不稳时可增大 **图片超时**。

### 文档类型

- **lake**：语雀 Lake 导出，多数情况下更稳定，**建议优先**。  
- **md**：直接拉 Markdown 接口；网络波动时文中图片更容易超时。

### 文中图片（仅文档类型为 md 时显示）

- **下载到本地**：图片写入文档旁 `assets`，Markdown 改为相对路径。  
- **仅保留链接**：不下载图片，不建 `assets`，保留原文链接。

### 图片下载失败策略（仅文档类型为 md 时显示）

- **继续下载文档**：单张图失败时保留原图链，文档仍保存。  
- **文档下载失败**：任一张图失败则该篇文档记为失败。  

> 选择 **lake** 或未开启「下载到本地」时，上述策略对当前任务无实际效果，对应选项会隐藏。

## 常见问题

### 下载失败？

检查 URL、Cookie 是否过期、网络、延迟是否过小；私有库确保含 `_yuque_session`、`yuque_ctoken` 等关键字段。

### 日志里的 markdown / lake 链接？

为排查 401/404 打印的实际请求地址；多租户时注意与知识库 URL 域名一致。

### 部分图片打不开？

提高图片超时与重试；**md** 模式下可改为「继续下载文档」或改用 **lake**。

### SUMMARY.md？

自动生成的目录索引，便于导航。

### 多知识库？

支持多任务、批量导入、开始全部、清除完成。

## 技术支持

- Issues: [DiCraft-Jerry/yuque-spider-gui](https://github.com/DiCraft-Jerry/yuque-spider-gui/issues)  
- 更新记录见 [CHANGELOG.md](CHANGELOG.md)
