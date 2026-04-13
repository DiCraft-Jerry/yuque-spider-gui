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

1. 在 **知识库 URL** 区域填写链接（默认一行，可新增多行；第一行固定保留）  
2. 确认 **保存路径**（未选时先点 **选择目录**）  
3. 同团队知识库可复用同一个 **Cookie（可选）**  
4. 点击卡片标题栏右侧 **添加**（勿与「选择目录」混淆）  
5. 在任务列表对任务点击 **开始**，或使用 **开始全部**

### 5. 按行导入链接

1. 在 **知识库 URL** 区域点击 **按行导入**  
2. 弹窗中每行粘贴一个 URL  
3. 点击 **导入** 后，会按行自动新增知识库 URL 输入行  
4. 导入后同样点击 **添加** 创建任务

### 6. 手动新增多行

1. 在「添加链接」左侧输入新增数量（最小 1）  
2. 点击 **添加链接** 一次新增对应条数输入行

### 7. 进度与状态

任务卡片展示：知识库名、当前文档、进度条、完成数/总数、错误信息等。

### 8. 文档类型判定（lake 模式）

1. 先读取知识库 TOC（目录结构）用于构建层级与顺序  
2. 再调用 `GET /api/docs?book_id=...` 获取文档真实元信息  
3. 通过 `slug` 匹配后，按 `format/type` 选择下载地址：  
   - `lakesheet` / `Sheet` -> `.../lakesheet?attachment=true`  
   - `lake` / `Doc` -> `.../lake?attachment=true`

### 9. 父文档与子文档

- 若某节点既是文档又挂子文档，会同时生成：  
  - 父文档文件：`A.md`  
  - 子文档目录：`A/`（子文档在此目录下）

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

支持多任务、按行导入链接、开始全部、清除完成。

## 技术支持

- Issues: [DiCraft-Jerry/yuque-spider-gui](https://github.com/DiCraft-Jerry/yuque-spider-gui/issues)  
- 更新记录见 [CHANGELOG.md](CHANGELOG.md)
