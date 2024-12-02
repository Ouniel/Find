package finder

import (
	"file-finder/internal/utils"
	"fmt"
	"os"
)

func FindFilesByKeyword(keyword string, config *SearchConfig) (map[string]FileInfo, error) {
	indexer := GetIndexer()

	// 如果是全局搜索或者索引不存在，先构建索引
	if config.GlobalSearch || len(indexer.fileIndices) == 0 {
		if err := indexer.BuildIndex(config.StartDir, config); err != nil {
			return nil, err
		}
	}

	return indexer.Search(keyword, config)
}

func logAndPrint(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	utils.Logger.Print(msg)
	fmt.Fprintln(os.Stderr, msg)
}
