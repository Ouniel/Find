package finder

import (
	"os"
	"path/filepath"
	"time"
)

func FindModifiedFiles(limitTime time.Time, config *SearchConfig) ([]string, error) {
	var results []string
	startDir := config.StartDir

	err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			return nil
		}

		// 应用配置中的搜索限制
		if shouldSkipFile(path, info, config) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 如果是目录且不包括目录，则跳过
		if info.IsDir() && !config.IncludeDir {
			return nil
		}

		// 只返回指定时间后修改的文件
		if info.ModTime().After(limitTime) {
			results = append(results, path)
		}
		return nil
	})

	return results, err
}
