package finder

type FileInfo struct {
	Path        string
	Size        int64
	ModTime     string
	Permissions string
	Content     string
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
	}
}
