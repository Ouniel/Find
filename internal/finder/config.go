package finder

type FileInfo struct {
	Path        string
	Size        int64
	ModTime     string
	Permissions string
	Content     string
	// 新增匹配信息
	MatchType  string   // 匹配类型：filename, content, both
	MatchLines []int    // 匹配的行号
	MatchCount int      // 匹配次数
	Context    []string // 上下文内容
}

type SearchConfig struct {
	StartDir     string
	MaxDepth     int
	Concurrent   bool
	MaxWorkers   int
	IncludeDir   bool
	SizeLimit    int64
	FileTypes    []string
	ExcludeDirs  []string
	GlobalSearch bool
	// 新增内容搜索相关配置
	ContentSearch  bool   // 是否启用内容搜索
	SearchMode     string // 搜索模式：filename, content, both
	ContextLines   int    // 上下文行数
	MaxContentSize int64  // 最大内容搜索文件大小
	CaseSensitive  bool   // 是否区分大小写
}

func NewDefaultConfig() *SearchConfig {
	return &SearchConfig{
		StartDir:     ".",
		MaxDepth:     -1,
		Concurrent:   true,
		MaxWorkers:   5,
		IncludeDir:   false,
		SizeLimit:    -1,
		FileTypes:    []string{},
		ExcludeDirs:  []string{".git", "node_modules"},
		GlobalSearch: false,
		// 内容搜索默认配置
		ContentSearch:  false,
		SearchMode:     "filename",       // 默认只搜索文件名
		ContextLines:   2,                // 默认显示2行上下文
		MaxContentSize: 10 * 1024 * 1024, // 默认最大10MB文件进行内容搜索
		CaseSensitive:  false,            // 默认不区分大小写
	}
}
