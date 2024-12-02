package main

import (
	"bytes"
	"file-finder/internal/finder"
	"file-finder/internal/utils"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const usage = `文件查找工具 (File Finder)

使用方法:
  file-finder [选项] 

基本选项:
  -keyword string   
        查找文件名包含指定关键字的文件
        示例: -keyword flag 或 -keyword back

  -perm string
        查找具有指定权限的文件:
        r  - 读权限
        w  - 写权限
        rw - 读写权限
        示例: -perm rw -global

  -time string
        查找指定时间后修改的文件
        格式: 2006-01-02
        示例: -time "2024-03-10"

  -global
        在根目录下进行全局搜索
        注意: Linux/Unix系统可能需要root权限

索引选项:
  -rebuild-index
        重建文件索引以提高搜索速度
        首次使用或文件变动较多时建议使用

搜索范围:
  -dir string
        指定搜索的起始目录 (默认: ".")
        
  -depth int
        限制搜索的目录深度 (默认: -1, 表示不限制)

过滤选项:
  -types string
        按文件类型过滤，多个类型用逗号分隔
        示例: -types "txt,log,conf"
        
  -size int
        限制处理的文件大小（字节）
        示例: -size 1048576 (限制为1MB)
        
  -exclude string
        排除指定目录，多个目录用逗号分隔
        示例: -exclude "tmp,cache"

性能选项:
  -concurrent
        启用并发搜索 (默认: true)
        
  -workers int
        并发搜索的工作协程数 (默认: 5)

其他选项:
  -log
        是否记录日志到文件
        默认: false

常用示例:
  1. 首次使用，建立索引:
     file-finder -rebuild-index -global

  2. 全局搜索flag文件:
     file-finder -keyword flag -global

  3. 搜索最近24小时修改的文件:
     file-finder -time "2024-03-20" -global

  4. 在指定目录搜索配置文件:
     file-finder -keyword conf -dir /etc -types "conf,cfg,ini"

  5. 搜索大文件:
     file-finder -keyword backup -size 104857600

  6. 并发搜索提高性能:
     file-finder -keyword flag -concurrent -workers 10

  7. 全局搜索具有读写权限的文件:
     file-finder -perm rw -global

  8. 在指定目录搜索具有写权限的文件:
     file-finder -perm w -dir /path/to/search

注意事项:
  1. 首次使用建议先运行 -rebuild-index 立索引
  2. 索引会在30分钟后过期，需要重建
  3. 全局搜索时会遍历所有目录，包括系统目录
  4. 建议使用 -types 和 -size 选项限制搜索范围
  5. 搜索结果包含文件路径、大小、修改时间、权限和内容
`

// 获取 Windows 系统的所有驱动器
func getWindowsDrives() []string {
	var drives []string
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		path := string(drive) + ":\\"
		if _, err := os.Stat(path); err == nil {
			drives = append(drives, path)
		}
	}
	return drives
}

// 行搜索并返回结果
func executeSearchForPath(keyword *string, permType *string, timeLimit *string, config *finder.SearchConfig) (map[string]finder.FileInfo, error) {
	results := make(map[string]finder.FileInfo)

	if *keyword != "" {
		keywordResults, err := finder.FindFilesByKeyword(*keyword, config)
		if err != nil {
			return nil, fmt.Errorf("查找关键字文件出错: %v", err)
		}
		for k, v := range keywordResults {
			results[k] = v
		}
	}

	if *permType != "" {
		files, err := finder.FindFilesByPermission(*permType)
		if err != nil {
			return nil, fmt.Errorf("查找权限文件出错: %v", err)
		}
		for _, file := range files {
			if _, exists := results[file]; !exists {
				info, err := os.Stat(file)
				if err != nil {
					continue
				}
				fileInfo, err := finder.GetFileInfo(file, info, config)
				if err != nil {
					continue
				}
				results[file] = fileInfo
			}
		}
	}

	if *timeLimit != "" {
		limitTime, err := time.Parse("2006-01-02", *timeLimit)
		if err != nil {
			return nil, fmt.Errorf("时间格式错误: %v", err)
		}
		files, err := finder.FindModifiedFiles(limitTime)
		if err != nil {
			return nil, fmt.Errorf("查找修改文件出错: %v", err)
		}
		for _, file := range files {
			if _, exists := results[file]; !exists {
				info, err := os.Stat(file)
				if err != nil {
					continue
				}
				fileInfo, err := finder.GetFileInfo(file, info, config)
				if err != nil {
					continue
				}
				results[file] = fileInfo
			}
		}
	}

	return results, nil
}

// 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// 截断路径以适应显示宽度
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-(maxLen-3):]
}

func main() {
	// 首先检查是否有任何命令行参数
	if len(os.Args) == 1 || (len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help")) {
		fmt.Println(usage)
		return
	}

	config := finder.NewDefaultConfig()

	// 设置自定义帮助信息
	flag.Usage = func() {
		fmt.Println(usage)
	}

	// 基本参数
	keyword := flag.String("keyword", "", "查找文件名包含指定关键字的文件")
	timeLimit := flag.String("time", "", "查找在指定时间内修改的文件 (格式: 2006-01-02)")

	// 新增全局搜索参数
	flag.BoolVar(&config.GlobalSearch, "global", false, "是否在根目录下进行全局搜索")

	// 其他配置参数
	flag.StringVar(&config.StartDir, "dir", ".", "起始搜索目录")
	flag.IntVar(&config.MaxDepth, "depth", -1, "最大搜索深度")
	flag.BoolVar(&config.Concurrent, "concurrent", true, "是否使用并发搜索")
	flag.IntVar(&config.MaxWorkers, "workers", 5, "并发工作协程数")
	flag.Int64Var(&config.SizeLimit, "size", -1, "文件大小限制(字节)")

	// 处理文件类型和排除目录参数
	fileTypes := flag.String("types", "", "文件类型过滤(逗号分隔，如: go,txt)")
	excludeDirs := flag.String("exclude", "", "排除的目录(逗号分隔)")

	// 在 main 函数中添加参数
	rebuildIndex := flag.Bool("rebuild-index", false, "重建文件索引")

	// 添加日志参数
	enableLog := flag.Bool("log", false, "是否记录日志")

	// 修改权限参数
	permType := flag.String("perm", "", `查找具有指定权限的文件:
        r  - 读权限
        w  - 写权限
        rw - 读写权限
        示例: -perm rw`)

	// 解析参数
	flag.Parse()

	// 检查是否有任何有效的搜索参数
	if *keyword == "" && *permType == "" && *timeLimit == "" {
		fmt.Println("错误: 请至少指定一个搜索条件（-keyword、-perm 或 -time）")
		fmt.Println("\n可用的命令行选项：")
		flag.Usage()
		return
	}

	// 处理文件类型和排除目录
	if *fileTypes != "" {
		config.FileTypes = strings.Split(*fileTypes, ",")
	}
	if *excludeDirs != "" {
		config.ExcludeDirs = strings.Split(*excludeDirs, ",")
	}

	// 添加默认排除的系统目录
	config.ExcludeDirs = append(config.ExcludeDirs,
		"$Recycle.Bin", "$RECYCLE.BIN", "System Volume Information")

	// 如果开启全局搜索，设置起始目录为根目录
	if config.GlobalSearch {
		if runtime.GOOS == "windows" {
			drives := getWindowsDrives()
			if len(drives) > 0 {
				allResults := make(map[string]finder.FileInfo)
				for _, drive := range drives {
					config.StartDir = drive
					results, err := executeSearchForPath(keyword, permType, timeLimit, config)
					if err != nil {
						continue
					}
					for k, v := range results {
						allResults[k] = v
					}
				}
				if *enableLog {
					logSearchResults(allResults)
				}
				printSearchResults(allResults, *keyword)
				return
			}
		} else {
			config.StartDir = "/"
		}
	}

	// 在参数解析添加
	if *rebuildIndex {
		config.GlobalSearch = true // 重建索引时默认全局搜索
		indexer := finder.GetIndexer()
		logAndPrint("开始重建文件索引...")
		if err := indexer.BuildIndex(config.StartDir, config); err != nil {
			logAndPrint("重建索引时出错: %v", err)
			os.Exit(1)
		}
		logAndPrint("索引重建完成")
		return
	}

	results, err := executeSearchForPath(keyword, permType, timeLimit, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "搜索出错: %v\n", err)
		os.Exit(1)
	}

	if *enableLog {
		logSearchResults(results)
	}
	printSearchResults(results, *keyword)
}

// 修改 logAndPrint 函数
func logAndPrint(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, msg)
}

// 添加日志记录搜索结果的函数
func logSearchResults(results map[string]finder.FileInfo) {
	var logBuf bytes.Buffer
	logBuf.WriteString(fmt.Sprintf("\n找到 %d 个匹配文件:\n\n", len(results)))
	printResultTable(results, &logBuf)
	utils.Logger.Print(logBuf.String())
}

// 打印搜索结果
func printSearchResults(results map[string]finder.FileInfo, keyword string) {
	if len(results) == 0 {
		if keyword != "" {
			fmt.Printf("未找到包含 '%s' 的文件\n", keyword)
		} else {
			fmt.Println("未找到匹配的文件")
		}
		return
	}

	fmt.Printf("\n找到 %d 个匹配文件:\n\n", len(results))
	printResultTable(results, os.Stdout)
}

// 添加一个辅助函数来处理文本换行
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	// 尝试在路径分隔符处换行
	parts := strings.Split(text, string(os.PathSeparator))
	var lines []string
	currentLine := ""

	for i, part := range parts {
		if i > 0 {
			if len(currentLine)+len(part)+1 > width {
				lines = append(lines, currentLine)
				currentLine = part
			} else {
				if currentLine != "" {
					currentLine += string(os.PathSeparator)
				}
				currentLine += part
			}
		} else {
			currentLine = part
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func printResultTable(results map[string]finder.FileInfo, writer io.Writer) {
	// 定义固定列宽
	const (
		pathWidth    = 50 // 文件路径固定列宽
		sizeWidth    = 8  // 文件大小列宽
		timeWidth    = 19 // 修改时间列宽
		permWidth    = 11 // 权限列宽
		contentWidth = 35 // 内容列宽
	)

	// 辅助函数：计算中文字符串实占用的宽度
	getStringWidth := func(s string) int {
		width := 0
		for _, r := range s {
			if r > 0x7F { // 中文字符
				width += 2
			} else {
				width++
			}
		}
		return width
	}

	// 辅助函数：生成表头
	makeHeader := func(text string, width int) string {
		textLen := getStringWidth(text)
		totalSpaces := width - textLen
		leftSpaces := totalSpaces / 2
		rightSpaces := totalSpaces - leftSpaces

		return strings.Repeat(" ", leftSpaces) + text + strings.Repeat(" ", rightSpaces)
	}

	// 辅助函数：打印分隔线
	printLine := func(writer io.Writer, style string) {
		const (
			corner = "+"
			line   = "-"
			double = "="
		)

		char := line
		if style == "double" {
			char = double
		}

		widths := []int{pathWidth, sizeWidth, timeWidth, permWidth, contentWidth}
		var sb strings.Builder

		for i, w := range widths {
			if i == 0 {
				sb.WriteString(corner)
			}
			sb.WriteString(strings.Repeat(char, w+2))
			sb.WriteString(corner)
		}
		fmt.Fprintln(writer, sb.String())
	}

	// 打印表头
	printHeaders := func(writer io.Writer) {
		headers := []struct {
			text     string
			width    int
			padding  int  // 添加额外的左右padding控制
			rightPad bool // true表示右padding，false表示左padding
		}{
			{"文件路径", pathWidth, 0, false},
			{"文件大小", sizeWidth, 2, true}, // 向右偏移2个空格
			{"修改时间", timeWidth, 0, false},
			{"权限", permWidth, 2, false},   // 向左偏移2个空格
			{"内容", contentWidth, 8, true}, // 向右偏移8个空格
		}

		// 构建表头行
		var headerLine strings.Builder
		for _, h := range headers {
			text := h.text
			width := h.width + 4 // 基础padding

			if h.padding > 0 {
				if h.rightPad {
					// 右侧添加额外空格
					text = text + strings.Repeat(" ", h.padding)
				} else {
					// 左侧添加额外空格
					text = strings.Repeat(" ", h.padding) + text
				}
				width -= h.padding // 调整总宽度以保持对齐
			}

			headerLine.WriteString(makeHeader(text, width))
			headerLine.WriteString("  ")
		}
		fmt.Fprintln(writer, headerLine.String())
	}

	// 按文件路径排序
	sortPaths := func(results map[string]finder.FileInfo) []string {
		paths := make([]string, 0, len(results))
		for path := range results {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		return paths
	}

	// 打印文件信息
	printFileInfo := func(writer io.Writer, fileInfo finder.FileInfo) {
		content := fileInfo.Content
		if content == "[二进制文件]" {
			content = "[二进制]"
		} else if content != "" {
			lines := strings.Split(content, "\n")
			content = lines[0]
			if len(content) > contentWidth {
				content = content[:contentWidth-3] + "..."
			}
		}

		// 处理路径换行
		pathLines := wrapText(fileInfo.Path, pathWidth)
		size := formatFileSize(fileInfo.Size)
		size = fmt.Sprintf("%*s", sizeWidth+1, size)

		// 打印第一行
		fmt.Fprintf(writer, "| %-*s | %s | %-*s | %-*s | %-*s |\n",
			pathWidth, pathLines[0],
			size,
			timeWidth, fileInfo.ModTime,
			permWidth, fileInfo.Permissions,
			contentWidth, content)

		// 如果路径有多行，直接打印剩余行，不添加分隔线
		for i := 1; i < len(pathLines); i++ {
			fmt.Fprintf(writer, "| %-*s | %*s | %-*s | %-*s | %-*s |\n",
				pathWidth, pathLines[i],
				sizeWidth+1, "",
				timeWidth, "",
				permWidth, "",
				contentWidth, "")
		}
	}

	// 打印表头
	printHeaders(writer)
	printLine(writer, "single")

	// 按文件路径排序
	paths := sortPaths(results)

	// 打印文件信息
	var prevPath string
	for i, path := range paths {
		if i > 0 {
			// 只在不同文件之间使用双线
			if filepath.Dir(path) != filepath.Dir(prevPath) {
				printLine(writer, "double")
			}
		}

		fileInfo := results[path]
		printFileInfo(writer, fileInfo)
		prevPath = path
	}

	printLine(writer, "single")
}

// 辅助函数：处理内容显示
func formatContent(content string, width int) string {
	if content == "[二进制文件]" {
		return "[二进制]"
	}
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	content = lines[0]
	if len(content) > width {
		return content[:width-3] + "..."
	}
	return content
}
