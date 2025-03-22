package finder

import (
	"os"
	"path/filepath"
)

const (
	READ_PERM  = 0444 // 读权限掩码
	WRITE_PERM = 0222 // 写权限掩码
)

// 更新函数以使用Config并支持不同的权限检查
func FindFilesByPermission(permType string, config *SearchConfig) ([]string, error) {
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

		mode := info.Mode().Perm()
		switch permType {
		case "r":
			// 检查读权限 - 用户、组或其他人可读
			if mode&0444 != 0 {
				results = append(results, path)
			}
		case "w":
			// 检查写权限 - 用户、组或其他人可写
			if mode&0222 != 0 {
				results = append(results, path)
			}
		case "rw":
			// 检查读写权限 - 用户、组或其他人可读写
			if mode&0444 != 0 && mode&0222 != 0 {
				results = append(results, path)
			}
		}
		return nil
	})

	return results, err
}
