# 使用指南

## 安装与运行

### Windows 用户

1. 下载 `yuque-spider-gui-vX.X.X-windows-amd64.zip`
2. 解压到任意目录
3. 双击 `yuque-spider-gui.exe` 运行

### macOS 用户

1. 下载对应芯片的版本:
   - Intel 芯片: `yuque-spider-gui-vX.X.X-darwin-amd64.tar.gz`
   - Apple Silicon: `yuque-spider-gui-vX.X.X-darwin-arm64.tar.gz`

2. 解压文件:
   ```bash
   tar -xzf yuque-spider-gui-vX.X.X-darwin-*.tar.gz
   ```

3. 运行应用:
   - 双击 `yuque-spider-gui.app`
   - 或命令行: `open yuque-spider-gui.app`

4. 如果遇到"无法打开,因为无法验证开发者"的提示:
   ```bash
   xattr -cr yuque-spider-gui.app
   ```

### Linux 用户

1. 下载 `yuque-spider-gui-vX.X.X-linux-amd64.tar.gz`
2. 解压并运行:
   ```bash
   tar -xzf yuque-spider-gui-vX.X.X-linux-amd64.tar.gz
   chmod +x yuque-spider-gui
   ./yuque-spider-gui
   ```

## 基础使用

### 1. 获取知识库 URL

访问你想要下载的语雀知识库,复制浏览器地址栏的 URL。

例如: `https://terminuscloud.yuque.com/uoaf0k/eyer97`

### 2. 获取 Cookie (仅私有知识库需要)

如果知识库是私有的,需要登录语雀后获取 Cookie:

1. 在 Chrome/Edge 浏览器中访问语雀
2. 按 `F12` 打开开发者工具
3. 切换到 `Application` (或 `应用`) 标签
4. 左侧找到 `Cookies` → 当前语雀域名 (例如 `https://terminuscloud.yuque.com`)
5. 复制所有 Cookie 值,格式类似:
   ```
   _yuque_session=xxx; yuque_ctoken=yyy; ...
   ```

### 3. 新建任务

1. 在右侧「新建任务」中填写:
   - 知识库 URL
   - Cookie (可选,私有库必填)
   - 保存路径
2. 点击「添加」
3. 在任务列表点击「开始」

### 4. 批量导入任务

1. 点击「批量导入」
2. 每行一个任务,格式: `URL,Cookie(可选)`
   ```text
   https://www.yuque.com/user/book1
   https://www.yuque.com/user/book2,_yuque_session=xxx; yuque_ctoken=yyy
   ```
3. 点击「导入」

### 5. 查看进度

任务卡片会显示:
- 当前知识库名称
- 正在下载的文档名称
- 进度条和百分比
- 已完成/总文档数
- 失败原因

## 高级设置说明

### 最小延迟和最大延迟

- **作用**: 控制下载每个文档之间的等待时间
- **建议值**: 1-4 秒
- **说明**:
  - 设置延迟可以避免请求过快被语雀限流
  - 如果遇到频繁失败,可以适当增加延迟
  - 私有知识库建议设置较长延迟(3-6秒)

### 请求超时

- **作用**: 单个网络请求的最长等待时间
- **建议值**: 30 秒
- **说明**:
  - 网络较慢时可以适当增加
  - 一般不需要修改默认值

### 图片超时

- **作用**: 下载图片资源时的超时时间
- **建议值**: 60-120 秒
- **说明**:
  - 图片通常比文档正文更容易超时
  - 网络波动较大时建议增大此值

### 文档类型

- **md**: 使用 `.../markdown?...` 导出 Markdown 文本
- **lake**: 使用 `.../lake?attachment=true` 导出附件格式
- **建议**:
  - 需要稳定 Markdown 输出时选 `md`
  - 需要和语雀导出一致时选 `lake`

### 图片失败策略

- **继续下载文档**: 图片失败时保留原链接,文档继续保存
- **文档下载失败**: 只要图片下载失败,整篇文档标记失败
- **建议**:
  - 批量导出优先选「继续下载文档」
  - 对完整性要求高时选「文档下载失败」

## 常见问题

### Q: 下载失败怎么办?

**A:** 可能的原因和解决方法:

1. **URL 错误** - 检查 URL 是否完整正确
2. **Cookie 过期** - 重新获取 Cookie
3. **网络问题** - 检查网络连接
4. **请求限流** - 增加延迟时间
5. **Cookie 缺失关键字段** - 确保包含 `_yuque_session` 和 `yuque_ctoken`

### Q: 为什么日志里是 markdown/lake 链接?

**A:**
- 程序会打印每篇文档实际请求链接,用于排查 401/404
- 多租户域名场景下,链接前缀应与知识库 URL 保持一致
- 若链接前缀不一致,请检查知识库 URL 是否填写正确

### Q: 部分图片无法显示?

**A:**
- 先提高「图片超时」配置
- 增加重试次数
- 将「图片失败策略」改为「继续下载文档」避免任务中断

### Q: 如何导入到其他笔记软件?

**A:** 下载的文件是标准 Markdown 格式:

- **Obsidian**: 直接将文件夹作为 vault 打开
- **Notion**: 使用 Notion 的导入功能
- **Typora**: 直接打开 Markdown 文件
- **语雀**: 可以重新导入

### Q: SUMMARY.md 是什么?

**A:**
- 自动生成的目录索引文件
- 包含所有文档的链接
- 可以用于快速导航

### Q: 支持批量下载多个知识库吗?

**A:**
- 支持
- 可以批量导入多个任务并分别启动
- 支持「开始全部」和「清除完成」

## 技术支持

- GitHub Issues: [提交问题](https://github.com/your-username/yuque-spider-gui/issues)
- Email: 3058886310@qq.com

## 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新历史。
