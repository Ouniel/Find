package finder

import (
	"file-finder/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileIndex struct {
	Path        string
	Name        string
	Size        int64
	ModTime     time.Time
	IsDir       bool
	Permissions os.FileMode
}

type Indexer struct {
	mu          sync.RWMutex
	fileIndices map[string]FileIndex
	nameIndices map[string][]string
	lastUpdate  time.Time
}

var (
	defaultIndexer *Indexer
	once           sync.Once
)

func GetIndexer() *Indexer {
	once.Do(func() {
		defaultIndexer = &Indexer{
			fileIndices: make(map[string]FileIndex),
			nameIndices: make(map[string][]string),
		}
	})
	return defaultIndexer
}

func (idx *Indexer) BuildIndex(startDir string, config *SearchConfig) error {
	// 创建临时映射以存储结果
	tempFileIndices := make(map[string]FileIndex)
	tempNameIndices := make(map[string][]string)

	// 创建进度条
	progress := utils.NewProgressBar(50)
	progress.Start()

	// 使用适度的缓冲区大小
	filesChan := make(chan FileIndex, 5000)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// 使用固定的工作协程数
	const workerCount = 4 // 使用固定的4个工作协程
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 每个工作协程使用本地缓存
			localCache := make(map[string]FileIndex, 1000)

			for file := range filesChan {
				localCache[file.Path] = file

				// 当本地缓存达到一定大小时批量提交
				if len(localCache) >= 1000 {
					idx.mu.Lock()
					for k, v := range localCache {
						tempFileIndices[k] = v
						tempNameIndices[v.Name] = append(tempNameIndices[v.Name], v.Path)
					}
					idx.mu.Unlock()

					// 清空本地缓存
					localCache = make(map[string]FileIndex, 1000)
				}
				progress.Increment()
			}

			// 提交剩余的本地缓存
			if len(localCache) > 0 {
				idx.mu.Lock()
				for k, v := range localCache {
					tempFileIndices[k] = v
					tempNameIndices[v.Name] = append(tempNameIndices[v.Name], v.Path)
				}
				idx.mu.Unlock()
			}
		}()
	}

	// 启动一个协程来遍历文件系统
	go func() {
		err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsPermission(err) {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				// 更新当前正在处理的目录
				progress.SetCurrentDir(path)
			}

			if shouldSkipFile(path, info, config) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 使用非阻塞的方式发送文件信息
			select {
			case filesChan <- FileIndex{
				Path: path,
				Name: info.Name(),
				Size: info.Size(),

				ModTime:     info.ModTime(),
				IsDir:       info.IsDir(),
				Permissions: info.Mode(),
			}:
			case <-time.After(10 * time.Millisecond):
				// 如果通道已满，短暂等待后继续
				return nil
			}

			return nil
		})

		close(filesChan)
		if err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	// 等待所有工作协程完成
	wg.Wait()

	// 检查是否有错误发生
	if err := <-errChan; err != nil {
		progress.Stop(false)
		return err
	}

	// 更新索引
	idx.mu.Lock()
	idx.fileIndices = tempFileIndices
	idx.nameIndices = tempNameIndices
	idx.lastUpdate = time.Now()
	idx.mu.Unlock()

	progress.Stop(true)
	return nil
}

func (idx *Indexer) addToIndex(file FileIndex) {
	idx.fileIndices[file.Path] = file
	idx.nameIndices[file.Name] = append(idx.nameIndices[file.Name], file.Path)
}

func (idx *Indexer) Search(keyword string, config *SearchConfig) (map[string]FileInfo, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	results := make(map[string]FileInfo)

	// 如果索引太旧，建议重建
	if time.Since(idx.lastUpdate) > 30*time.Minute {
		utils.Logger.Print("[索引] 索引已过期，建议重建")
	}

	// 使用名称索引快速查找
	for name, paths := range idx.nameIndices {
		if !strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) {
			continue
		}

		for _, path := range paths {
			// 使用索引中的信息而不是重新获取
			if fileIndex, ok := idx.fileIndices[path]; ok && !fileIndex.IsDir {
				// 验证文件是否仍然存在
				info, err := os.Stat(path)
				if err != nil {
					continue
				}

				// 检查文件是否被修改
				if info.ModTime() != fileIndex.ModTime {
					// 如果文件被修改，更新索引
					fileIndex = FileIndex{
						Path:        path,
						Name:        info.Name(),
						Size:        info.Size(),
						ModTime:     info.ModTime(),
						IsDir:       info.IsDir(),
						Permissions: info.Mode(),
					}
					idx.fileIndices[path] = fileIndex
				}

				fileInfo, err := GetFileInfo(path, info, config)
				if err != nil {
					continue
				}
				results[path] = fileInfo
			}
		}
	}

	return results, nil
}
