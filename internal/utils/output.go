package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ResultType 定义结果类型
type ResultType string

const (
	FILE_FOUND    ResultType = "FILE_FOUND" // 文件发现
	CONTENT_MATCH ResultType = "CONTENT"    // 内容匹配
	PERM_MATCH    ResultType = "PERMISSION" // 权限匹配
	TIME_MATCH    ResultType = "TIME"       // 时间匹配
)

// SearchResult 搜索结果结构
type SearchResult struct {
	Time        time.Time              `json:"time"`        // 发现时间
	Type        ResultType             `json:"type"`        // 结果类型
	Path        string                 `json:"path"`        // 文件路径
	Size        int64                  `json:"size"`        // 文件大小
	ModTime     string                 `json:"mod_time"`    // 修改时间
	Permissions string                 `json:"permissions"` // 权限
	MatchType   string                 `json:"match_type"`  // 匹配类型
	MatchCount  int                    `json:"match_count"` // 匹配次数
	Content     string                 `json:"content"`     // 内容预览
	Details     map[string]interface{} `json:"details"`     // 详细信息
	Keyword     string                 `json:"keyword"`     // 匹配的关键字
}

// OutputManager 输出管理器
type OutputManager struct {
	mu            sync.Mutex
	outputPath    string
	outputFormat  string
	file          *os.File
	csvWriter     *csv.Writer
	jsonEncoder   *json.Encoder
	isInitialized bool
	results       []*SearchResult // 缓存结果用于终端显示
}

// GlobalOutputManager 全局输出管理器实例
var GlobalOutputManager *OutputManager

// InitOutputManager 初始化输出管理器
// outputPath: 输出文件路径
// outputFormat: 输出格式 (txt, json, csv)
func InitOutputManager(outputPath, outputFormat string) error {
	if outputFormat == "" {
		outputFormat = "txt" // 默认格式
	}

	// 验证输出格式
	switch outputFormat {
	case "txt", "json", "csv":
		// 有效的格式
	default:
		return fmt.Errorf("不支持的输出格式: %s", outputFormat)
	}

	// 如果指定了输出路径，确保目录存在
	if outputPath != "" {
		dir := filepath.Dir(outputPath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建输出目录失败: %v", err)
			}
		}
	}

	manager := &OutputManager{
		outputPath:   outputPath,
		outputFormat: outputFormat,
		results:      make([]*SearchResult, 0),
	}

	if outputPath != "" {
		if err := manager.initialize(); err != nil {
			return err
		}
	}

	GlobalOutputManager = manager
	return nil
}

// initialize 初始化文件输出
func (om *OutputManager) initialize() error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.isInitialized || om.outputPath == "" {
		return nil
	}

	file, err := os.OpenFile(om.outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开输出文件失败: %v", err)
	}
	om.file = file

	switch om.outputFormat {
	case "csv":
		om.csvWriter = csv.NewWriter(file)
		headers := []string{"Time", "Type", "Path", "Size", "ModTime", "Permissions", "MatchType", "MatchCount", "Content"}
		if err := om.csvWriter.Write(headers); err != nil {
			file.Close()
			return fmt.Errorf("写入CSV头部失败: %v", err)
		}
		om.csvWriter.Flush()
	case "json":
		om.jsonEncoder = json.NewEncoder(file)
		om.jsonEncoder.SetIndent("", "  ")
	case "txt":
		// TXT格式不需要特殊初始化
	}

	om.isInitialized = true
	return nil
}

// AddResult 添加搜索结果
func (om *OutputManager) AddResult(result *SearchResult) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	// 缓存结果
	om.results = append(om.results, result)

	// 如果初始化了文件输出，写入文件
	if om.isInitialized {
		return om.writeToFile(result)
	}

	return nil
}

// writeToFile 写入结果到文件
func (om *OutputManager) writeToFile(result *SearchResult) error {
	switch om.outputFormat {
	case "txt":
		return om.writeTxt(result)
	case "json":
		return om.writeJson(result)
	case "csv":
		return om.writeCsv(result)
	default:
		return fmt.Errorf("不支持的输出格式: %s", om.outputFormat)
	}
}

// writeTxt 以TXT格式写入
func (om *OutputManager) writeTxt(result *SearchResult) error {
	var details []string
	if result.MatchType != "" {
		details = append(details, fmt.Sprintf("匹配类型=%s", result.MatchType))
	}
	if result.MatchCount > 0 {
		details = append(details, fmt.Sprintf("匹配次数=%d", result.MatchCount))
	}
	if result.Content != "" && result.Content != "[二进制文件]" {
		content := strings.ReplaceAll(result.Content, "\n", " ")
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		details = append(details, fmt.Sprintf("内容=%s", content))
	}

	txt := fmt.Sprintf("[%s] [%s] %s | 大小=%s | 修改时间=%s | 权限=%s%s\n",
		result.Time.Format("2006-01-02 15:04:05"),
		result.Type,
		result.Path,
		formatFileSize(result.Size),
		result.ModTime,
		result.Permissions,
		func() string {
			if len(details) > 0 {
				return " | " + strings.Join(details, " | ")
			}
			return ""
		}(),
	)

	_, err := om.file.WriteString(txt)
	return err
}

// writeJson 以JSON格式写入
func (om *OutputManager) writeJson(result *SearchResult) error {
	return om.jsonEncoder.Encode(result)
}

// writeCsv 以CSV格式写入
func (om *OutputManager) writeCsv(result *SearchResult) error {
	details, _ := json.Marshal(result.Details)

	record := []string{
		result.Time.Format("2006-01-02 15:04:05"),
		string(result.Type),
		result.Path,
		formatFileSize(result.Size),
		result.ModTime,
		result.Permissions,
		result.MatchType,
		fmt.Sprintf("%d", result.MatchCount),
		result.Content,
		string(details),
	}

	if err := om.csvWriter.Write(record); err != nil {
		return err
	}
	om.csvWriter.Flush()
	return om.csvWriter.Error()
}

// GetResults 获取所有缓存的结果
func (om *OutputManager) GetResults() []*SearchResult {
	om.mu.Lock()
	defer om.mu.Unlock()
	return om.results
}

// PrintResults 打印结果到终端（列对齐格式）
// 格式: [时间] [id][文件路径][大小][修改时间][权限][匹配内容]
func (om *OutputManager) PrintResults(writer io.Writer) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if len(om.results) == 0 {
		fmt.Fprintln(writer, "[*] 未找到匹配的文件")
		return
	}

	// 打印扫描统计
	stats := make(map[ResultType]int)
	for _, r := range om.results {
		stats[r.Type]++
	}

	// 打印统计信息（单行）
	var statParts []string
	for t, count := range stats {
		statParts = append(statParts, fmt.Sprintf("%s:%d", t, count))
	}
	fmt.Fprintf(writer, "[%s] [+] 找到 %d 个结果 %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		len(om.results),
		strings.Join(statParts, " "))

	// 计算动态列宽
	maxPathLen := 30 // 路径最小宽度
	maxSizeLen := 8  // 大小最小宽度
	for _, result := range om.results {
		pathLen := len(result.Path)
		if pathLen > maxPathLen && pathLen < 60 {
			maxPathLen = pathLen
		}
		sizeLen := len(formatFileSize(result.Size))
		if sizeLen > maxSizeLen {
			maxSizeLen = sizeLen
		}
	}

	// 打印详细结果（列对齐格式，无表头）
	fmt.Fprintln(writer)
	for i, result := range om.results {
		// 获取内容预览（关键字附近10-20个字符）
		preview := om.extractKeywordPreview(result.Content, result.Details)
		if preview == "" {
			preview = "-"
		}

		// 对预览内容中的关键字进行高亮
		highlightedPreview := HighlightKeyword(preview, result.Keyword)

		// 对路径中的关键字进行高亮（如果是文件名匹配）
		highlightedPath := truncateString(result.Path, maxPathLen)
		if result.MatchType == "filename" || result.MatchType == "FILENAME" {
			highlightedPath = HighlightKeyword(highlightedPath, result.Keyword)
		}

		// 列对齐输出 - 动态宽度
		fmt.Fprintf(writer, "%-4d %-*s %-*s %-20s %-10s %s\n",
			i+1,
			maxPathLen, highlightedPath,
			maxSizeLen, formatFileSize(result.Size),
			result.Time.Format("2006-01-02 15:04:05"),
			result.Permissions,
			highlightedPreview,
		)
	}
}

// extractKeywordPreview 提取关键字附近的文本预览
func (om *OutputManager) extractKeywordPreview(content string, details map[string]interface{}) string {
	if content == "" || content == "[二进制文件]" {
		return ""
	}

	// 清理内容
	content = strings.ReplaceAll(content, "\r", " ")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.TrimSpace(content)

	// 如果内容很短，直接返回
	if len(content) <= 30 {
		return fmt.Sprintf("[%s]", content)
	}

	// 尝试从details中获取匹配行信息
	if matchLines, ok := details["match_lines"].([]int); ok && len(matchLines) > 0 {
		// 有匹配行信息，尝试提取匹配位置附近的内容
		if contextLines, ok := details["context"].([]string); ok && len(contextLines) > 0 {
			// 使用第一行匹配上下文
			firstLine := strings.TrimSpace(contextLines[0])
			firstLine = strings.ReplaceAll(firstLine, "\r", " ")
			firstLine = strings.ReplaceAll(firstLine, "\n", " ")

			if len(firstLine) > 0 {
				// 提取前30个字符作为预览
				if len(firstLine) > 30 {
					return fmt.Sprintf("[%s...]", firstLine[:30])
				}
				return fmt.Sprintf("[%s]", firstLine)
			}
		}
	}

	// 默认返回前30个字符
	if len(content) > 30 {
		return fmt.Sprintf("[%s...]", content[:30])
	}
	return fmt.Sprintf("[%s]", content)
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// HighlightKeyword 高亮显示关键字（绿色）
func HighlightKeyword(text, keyword string) string {
	if keyword == "" {
		return text
	}
	// 不区分大小写的高亮
	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	var result strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerKeyword)
		if idx == -1 {
			result.WriteString(text[start:])
			break
		}
		idx += start
		// 写入关键字前的文本
		result.WriteString(text[start:idx])
		// 写入高亮的关键字
		result.WriteString(ColorGreen)
		result.WriteString(text[idx : idx+len(keyword)])
		result.WriteString(ColorReset)
		start = idx + len(keyword)
	}
	return result.String()
}

// Close 关闭输出管理器
func (om *OutputManager) Close() error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if !om.isInitialized || om.file == nil {
		return nil
	}

	if om.csvWriter != nil {
		om.csvWriter.Flush()
	}

	if err := om.file.Close(); err != nil {
		return fmt.Errorf("关闭输出文件失败: %v", err)
	}

	om.isInitialized = false
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ANSI颜色代码
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
)

// PrintBanner 打印程序艺术字横幅
func PrintBanner() {
	banner := `
    ███████╗██╗███╗   ██╗██████╗ ███████╗██████╗ 
    ██╔════╝██║████╗  ██║██╔══██╗██╔════╝██╔══██╗
    █████╗  ██║██╔██╗ ██║██║  ██║█████╗  ██████╔╝
    ██╔══╝  ██║██║╚██╗██║██║  ██║██╔══╝  ██╔══██╗
    ██║     ██║██║ ╚████║██████╔╝███████╗██║  ██║
    ╚═╝     ╚═╝╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═╝  ╚═╝
                                                  
                      v1.0.0 - File Finder Tool
`
	fmt.Println(ColorCyan + banner + ColorReset)
}

// PrintSimpleBanner 打印简化版横幅（用于帮助信息）
func PrintSimpleBanner() {
	fmt.Println(ColorCyan + "finder v1.0.0 - File Finder Tool" + ColorReset)
	fmt.Println()
}

// PrintInfo 打印信息级别日志（带时间戳，蓝色）
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [*] %s%s\n", ColorBlue, time.Now().Format("2006-01-02 15:04:05"), msg, ColorReset)
}

// PrintSuccess 打印成功级别日志（带时间戳，绿色）
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [+] %s%s\n", ColorGreen, time.Now().Format("2006-01-02 15:04:05"), msg, ColorReset)
}

// PrintWarning 打印警告级别日志（带时间戳，黄色）
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [!] %s%s\n", ColorYellow, time.Now().Format("2006-01-02 15:04:05"), msg, ColorReset)
}

// PrintError 打印错误级别日志（带时间戳，红色）
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s[%s] [-] %s%s\n", ColorRed, time.Now().Format("2006-01-02 15:04:05"), msg, ColorReset)
}

// PrintDebug 打印调试级别日志（带时间戳，洋红色）
func PrintDebug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [DEBUG] %s%s\n", ColorMagenta, time.Now().Format("2006-01-02 15:04:05"), msg, ColorReset)
}
