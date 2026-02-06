package finder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func FindFilesWithFlag(pattern string, config *SearchConfig) (map[string]FileInfo, error) {
	results := sync.Map{}

	if config.Concurrent {
		return findFilesWithFlagConcurrent(pattern, config)
	}

	err := filepath.Walk(config.StartDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldSkipFile(path, info, config) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
			fileInfo, err := GetFileInfo(path, info, config)
			if err != nil {
				return nil
			}
			results.Store(path, fileInfo)
		}
		return nil
	})

	return syncMapToMap(results), err
}

func findFilesWithFlagConcurrent(pattern string, config *SearchConfig) (map[string]FileInfo, error) {
	results := sync.Map{}
	paths := make(chan string, 100)
	var wg sync.WaitGroup

	for i := 0; i < config.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				info, err := os.Stat(path)
				if err != nil {
					continue
				}

				if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
					fileInfo, err := GetFileInfo(path, info, config)
					if err != nil {
						continue
					}
					results.Store(path, fileInfo)
				}
			}
		}()
	}

	err := filepath.Walk(config.StartDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldSkipFile(path, info, config) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			paths <- path
		}
		return nil
	})

	close(paths)
	wg.Wait()

	return syncMapToMap(results), err
}

func GetFileInfo(path string, info os.FileInfo, config *SearchConfig) (FileInfo, error) {
	content := ""
	if config.SizeLimit == -1 || info.Size() <= config.SizeLimit {
		data, err := os.ReadFile(path)
		if err != nil {
			return FileInfo{}, err
		}

		// 检查文件是否是二进制文件
		if isBinary(data) {
			content = "[二进制文件]"
		} else {
			// 尝试以 UTF-8 解码
			content = string(data)
			if !utf8.Valid(data) {
				// 如果不是有效的 UTF-8，尝试其他编码
				content = tryDecode(data)
			}
		}
	}

	return FileInfo{
		Path:        path,
		Size:        info.Size(),
		ModTime:     info.ModTime().Format("2006-01-02 15:04:05"),
		Permissions: info.Mode().String(),
		Content:     content,
	}, nil
}

// 检查是否是二进制文件
func isBinary(data []byte) bool {
	if len(data) > 1024 {
		data = data[:1024] // 只检查前1024字节
	}

	for _, b := range data {
		if b == 0 {
			return true
		}
	}

	return false
}

// 尝试不同的编码
func tryDecode(data []byte) string {
	// 尝试 GBK
	if reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder()); reader != nil {
		if decoded, err := io.ReadAll(reader); err == nil {
			return string(decoded)
		}
	}

	// 尝试 GB18030
	if reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GB18030.NewDecoder()); reader != nil {
		if decoded, err := io.ReadAll(reader); err == nil {
			return string(decoded)
		}
	}

	// 如果都失败了，返回十六进制表示
	return fmt.Sprintf("[无法解码的内容，前64字节: %x]", data[:min(64, len(data))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func shouldSkipFile(path string, info os.FileInfo, config *SearchConfig) bool {
	// 检查排除目录
	for _, excludeDir := range config.ExcludeDirs {
		if strings.Contains(path, excludeDir) {
			return true
		}
	}

	// 检查文件类型
	if len(config.FileTypes) > 0 && !info.IsDir() {
		ext := strings.ToLower(filepath.Ext(path))
		found := false
		for _, allowedType := range config.FileTypes {
			if ext == "."+allowedType {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// 检查最大深度
	if config.MaxDepth > 0 {
		depth := strings.Count(path, string(os.PathSeparator))
		if depth > config.MaxDepth {
			return true
		}
	}

	return false
}

func syncMapToMap(syncMap sync.Map) map[string]FileInfo {
	result := make(map[string]FileInfo)
	syncMap.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(FileInfo)
		return true
	})
	return result
}
