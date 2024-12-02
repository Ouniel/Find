package finder

import (
	"os"
	"path/filepath"
)

const (
	READ_PERM  = 0444 // 读权限掩码
	WRITE_PERM = 0222 // 写权限掩码
)

// 修改函数以支持不同的权限检查
func FindFilesByPermission(permType string) ([]string, error) {
	var results []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		mode := info.Mode().Perm()
		switch permType {
		case "r":
			// 检查读权限
			if mode&READ_PERM == READ_PERM {
				results = append(results, path)
			}
		case "w":
			// 检查写权限
			if mode&WRITE_PERM == WRITE_PERM {
				results = append(results, path)
			}
		case "rw":
			// 检查读写权限
			if mode&(READ_PERM|WRITE_PERM) == (READ_PERM | WRITE_PERM) {
				results = append(results, path)
			}
		}
		return nil
	})

	return results, err
}
