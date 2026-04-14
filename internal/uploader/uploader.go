package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type PlanItem struct {
	Kind      string
	Title     string
	RelPath   string
	StemRel   string
	ParentRel string
	Content   string
	DocType   string
	FileName  string
}

type CatalogRef struct {
	UUID       string
	Title      string
	TocNodeURL string // 文档 slug，对应 import 表单里的 toc_node_url（与节点 uuid 一起标识父文档/文档文件夹）
}

// importParentFields 子文档导入时在 multipart 中携带父节点信息（语雀字段名 toc_node_url，值为父文档 slug）
type importParentFields struct {
	PlaceUnderParent bool
	UsePrependChild  bool // true：action=prependChild 且带 toc_node_url；false：仅 insert（再由 catalog_nodes 移动）
	UUID             string
	Title            string
	Slug             string // 父文档 slug，写入 toc_node_url
}

type UploadConfig struct {
	RootPath        string  `json:"rootPath"`
	BaseURL         string  `json:"baseURL"`
	BookID          int     `json:"bookID"`
	Referer         string  `json:"referer"`
	Login           string  `json:"login"`
	CToken          string  `json:"ctoken"`
	Cookie          string  `json:"cookie"`
	CreateCatalogs  bool    `json:"createCatalogs"`
	NoMoveDocs      bool    `json:"noMoveDocs"`
	Limit           int     `json:"limit"`
	SleepMinSeconds float64 `json:"sleepMinSeconds"`
	SleepMaxSeconds float64 `json:"sleepMaxSeconds"`
}

type Progress struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

type Uploader struct {
	client     *http.Client
	config     UploadConfig
	nodeByRel  map[string]CatalogRef
	docByRel   map[string]CatalogRef
	titleRegex *regexp.Regexp
	logger     func(string)
}

func New(config UploadConfig) *Uploader {
	timeout := 60 * time.Second
	return &Uploader{
		client:     &http.Client{Timeout: timeout},
		config:     config,
		nodeByRel:  make(map[string]CatalogRef),
		docByRel:   make(map[string]CatalogRef),
		titleRegex: regexp.MustCompile(`(?is)<title>(.*?)</title>`),
	}
}

func (u *Uploader) Run(progress func(Progress), logf func(string)) error {
	u.logger = logf
	if u.config.RootPath == "" || u.config.BookID <= 0 || u.config.CToken == "" || u.config.Cookie == "" || u.config.Login == "" {
		return fmt.Errorf("上传参数不完整")
	}

	rootAbs, err := filepath.Abs(u.config.RootPath)
	if err != nil {
		return err
	}
	stat, err := os.Stat(rootAbs)
	if err != nil || !stat.IsDir() {
		return fmt.Errorf("上传目录不存在: %s", rootAbs)
	}

	plan, err := u.buildPlan(rootAbs, u.config.CreateCatalogs)
	if err != nil {
		return err
	}
	if u.config.Limit > 0 && u.config.Limit < len(plan) {
		plan = plan[:u.config.Limit]
	}
	total := len(plan)
	u.logf("开始上传任务")
	u.logf(fmt.Sprintf("上传目录: %s", rootAbs))
	u.logf(fmt.Sprintf("book_id: %d, 计划条目: %d", u.config.BookID, total))

	for idx, item := range plan {
		u.logf(fmt.Sprintf("[%d/%d] %s %s", idx+1, total, item.Kind, item.RelPath))
		progress(Progress{
			Current: idx + 1,
			Total:   total,
			Kind:    item.Kind,
			Path:    item.RelPath,
			Status:  "running",
		})

		var runErr error
		if item.Kind == "catalog" {
			runErr = u.uploadCatalog(item)
		} else {
			runErr = u.uploadDoc(item)
		}

		if runErr != nil {
			progress(Progress{
				Current: idx + 1,
				Total:   total,
				Kind:    item.Kind,
				Path:    item.RelPath,
				Status:  "failed",
				Error:   runErr.Error(),
			})
			u.logf(fmt.Sprintf("失败: %s => %v", item.RelPath, runErr))
			continue
		}
		u.logf(fmt.Sprintf("完成: %s", item.RelPath))

		progress(Progress{
			Current: idx + 1,
			Total:   total,
			Kind:    item.Kind,
			Path:    item.RelPath,
			Status:  "completed",
		})

		if u.config.SleepMaxSeconds > 0 {
			min := u.config.SleepMinSeconds
			max := u.config.SleepMaxSeconds
			if min < 0 {
				min = 0
			}
			if max < min {
				max = min
			}
			wait := min
			if max > min {
				wait = min + rand.Float64()*(max-min)
			}
			time.Sleep(time.Duration(wait * float64(time.Second)))
		}
	}
	u.logf("上传任务结束")
	return nil
}

func (u *Uploader) uploadCatalog(item PlanItem) error {
	u.logf(fmt.Sprintf("创建目录节点: %s", item.Title))
	parentUUID := ""
	if item.ParentRel != "" {
		if ref, ok := u.nodeByRel[item.ParentRel]; ok {
			parentUUID = ref.UUID
		}
	}

	createResp, err := u.putCatalogNodes(map[string]any{
		"book_id":     u.config.BookID,
		"format":      "list",
		"target_uuid": nil,
		"action":      "insert",
		"type":        "TITLE",
	})
	if err != nil {
		return err
	}
	createdUUID := firstStringField(createResp, "uuid")
	if createdUUID == "" {
		return fmt.Errorf("创建目录失败，未返回 uuid")
	}
	u.logf(fmt.Sprintf("目录节点已创建 uuid=%s", createdUUID))

	_, err = u.putCatalogNodes(map[string]any{
		"book_id":   u.config.BookID,
		"format":    "list",
		"node_uuid": createdUUID,
		"action":    "edit",
		"title":     item.Title,
		"doc_id":    nil,
	})
	if err != nil {
		return err
	}

	if parentUUID != "" {
		u.logf(fmt.Sprintf("移动目录节点 %s -> parent=%s", createdUUID, parentUUID))
		_, err = u.putCatalogNodes(map[string]any{
			"book_id":     u.config.BookID,
			"format":      "list",
			"node_uuid":   createdUUID,
			"action":      "prependChild",
			"target_uuid": parentUUID,
		})
		if err != nil {
			return err
		}
	}
	u.nodeByRel[item.RelPath] = CatalogRef{UUID: createdUUID, Title: item.Title}
	return nil
}

func (u *Uploader) uploadDoc(item PlanItem) error {
	u.logf(fmt.Sprintf("导入文档: %s", item.Title))
	parentUUID := ""
	parentTitle := ""
	parentSlug := ""
	if item.ParentRel != "" {
		if ref, ok := u.docByRel[item.ParentRel]; ok {
			parentUUID = ref.UUID
			parentTitle = ref.Title
			parentSlug = ref.TocNodeURL
		} else if ref, ok := u.nodeByRel[item.ParentRel]; ok {
			parentUUID = ref.UUID
			parentTitle = ref.Title
			parentSlug = ref.TocNodeURL
		}
	}

	placeUnderParent := item.ParentRel != "" && parentUUID != "" && !u.config.NoMoveDocs
	// 父文档且已缓存 slug：import 请求带 toc_node_url + prependChild（与浏览器一致）；仅 TITLE 目录无 slug 时用 insert + catalog_nodes
	usePrependImport := placeUnderParent && strings.TrimSpace(parentSlug) != ""

	importResp, err := u.postImport(item, importParentFields{
		PlaceUnderParent: placeUnderParent,
		UsePrependChild:  usePrependImport,
		UUID:             parentUUID,
		Title:            parentTitle,
		Slug:             parentSlug,
	})
	if err != nil {
		return err
	}
	docID := extractImportDocID(importResp)
	docNodeUUID := extractNodeUUID(importResp)
	if docNodeUUID == "" && docID != 0 {
		docNodeUUID = findNodeUUIDByDocID(importResp, docID)
	}
	if docNodeUUID == "" && docID != 0 && strings.TrimSpace(parentUUID) != "" && strings.TrimSpace(parentTitle) != "" {
		u.logf("import 未含 node_uuid，尝试通过刷新父目录节点反查 doc_id")
		tree, refreshErr := u.refreshTreeByEditParent(parentUUID, parentTitle)
		if refreshErr != nil {
			u.logf(fmt.Sprintf("刷新目录树失败: %v", refreshErr))
		} else if tree != nil {
			docNodeUUID = findNodeUUIDByDocID(tree, docID)
		}
	}
	if docNodeUUID == "" && docID != 0 {
		u.logf("尝试 GET 目录树反查 doc_id")
		if tree, getErr := u.fetchCatalogTreeGET(); getErr != nil {
			u.logf(fmt.Sprintf("GET 目录树失败: %v", getErr))
		} else if tree != nil {
			docNodeUUID = findNodeUUIDByDocID(tree, docID)
		}
	}
	if docNodeUUID == "" {
		if docID != 0 {
			return fmt.Errorf("导入成功但未解析到目录树节点 uuid，doc_id=%d（响应里通常无 meta.node_uuid，需反查失败）", docID)
		}
		u.logf("导入成功但未返回 node_uuid，跳过")
		return nil
	}
	u.logf(fmt.Sprintf("文档节点 uuid=%s", docNodeUUID))

	tocSlug := extractTocSlugFromImportResponse(importResp)

	if placeUnderParent && !usePrependImport {
		u.logf(fmt.Sprintf("移动文档节点 %s -> parent=%s", docNodeUUID, parentUUID))
		_, err = u.putCatalogNodes(map[string]any{
			"book_id":     u.config.BookID,
			"format":      "list",
			"node_uuid":   docNodeUUID,
			"action":      "prependChild",
			"target_uuid": parentUUID,
		})
		if err != nil {
			return err
		}
	}
	u.docByRel[item.StemRel] = CatalogRef{UUID: docNodeUUID, Title: item.Title, TocNodeURL: tocSlug}
	return nil
}

func (u *Uploader) putCatalogNodes(payload map[string]any) (map[string]any, error) {
	endpoint := strings.TrimRight(u.config.BaseURL, "/") + "/api/catalog_nodes"
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", u.config.Cookie)
	req.Header.Set("Origin", strings.TrimRight(u.config.BaseURL, "/"))
	req.Header.Set("Referer", u.config.Referer)
	req.Header.Set("User-Agent", "yuque-manager-gui-uploader/1.0")
	req.Header.Set("X-CSRF-Token", u.config.CToken)
	req.Header.Set("X-Login", u.config.Login)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	return u.doJSON(req)
}

func (u *Uploader) refreshTreeByEditParent(parentUUID, parentTitle string) (map[string]any, error) {
	return u.putCatalogNodes(map[string]any{
		"book_id":   u.config.BookID,
		"format":    "list",
		"node_uuid": parentUUID,
		"action":    "edit",
		"title":     parentTitle,
		"doc_id":    nil,
	})
}

func (u *Uploader) fetchCatalogTreeGET() (map[string]any, error) {
	endpoint := fmt.Sprintf("%s/api/catalog_nodes?book_id=%d&format=list", strings.TrimRight(u.config.BaseURL, "/"), u.config.BookID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", u.config.Cookie)
	req.Header.Set("Origin", strings.TrimRight(u.config.BaseURL, "/"))
	req.Header.Set("Referer", u.config.Referer)
	req.Header.Set("User-Agent", "yuque-manager-gui-uploader/1.0")
	req.Header.Set("X-CSRF-Token", u.config.CToken)
	req.Header.Set("X-Login", u.config.Login)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	return u.doJSON(req)
}

func (u *Uploader) postImport(item PlanItem, parent importParentFields) (map[string]any, error) {
	endpoint := fmt.Sprintf("%s/api/import?ctoken=%s", strings.TrimRight(u.config.BaseURL, "/"), u.config.CToken)
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	_ = writer.WriteField("insert_to_catalog", "true")
	if parent.PlaceUnderParent && parent.UsePrependChild && strings.TrimSpace(parent.UUID) != "" && strings.TrimSpace(parent.Slug) != "" {
		_ = writer.WriteField("action", "prependChild")
		_ = writer.WriteField("target_uuid", strings.TrimSpace(parent.UUID))
		_ = writer.WriteField("toc_node_uuid", strings.TrimSpace(parent.UUID))
		_ = writer.WriteField("toc_node_url", strings.TrimSpace(parent.Slug))
		if strings.TrimSpace(parent.Title) != "" {
			_ = writer.WriteField("toc_node_title", parent.Title)
		}
		_ = writer.WriteField("create_from", "doc_toc")
	} else {
		_ = writer.WriteField("action", "insert")
	}
	docType := normalizeImportType(item.DocType)
	_ = writer.WriteField("book_id", fmt.Sprintf("%d", u.config.BookID))
	_ = writer.WriteField("type", docType)
	_ = writer.WriteField("import_type", "create")
	_ = writer.WriteField("filename", "file")
	fileName := item.FileName
	if strings.TrimSpace(fileName) == "" {
		fileName = item.Title + "." + docType
	}
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	_, _ = io.Copy(fileWriter, strings.NewReader(item.Content))
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, endpoint, &payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", u.config.Cookie)
	req.Header.Set("Origin", strings.TrimRight(u.config.BaseURL, "/"))
	req.Header.Set("Referer", u.config.Referer)
	req.Header.Set("User-Agent", "yuque-manager-gui-uploader/1.0")
	return u.doJSON(req)
}

func (u *Uploader) doJSON(req *http.Request) (map[string]any, error) {
	u.logf(fmt.Sprintf("请求: %s %s", req.Method, req.URL.String()))
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	u.logf(fmt.Sprintf("响应: %s %s status=%d", req.Method, req.URL.Path, resp.StatusCode))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("请求失败 status=%d body=%s", resp.StatusCode, string(body))
	}
	out := map[string]any{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("响应解析失败: %w", err)
	}
	return out, nil
}

func (u *Uploader) buildPlan(root string, createCatalogs bool) ([]PlanItem, error) {
	docRelSet := make(map[string]bool)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if isUploadDocFile(d.Name()) {
			rel, err := filepath.Rel(root, path)
			if err == nil {
				docRelSet[filepath.ToSlash(rel)] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	plan := make([]PlanItem, 0)
	if createCatalogs {
		dirs := make([]string, 0)
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil || !d.IsDir() || path == root {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err == nil {
				dirs = append(dirs, filepath.ToSlash(rel))
			}
			return nil
		})
		sort.Slice(dirs, func(i, j int) bool {
			di := len(strings.Split(dirs[i], "/"))
			dj := len(strings.Split(dirs[j], "/"))
			if di != dj {
				return di < dj
			}
			return dirs[i] < dirs[j]
		})
		for _, rel := range dirs {
			if docRelSet[rel+".lake"] || docRelSet[rel+".lakesheet"] {
				continue
			}
			parent := ""
			if idx := strings.LastIndex(rel, "/"); idx >= 0 {
				parent = rel[:idx]
			}
			title := rel
			if idx := strings.LastIndex(rel, "/"); idx >= 0 {
				title = rel[idx+1:]
			}
			plan = append(plan, PlanItem{Kind: "catalog", Title: title, RelPath: rel, ParentRel: parent})
		}
	}

	files := make([]string, 0)
	for rel := range docRelSet {
		files = append(files, rel)
	}
	sort.Slice(files, func(i, j int) bool {
		di := len(strings.Split(files[i], "/"))
		dj := len(strings.Split(files[j], "/"))
		if di != dj {
			return di < dj
		}
		return files[i] < files[j]
	})

	for _, rel := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		contentBytes, err := os.ReadFile(abs)
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)
		title := extractTitle(filepath.Base(rel), content, u.titleRegex)
		fileName := filepath.Base(rel)
		docType := detectImportTypeByFileName(fileName)
		stemRel := strings.TrimSuffix(rel, filepath.Ext(rel))
		parent := ""
		if idx := strings.LastIndex(rel, "/"); idx >= 0 {
			parent = rel[:idx]
		}
		plan = append(plan, PlanItem{
			Kind:      "doc",
			Title:     title,
			RelPath:   rel,
			StemRel:   stemRel,
			ParentRel: parent,
			Content:   content,
			DocType:   docType,
			FileName:  fileName,
		})
	}
	return plan, nil
}

func isUploadDocFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".lake") || strings.HasSuffix(lower, ".lakesheet")
}

func detectImportTypeByFileName(fileName string) string {
	lower := strings.ToLower(fileName)
	if strings.HasSuffix(lower, ".lakesheet") {
		return "lakesheet"
	}
	return "lake"
}

func normalizeImportType(docType string) string {
	t := strings.ToLower(strings.TrimSpace(docType))
	if strings.Contains(t, "lakesheet") || strings.Contains(t, "sheet") {
		return "lakesheet"
	}
	return "lake"
}

func extractTitle(filename, content string, re *regexp.Regexp) string {
	m := re.FindStringSubmatch(content)
	if len(m) > 1 {
		title := strings.TrimSpace(m[1])
		if title != "" {
			return title
		}
	}
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func firstStringField(v any, key string) string {
	switch t := v.(type) {
	case map[string]any:
		if s, ok := t[key].(string); ok && s != "" {
			return s
		}
		for _, child := range t {
			if s := firstStringField(child, key); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range t {
			if s := firstStringField(child, key); s != "" {
				return s
			}
		}
	}
	return ""
}

func anyToInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}

func extractImportDocID(v any) int {
	root, ok := v.(map[string]any)
	if !ok {
		return 0
	}
	return extractDocIDFromPayload(root["data"])
}

func extractDocIDFromPayload(v any) int {
	m, ok := v.(map[string]any)
	if !ok {
		return 0
	}
	if inner, ok := m["data"].(map[string]any); ok {
		if id := docLikePrimaryID(inner); id != 0 {
			return id
		}
	}
	if id := docLikePrimaryID(m); id != 0 {
		return id
	}
	return 0
}

func docLikePrimaryID(m map[string]any) int {
	if m == nil {
		return 0
	}
	id := anyToInt(m["id"])
	if id == 0 {
		return 0
	}
	if _, ok := m["book_id"]; ok {
		return id
	}
	if _, ok := m["slug"].(string); ok {
		return id
	}
	return 0
}

func findNodeUUIDByDocID(v any, docID int) string {
	if docID <= 0 {
		return ""
	}
	switch t := v.(type) {
	case map[string]any:
		nodeDoc := anyToInt(t["doc_id"])
		nodeUUID, _ := t["uuid"].(string)
		if nodeDoc == docID && strings.TrimSpace(nodeUUID) != "" {
			return strings.TrimSpace(nodeUUID)
		}
		for _, child := range t {
			if s := findNodeUUIDByDocID(child, docID); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range t {
			if s := findNodeUUIDByDocID(child, docID); s != "" {
				return s
			}
		}
	}
	return ""
}

func extractTocSlugFromImportResponse(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if s := tocSlugFromImportMap(m); s != "" {
		return s
	}
	if d, ok := m["data"].(map[string]any); ok {
		if s := tocSlugFromImportMap(d); s != "" {
			return s
		}
	}
	return ""
}

func tocSlugFromImportMap(m map[string]any) string {
	if meta, ok := m["meta"].(map[string]any); ok {
		if s, ok := meta["toc_node_url"].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	if s, ok := m["slug"].(string); ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	if d, ok := m["data"].(map[string]any); ok {
		if s, ok := d["slug"].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func extractNodeUUID(v any) string {
	switch t := v.(type) {
	case map[string]any:
		if meta, ok := t["meta"].(map[string]any); ok {
			if nodeUUID, ok := meta["node_uuid"].(string); ok && nodeUUID != "" {
				return nodeUUID
			}
		}
		if data, ok := t["data"]; ok {
			if s := extractNodeUUID(data); s != "" {
				return s
			}
		}
		for _, child := range t {
			if s := extractNodeUUID(child); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range t {
			if s := extractNodeUUID(child); s != "" {
				return s
			}
		}
	}
	return ""
}

func (u *Uploader) logf(line string) {
	if u.logger != nil && strings.TrimSpace(line) != "" {
		u.logger(line)
	}
}
