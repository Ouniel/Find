package finder

import (
	"file-finder/internal/parser"
	"file-finder/internal/search"
	"file-finder/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func FindFilesByKeyword(keyword string, config *SearchConfig) (map[string]FileInfo, error) {
	indexer := GetIndexer()

	// 如果是全局搜索或者索引不存在，先构建索引
	if config.GlobalSearch || len(indexer.fileIndices) == 0 {
		if err := indexer.BuildIndex(config.StartDir, config); err != nil {
			return nil, err
		}
	}

	// 根据搜索模式执行不同的搜索策略
	switch config.SearchMode {
	case "content":
		return findByContentOnly(keyword, config)
	case "both":
		return findByBoth(keyword, config)
	default: // "filename"
		return indexer.Search(keyword, config)
	}
}

// findByContentOnly 仅搜索文件内容
func findByContentOnly(keyword string, config *SearchConfig) (map[string]FileInfo, error) {
	results := make(map[string]FileInfo)
	textParser := parser.NewTextParser(config.MaxContentSize)

	err := filepath.Walk(config.StartDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否需要排除此目录
			if shouldExcludeDir(path, config.ExcludeDirs) {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件大小限制
		if config.SizeLimit > 0 && info.Size() > config.SizeLimit {
			return nil
		}

		// 检查文件类型
		if len(config.FileTypes) > 0 && !isAllowedFileType(path, config.FileTypes) {
			return nil
		}

		// 检查是否为文本文件
		if !textParser.IsTextFile(path) {
			return nil
		}

		// 搜索文件内容
		fileInfo, found := searchFileContent(path, keyword, config, textParser)
		if found {
			results[path] = fileInfo
		}

		return nil
	})

	return results, err
}

// findByBoth 同时搜索文件名和内容
func findByBoth(keyword string, config *SearchConfig) (map[string]FileInfo, error) {
	// 先获取文件名匹配的结果
	indexer := GetIndexer()
	filenameResults, err := indexer.Search(keyword, config)
	if err != nil {
		return nil, err
	}

	// 再获取内容匹配的结果
	contentResults, err := findByContentOnly(keyword, config)
	if err != nil {
		return nil, err
	}

	// 合并结果
	results := make(map[string]FileInfo)

	// 添加文件名匹配的结果
	for path, info := range filenameResults {
		info.MatchType = "filename"
		results[path] = info
	}

	// 添加内容匹配的结果
	for path, info := range contentResults {
		if existing, exists := results[path]; exists {
			// 如果文件既匹配文件名又匹配内容
			existing.MatchType = "both"
			existing.MatchLines = info.MatchLines
			existing.MatchCount = info.MatchCount
			existing.Context = info.Context
			results[path] = existing
		} else {
			info.MatchType = "content"
			results[path] = info
		}
	}

	return results, nil
}

// searchFileContent 搜索单个文件的内容
func searchFileContent(filePath, keyword string, config *SearchConfig, textParser *parser.TextParser) (FileInfo, bool) {
	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		return FileInfo{}, false
	}

	// 获取文件行内容
	lines, err := textParser.GetFileLines(filePath)
	if err != nil {
		return FileInfo{}, false
	}

	// 使用Boyer-Moore算法搜索
	contextSearch := search.NewContextSearch(keyword, config.CaseSensitive, config.ContextLines)
	matches := contextSearch.SearchWithContext(lines)

	if len(matches) == 0 {
		return FileInfo{}, false
	}

	// 构建FileInfo
	fileInfo := FileInfo{
		Path:        filePath,
		Size:        info.Size(),
		ModTime:     info.ModTime().Format("2006-01-02 15:04:05"),
		Permissions: info.Mode().String(),
		Content:     "",
	}

	// 添加匹配信息
	fileInfo.MatchLines = make([]int, len(matches))
	fileInfo.Context = make([]string, 0)
	totalMatches := 0

	for i, match := range matches {
		fileInfo.MatchLines[i] = match.LineMatch.LineNumber
		totalMatches += match.LineMatch.Count

		// 添加上下文（避免重复）
		for _, contextLine := range match.Context {
			if !contains(fileInfo.Context, contextLine) {
				fileInfo.Context = append(fileInfo.Context, contextLine)
			}
		}
	}

	fileInfo.MatchCount = totalMatches
	fileInfo.MatchType = "content"

	// 设置内容预览（显示第一个匹配的上下文）
	if len(matches) > 0 && len(matches[0].Context) > 0 {
		fileInfo.Content = strings.Join(matches[0].Context, "\n")
	}

	return fileInfo, true
}

// 并发搜索文件内容
func searchFileContentConcurrent(filePaths []string, keyword string, config *SearchConfig) map[string]FileInfo {
	results := make(map[string]FileInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建工作通道
	jobs := make(chan string, len(filePaths))

	// 启动工作协程
	for i := 0; i < config.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			textParser := parser.NewTextParser(config.MaxContentSize)

			for filePath := range jobs {
				fileInfo, found := searchFileContent(filePath, keyword, config, textParser)
				if found {
					mu.Lock()
					results[filePath] = fileInfo
					mu.Unlock()
				}
			}
		}()
	}

	// 发送任务
	for _, filePath := range filePaths {
		jobs <- filePath
	}
	close(jobs)

	// 等待完成
	wg.Wait()

	return results
}

// 辅助函数
func shouldExcludeDir(dirPath string, excludeDirs []string) bool {
	dirName := filepath.Base(dirPath)
	for _, exclude := range excludeDirs {
		if dirName == exclude {
			return true
		}
	}
	return false
}

func isAllowedFileType(filePath string, allowedTypes []string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != "" {
		ext = ext[1:] // 移除点号
	}

	for _, allowedType := range allowedTypes {
		if ext == strings.ToLower(allowedType) {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func logAndPrint(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	utils.Logger.Print(msg)
	fmt.Fprintln(os.Stderr, msg)
}
