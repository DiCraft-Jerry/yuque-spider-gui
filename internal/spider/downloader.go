package spider

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Downloader 文档下载器
type Downloader struct {
	fetcher    *Fetcher
	outputPath string
	config     Config
}

// NewDownloader 创建新的下载器
func NewDownloader(cookie string, outputPath string, config Config) *Downloader {
	return &Downloader{
		fetcher:    NewFetcher(cookie, config),
		outputPath: outputPath,
		config:     config,
	}
}

// SaveDocument 保存文档
func (d *Downloader) SaveDocument(bookID int, slug, docURL, bookURL, docType, title, parentPath string) error {
	// 获取文档内容
	docData, err := d.fetcher.FetchDocument(bookID, slug, docURL, bookURL, docType)
	if err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}

	ext := docData.FileExt
	if ext == "" {
		ext = ".md"
	}
	filePath := filepath.Join(d.outputPath, parentPath, cleanFileName(title)+ext)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 非 markdown 内容直接按原始字节保存
	if ext != ".md" {
		content := docData.RawContent
		if len(content) == 0 {
			content = []byte(docData.SourceCode)
		}
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
		return nil
	}

	content := docData.SourceCode
	if !d.config.SkipMarkdownImages {
		var processErr error
		content, processErr = d.processImages(docData.SourceCode, filepath.Dir(filePath))
		if processErr != nil {
			return processErr
		}
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// processImages 处理 Markdown 中的图片
func (d *Downloader) processImages(markdown, docDir string) (string, error) {
	// 创建 assets 目录
	assetsDir := filepath.Join(docDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return markdown, fmt.Errorf("创建图片目录失败: %w", err)
	}

	// 正则匹配图片链接
	imgRegex := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	var firstErr error

	result := imgRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		// 提取 URL
		urlRegex := regexp.MustCompile(`\((.*?)\)`)
		urlMatches := urlRegex.FindStringSubmatch(match)
		if len(urlMatches) < 2 {
			return match
		}

		imageURL := urlMatches[1]

		// 跳过非 HTTP 链接
		if !strings.HasPrefix(imageURL, "http") {
			return match
		}

		// 移除 URL 中的锚点
		imageURL = strings.Split(imageURL, "#")[0]

		// 生成图片文件名
		timestamp := time.Now().UnixMilli()
		ext := filepath.Ext(imageURL)
		if ext == "" {
			ext = ".png"
		}
		imageName := fmt.Sprintf("image-%d%s", timestamp, ext)
		imageName = cleanFileName(imageName)

		// 下载图片
		imageData, err := d.fetcher.DownloadImage(imageURL)
		if err != nil {
			fmt.Printf("图片下载失败 %s: %v\n", imageURL, err)
			if d.config.FailOnImageError && firstErr == nil {
				firstErr = fmt.Errorf("图片下载失败 %s: %w", imageURL, err)
			}
			return match
		}

		// 保存图片
		imagePath := filepath.Join(assetsDir, imageName)
		if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
			fmt.Printf("保存图片失败 %s: %v\n", imagePath, err)
			if d.config.FailOnImageError && firstErr == nil {
				firstErr = fmt.Errorf("保存图片失败 %s: %w", imagePath, err)
			}
			return match
		}

		// 返回新的 Markdown 链接
		return fmt.Sprintf("![image-%d](./assets/%s)", timestamp, imageName)
	})

	if firstErr != nil {
		return result, firstErr
	}
	return result, nil
}

// cleanFileName 清理文件名中的非法字符
func cleanFileName(name string) string {
	// 替换非法字符
	invalidChars := regexp.MustCompile(`[\/\\:*?"<>|\n\r]`)
	return invalidChars.ReplaceAllString(name, "_")
}
