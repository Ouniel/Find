package parser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// TextParser 文本解析器
type TextParser struct {
	maxSize int64
}

// NewTextParser 创建新的文本解析器
func NewTextParser(maxSize int64) *TextParser {
	return &TextParser{
		maxSize: maxSize,
	}
}

// ParseFile 解析文件内容
func (p *TextParser) ParseFile(filePath string) (string, error) {
	// 检查文件大小
	info, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	if p.maxSize > 0 && info.Size() > p.maxSize {
		return "", fmt.Errorf("文件大小超过限制: %d bytes", p.maxSize)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 检查是否为二进制文件
	if p.isBinaryFile(file) {
		return "[二进制文件]", nil
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 读取文件内容
	content, err := p.readFileContent(file)
	if err != nil {
		return "", err
	}

	return content, nil
}

// isBinaryFile 检查是否为二进制文件
func (p *TextParser) isBinaryFile(file *os.File) bool {
	// 读取前512字节进行检测
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return true
	}

	// 检查是否包含null字节或大量非打印字符
	nullCount := 0
	nonPrintCount := 0

	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			nullCount++
		}
		if buffer[i] < 32 && buffer[i] != 9 && buffer[i] != 10 && buffer[i] != 13 {
			nonPrintCount++
		}
	}

	// 如果包含null字节或非打印字符过多，认为是二进制文件
	if nullCount > 0 || float64(nonPrintCount)/float64(n) > 0.3 {
		return true
	}

	return false
}

// readFileContent 读取文件内容并处理编码
func (p *TextParser) readFileContent(file *os.File) (string, error) {
	// 读取所有内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// 检测和转换编码
	text, err := p.detectAndConvertEncoding(content)
	if err != nil {
		return "", err
	}

	return text, nil
}

// detectAndConvertEncoding 检测并转换编码
func (p *TextParser) detectAndConvertEncoding(data []byte) (string, error) {
	// 首先尝试UTF-8
	if utf8.Valid(data) {
		return string(data), nil
	}

	// 尝试常见编码
	encodings := []encoding.Encoding{
		unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
		unicode.UTF16(unicode.BigEndian, unicode.UseBOM),
		charmap.Windows1252,
		charmap.ISO8859_1,
	}

	for _, enc := range encodings {
		decoder := enc.NewDecoder()
		result, err := decoder.Bytes(data)
		if err == nil && utf8.Valid(result) {
			return string(result), nil
		}
	}

	// 如果都失败了，尝试忽略无效字符
	reader := transform.NewReader(bytes.NewReader(data),
		transform.Chain(unicode.UTF8.NewDecoder(), transform.RemoveFunc(func(r rune) bool {
			return r == utf8.RuneError
		})))

	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// IsTextFile 判断文件是否为文本文件
func (p *TextParser) IsTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// 常见文本文件扩展名
	textExts := map[string]bool{
		".txt":  true,
		".log":  true,
		".conf": true,
		".cfg":  true,
		".ini":  true,
		".json": true,
		".xml":  true,
		".yaml": true,
		".yml":  true,
		".md":   true,
		".go":   true,
		".py":   true,
		".js":   true,
		".html": true,
		".css":  true,
		".sql":  true,
		".sh":   true,
		".bat":  true,
		".ps1":  true,
	}

	// 如果有扩展名且在列表中
	if ext != "" {
		return textExts[ext]
	}

	// 无扩展名文件，需要进一步检测
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	return !p.isBinaryFile(file)
}

// GetFileLines 获取文件的所有行
func (p *TextParser) GetFileLines(filePath string) ([]string, error) {
	content, err := p.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	if content == "[二进制文件]" {
		return nil, fmt.Errorf("二进制文件无法按行读取")
	}

	lines := strings.Split(content, "\n")
	return lines, nil
}

// SearchInContent 在内容中搜索关键词
func (p *TextParser) SearchInContent(content, keyword string, caseSensitive bool) []MatchResult {
	var results []MatchResult

	searchContent := content
	searchKeyword := keyword

	if !caseSensitive {
		searchContent = strings.ToLower(content)
		searchKeyword = strings.ToLower(keyword)
	}

	lines := strings.Split(content, "\n")
	searchLines := strings.Split(searchContent, "\n")

	for i, line := range searchLines {
		if strings.Contains(line, searchKeyword) {
			results = append(results, MatchResult{
				LineNumber:     i + 1,
				LineContent:    lines[i],
				MatchPositions: p.findMatchPositions(line, searchKeyword),
			})
		}
	}

	return results
}

// MatchResult 匹配结果
type MatchResult struct {
	LineNumber     int
	LineContent    string
	MatchPositions []int
}

// findMatchPositions 查找匹配位置
func (p *TextParser) findMatchPositions(line, keyword string) []int {
	var positions []int
	start := 0

	for {
		pos := strings.Index(line[start:], keyword)
		if pos == -1 {
			break
		}
		positions = append(positions, start+pos)
		start += pos + len(keyword)
	}

	return positions
}
