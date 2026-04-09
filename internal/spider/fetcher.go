package spider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Fetcher 网络请求处理器
type Fetcher struct {
	client *http.Client
	cookie string
	config Config
}

// NewFetcher 创建新的 Fetcher
func NewFetcher(cookie string, config Config) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		cookie: cookie,
		config: config,
	}
}

// FetchBookTitle 获取知识库标题
func (f *Fetcher) FetchBookTitle(rawURL string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	f.applyCommonHeaders(req)

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败,状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	title := doc.Find("title").Text()
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, " · 语雀", "")

	// 清理非法文件名字符
	invalidChars := regexp.MustCompile(`[\/\\:*?"<>|\n\r]`)
	title = invalidChars.ReplaceAllString(title, "-")

	// 从 URL 提取标识符
	re := regexp.MustCompile(`u\d+/([\w-]+)`)
	if matches := re.FindStringSubmatch(rawURL); len(matches) > 1 {
		title = matches[1] + "-" + title
	}

	return title, nil
}

// FetchBookData 获取知识库数据
func (f *Fetcher) FetchBookData(rawURL string) (*YuqueData, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	f.applyCommonHeaders(req)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 从页面中提取 JSON 数据
	re := regexp.MustCompile(`decodeURIComponent\("(.+?)"\)\);`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, fmt.Errorf("无法从页面提取数据")
	}

	decodedData, err := url.QueryUnescape(matches[1])
	if err != nil {
		return nil, err
	}

	var yuqueData YuqueData
	if err := json.Unmarshal([]byte(decodedData), &yuqueData); err != nil {
		return nil, err
	}

	return &yuqueData, nil
}

// FetchDocument 获取文档内容
func (f *Fetcher) FetchDocument(_ int, slug, docURL, bookURL string) (*DocData, error) {
	docSlug := normalizeDocSlug(slug)
	mode := strings.ToLower(strings.TrimSpace(f.config.DownloadMode))
	if mode == "" {
		mode = "lake"
	}

	if mode == "md" {
		mdDoc, mdErr := f.fetchDocumentFromMarkdown(docURL, docSlug, bookURL)
		if mdErr == nil {
			return mdDoc, nil
		}
		return nil, fmt.Errorf("文档下载失败: %w", mdErr)
	}

	lakeDoc, lakeErr := f.fetchDocumentFromLake(docURL, docSlug, bookURL)
	if lakeErr == nil {
		return lakeDoc, nil
	}
	return nil, fmt.Errorf("文档下载失败: %w", lakeErr)
}

// DownloadImage 下载图片
func (f *Fetcher) DownloadImage(imageURL string) ([]byte, error) {
	maxRetries := f.config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	timeoutSeconds := f.config.ImageTimeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	if timeoutSeconds < 10 {
		timeoutSeconds = 10
	}

	imageClient := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			return nil, err
		}
		f.applyCommonHeaders(req)
		req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

		resp, err := imageClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries && isRetryableImageError(err) {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, err
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt < maxRetries && isRetryableImageError(readErr) {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, readErr
		}

		if resp.StatusCode == http.StatusOK {
			return body, nil
		}

		lastErr = fmt.Errorf("图片下载失败,状态码: %d", resp.StatusCode)
		if attempt < maxRetries && resp.StatusCode >= 500 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		return nil, lastErr
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("图片下载失败")
}

func (f *Fetcher) applyCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://www.yuque.com/")
	req.Header.Set("Origin", "https://www.yuque.com")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	if f.cookie != "" {
		req.Header.Set("Cookie", f.cookie)
		if ctoken := extractCookieValue(f.cookie, "yuque_ctoken"); ctoken != "" {
			req.Header.Set("x-csrf-token", ctoken)
			req.Header.Set("x-xsrf-token", ctoken)
		}
	}
}

func extractCookieValue(cookieHeader, key string) string {
	parts := strings.Split(cookieHeader, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

func normalizeDocSlug(slug string) string {
	raw := strings.TrimSpace(slug)
	if raw == "" {
		return raw
	}

	if u, err := url.Parse(raw); err == nil && u.Path != "" {
		raw = u.Path
	}

	raw = strings.Trim(raw, "/")
	segments := strings.Split(raw, "/")
	if len(segments) > 0 {
		return segments[len(segments)-1]
	}
	return raw
}

func (f *Fetcher) fetchDocumentFromPage(docURL, docSlug, bookURL string) (*DocData, error) {
	fullURL := buildDocumentURL(docURL, docSlug, bookURL)
	if fullURL == "" {
		return nil, fmt.Errorf("文档页面地址为空")
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	f.applyCommonHeaders(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("文档页面请求失败,状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	page := string(body)

	sourceCode := extractJSONStringField(page, "sourcecode")
	if sourceCode == "" {
		return nil, fmt.Errorf("页面中未找到 sourcecode")
	}

	title := extractJSONStringField(page, "title")
	return &DocData{
		Title:      title,
		SourceCode: sourceCode,
		FileExt:    ".md",
		RawContent: []byte(sourceCode),
	}, nil
}

func (f *Fetcher) fetchDocumentFromLake(docURL, docSlug, bookURL string) (*DocData, error) {
	lakeURL := buildLakeURL(docURL, docSlug, bookURL)
	if lakeURL == "" {
		return nil, fmt.Errorf("lake 下载地址为空")
	}
	fmt.Printf("文档下载链接(%s): %s\n", docSlug, lakeURL)

	req, err := http.NewRequest("GET", lakeURL, nil)
	if err != nil {
		return nil, err
	}
	f.applyCommonHeaders(req)
	req.Header.Set("Accept", "text/markdown,text/plain,*/*")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lake 下载失败,状态码: %d, 链接: %s", resp.StatusCode, lakeURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	trimmed := strings.TrimSpace(content)
	if trimmed == "" || strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") || strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		return nil, fmt.Errorf("lake 响应不是文档内容, 链接: %s", lakeURL)
	}

	fileExt := detectLakeFileExt(resp.Header.Get("Content-Disposition"), resp.Header.Get("Content-Type"))
	if fileExt == "" {
		fileExt = ".md"
	}

	return &DocData{
		Title:      docSlug,
		SourceCode: content,
		FileExt:    fileExt,
		RawContent: body,
	}, nil
}

func (f *Fetcher) fetchDocumentFromMarkdown(docURL, docSlug, bookURL string) (*DocData, error) {
	markdownURL := buildMarkdownURL(docURL, docSlug, bookURL)
	if markdownURL == "" {
		return nil, fmt.Errorf("markdown 下载地址为空")
	}
	fmt.Printf("文档下载链接(%s): %s\n", docSlug, markdownURL)

	req, err := http.NewRequest("GET", markdownURL, nil)
	if err != nil {
		return nil, err
	}
	f.applyCommonHeaders(req)
	req.Header.Set("Accept", "text/markdown,text/plain,*/*")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("markdown 下载失败,状态码: %d, 链接: %s", resp.StatusCode, markdownURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	content := string(body)
	trimmed := strings.TrimSpace(content)
	if trimmed == "" || strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") || strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		return nil, fmt.Errorf("markdown 响应不是文档内容, 链接: %s", markdownURL)
	}

	return &DocData{
		Title:      docSlug,
		SourceCode: content,
		FileExt:    ".md",
		RawContent: body,
	}, nil
}

func extractJSONStringField(content, field string) string {
	pattern := fmt.Sprintf(`"%s":"((?:\\.|[^"\\])*)"`, regexp.QuoteMeta(field))
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}

	decoded, err := strconv.Unquote(`"` + match[1] + `"`)
	if err != nil {
		return ""
	}
	return decoded
}

func normalizeDocURL(docURL string) string {
	raw := strings.TrimSpace(docURL)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return "https://www.yuque.com" + raw
}

func buildLakeURL(docURL, docSlug, bookURL string) string {
	fullURL := buildDocumentURL(docURL, docSlug, bookURL)
	if fullURL == "" {
		return ""
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimRight(u.Path, "/") + "/lake"
	q := u.Query()
	q.Set("attachment", "true")
	u.RawQuery = q.Encode()
	return u.String()
}

func buildMarkdownURL(docURL, docSlug, bookURL string) string {
	fullURL := buildDocumentURL(docURL, docSlug, bookURL)
	if fullURL == "" {
		return ""
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimRight(u.Path, "/") + "/markdown"
	q := u.Query()
	q.Set("attachment", "true")
	q.Set("latexcode", "false")
	q.Set("anchor", "true")
	q.Set("linebreak", "true")
	q.Set("useMdai", "true")
	u.RawQuery = q.Encode()
	return u.String()
}

func buildDocumentURL(docURL, docSlug, bookURL string) string {
	// docURL 已经是完整链接时优先使用
	if strings.HasPrefix(docURL, "http://") || strings.HasPrefix(docURL, "https://") {
		return docURL
	}

	base, err := url.Parse(strings.TrimSpace(bookURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		// 最后兜底回旧逻辑，避免返回空
		if docURL != "" {
			return normalizeDocURL(docURL)
		}
		return normalizeDocURL(docSlug)
	}

	// 取知识库路径前缀: /{group}/{repo}
	baseSeg := splitPathSegments(base.Path)
	if len(baseSeg) < 2 {
		if docURL != "" {
			return normalizeDocURL(docURL)
		}
		return normalizeDocURL(docSlug)
	}
	prefix := "/" + baseSeg[0] + "/" + baseSeg[1]

	candidate := strings.TrimSpace(docURL)
	if candidate == "" {
		candidate = docSlug
	}
	candidate = strings.TrimSpace(candidate)
	if strings.HasPrefix(candidate, "/") {
		candidate = strings.Trim(candidate, "/")
		seg := splitPathSegments("/" + candidate)
		if len(seg) >= 3 {
			// 已是 /group/repo/doc 形式
			u := &url.URL{Scheme: base.Scheme, Host: base.Host, Path: "/" + strings.Join(seg, "/")}
			return u.String()
		}
	}

	seg := splitPathSegments("/" + candidate)
	docRef := candidate
	if len(seg) > 0 {
		docRef = seg[len(seg)-1]
	}

	u := &url.URL{
		Scheme: base.Scheme,
		Host:   base.Host,
		Path:   prefix + "/" + strings.Trim(docRef, "/"),
	}
	return u.String()
}

func splitPathSegments(path string) []string {
	raw := strings.Trim(path, "/")
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "/")
}

func detectLakeFileExt(contentDisposition, contentType string) string {
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			filename := params["filename"]
			if filename == "" {
				filename = params["filename*"]
			}
			if filename != "" {
				if strings.HasPrefix(strings.ToLower(filename), "utf-8''") {
					if decoded, err := url.QueryUnescape(filename[7:]); err == nil {
						filename = decoded
					}
				}
				if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
					return ext
				}
			}
		}
	}

	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "text/markdown"):
		return ".md"
	case strings.Contains(ct, "application/json"):
		return ".json"
	case strings.Contains(ct, "text/plain"):
		return ".txt"
	case strings.Contains(ct, "application/pdf"):
		return ".pdf"
	case strings.Contains(ct, "application/zip"):
		return ".zip"
	default:
		return ".md"
	}
}

func isRetryableImageError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "temporarily unavailable") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe")
}
