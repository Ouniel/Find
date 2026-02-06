package main

import (
	"file-finder/internal/finder"
	"file-finder/internal/utils"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const simpleUsage = `使用方法: finder [选项]

基本选项:
  -k, -keyword string    搜索关键字 (必填)
  -d, -dir string        搜索目录 (默认: ".")
  -g, -global            全局搜索
  -m, -mode string       搜索模式: filename/content/both (默认: filename)
  -h, -help              显示完整帮助信息

示例:
  finder -k flag -g                    # 全局搜索文件名包含flag的文件
  finder -k flag -m content -g         # 全局搜索文件内容包含flag的文件
  finder -k flag -d ./logs -m both     # 在logs目录搜索文件名和内容

使用 -h 查看完整帮助信息
`

const fullUsage = `使用方法: finder [选项]

基本选项:
  -k, -keyword string    搜索关键字
  -d, -dir string        搜索目录 (默认: ".")
  -g, -global            在根目录下进行全局搜索
  -D, -depth int         限制搜索深度 (默认: -1, 不限制)

内容搜索选项:
  -m, -mode string       搜索模式 (默认: filename)
                         filename - 仅搜索文件名
                         content  - 仅搜索文件内容
                         both     - 同时搜索文件名和内容
  -c, -context int       显示匹配内容的上下文行数 (默认: 2)
  -M, -max-content-size int   内容搜索的最大文件大小，单位字节 (默认: 10MB)
  -s, -case-sensitive    启用大小写敏感搜索

权限和时间搜索:
  -p, -perm string       按权限搜索: r/w/rw
  -t, -time string       搜索指定时间后修改的文件 (格式: 2006-01-02)

索引选项:
  -r, -rebuild-index     重建文件索引

过滤选项:
  -T, -types string      按文件类型过滤，逗号分隔 (如: go,txt,log)
  -S, -size int          限制文件大小（字节）
  -e, -exclude string    排除目录，逗号分隔

性能选项:
  -C, -concurrent        启用并发搜索 (默认: true)
  -w, -workers int       并发工作协程数 (默认: 5)

输出选项:
  -o, -output string     输出结果到指定文件
  -f, -format string     输出格式: txt/json/csv (默认: txt)
  -l, -log               记录调试日志到文件

常用示例:
  1. 首次使用，建立索引:
     finder -r -g

  2. 全局搜索文件名:
     finder -k flag -g

  3. 搜索文件内容:
     finder -k flag -m content -g

  4. 同时搜索文件名和内容:
     finder -k flag -m both -g

  5. 搜索并显示上下文:
     finder -k flag -m content -c 3 -g

  6. 区分大小写搜索:
     finder -k Flag -s -m both -g

  7. 限制内容搜索文件大小:
     finder -k flag -m content -M 1048576 -g

  8. 指定目录搜索配置文件:
     finder -k database -d /etc -T "conf,cfg,ini" -m content

  9. 搜索最近修改的文件:
     finder -t "2024-03-20" -g

  10. 搜索具有读写权限的文件:
      finder -p rw -g

  11. 保存结果到JSON:
      finder -k flag -m both -f json -o result.json

注意事项:
  1. 首次使用建议先运行 -r -g 建立索引
  2. 索引会在30分钟后过期，需要重建
  3. 全局搜索时会遍历所有目录
  4. 建议使用 -T 和 -S 选项限制搜索范围
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

// 执行搜索并返回结果
func executeSearchForPath(keyword *string, permType *string, timeLimit *string, config *finder.SearchConfig) (map[string]finder.FileInfo, error) {
	results := make(map[string]finder.FileInfo)

	if *keyword != "" {
		utils.PrintInfo("开始关键字搜索: %s", *keyword)
		keywordResults, err := finder.FindFilesByKeyword(*keyword, config)
		if err != nil {
			return nil, fmt.Errorf("查找关键字文件出错: %v", err)
		}
		for k, v := range keywordResults {
			results[k] = v
		}
		utils.PrintSuccess("关键字搜索完成，找到 %d 个结果", len(keywordResults))
	}

	if *permType != "" {
		utils.PrintInfo("开始权限搜索: %s", *permType)
		files, err := finder.FindFilesByPermission(*permType, config)
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
		utils.PrintSuccess("权限搜索完成，找到 %d 个结果", len(files))
	}

	if *timeLimit != "" {
		limitTime, err := time.Parse("2006-01-02", *timeLimit)
		if err != nil {
			return nil, fmt.Errorf("时间格式错误: %v", err)
		}
		utils.PrintInfo("开始时间搜索: %s 之后的文件", *timeLimit)
		files, err := finder.FindModifiedFiles(limitTime, config)
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
		utils.PrintSuccess("时间搜索完成，找到 %d 个结果", len(files))
	}

	return results, nil
}

// convertToSearchResults 将FileInfo转换为SearchResult
func convertToSearchResults(results map[string]finder.FileInfo, keyword string) []*utils.SearchResult {
	var searchResults []*utils.SearchResult
	for path, info := range results {
		resultType := utils.FILE_FOUND
		if info.MatchType == "content" {
			resultType = utils.CONTENT_MATCH
		}

		result := &utils.SearchResult{
			Time:        time.Now(),
			Type:        resultType,
			Path:        path,
			Size:        info.Size,
			ModTime:     info.ModTime,
			Permissions: info.Permissions,
			MatchType:   info.MatchType,
			MatchCount:  info.MatchCount,
			Content:     info.Content,
			Keyword:     keyword,
			Details: map[string]interface{}{
				"match_lines": info.MatchLines,
				"context":     info.Context,
			},
		}
		searchResults = append(searchResults, result)
	}
	return searchResults
}

func main() {
	config := finder.NewDefaultConfig()

	// 设置自定义帮助信息
	flag.Usage = func() {
		utils.PrintSimpleBanner()
		fmt.Println(simpleUsage)
	}

	// 基本参数（支持长短选项）
	var keyword string
	flag.StringVar(&keyword, "keyword", "", "搜索关键字")
	flag.StringVar(&keyword, "k", "", "搜索关键字")

	var timeLimit string
	flag.StringVar(&timeLimit, "time", "", "查找在指定时间内修改的文件 (格式: 2006-01-02)")
	flag.StringVar(&timeLimit, "t", "", "查找在指定时间内修改的文件 (格式: 2006-01-02)")

	// 内容搜索参数
	flag.StringVar(&config.SearchMode, "mode", "filename", "搜索模式: filename/content/both")
	flag.StringVar(&config.SearchMode, "m", "filename", "搜索模式: filename/content/both")
	flag.IntVar(&config.ContextLines, "context", 2, "显示匹配内容的上下文行数")
	flag.IntVar(&config.ContextLines, "c", 2, "显示匹配内容的上下文行数")
	flag.Int64Var(&config.MaxContentSize, "max-content-size", 10*1024*1024, "内容搜索的最大文件大小(字节)")
	flag.Int64Var(&config.MaxContentSize, "M", 10*1024*1024, "内容搜索的最大文件大小(字节)")
	flag.BoolVar(&config.CaseSensitive, "case-sensitive", false, "是否区分大小写")
	flag.BoolVar(&config.CaseSensitive, "s", false, "是否区分大小写")

	// 全局搜索参数
	flag.BoolVar(&config.GlobalSearch, "global", false, "是否在根目录下进行全局搜索")
	flag.BoolVar(&config.GlobalSearch, "g", false, "是否在根目录下进行全局搜索")

	// 其他配置参数
	flag.StringVar(&config.StartDir, "dir", ".", "起始搜索目录")
	flag.StringVar(&config.StartDir, "d", ".", "起始搜索目录")
	flag.IntVar(&config.MaxDepth, "depth", -1, "最大搜索深度")
	flag.IntVar(&config.MaxDepth, "D", -1, "最大搜索深度")
	flag.BoolVar(&config.Concurrent, "concurrent", true, "是否使用并发搜索")
	flag.BoolVar(&config.Concurrent, "C", true, "是否使用并发搜索")
	flag.IntVar(&config.MaxWorkers, "workers", 5, "并发工作协程数")
	flag.IntVar(&config.MaxWorkers, "w", 5, "并发工作协程数")
	flag.Int64Var(&config.SizeLimit, "size", -1, "文件大小限制(字节)")
	flag.Int64Var(&config.SizeLimit, "S", -1, "文件大小限制(字节)")

	// 处理文件类型和排除目录参数
	var fileTypes string
	flag.StringVar(&fileTypes, "types", "", "文件类型过滤(逗号分隔，如: go,txt)")
	flag.StringVar(&fileTypes, "T", "", "文件类型过滤(逗号分隔，如: go,txt)")
	var excludeDirs string
	flag.StringVar(&excludeDirs, "exclude", "", "排除的目录(逗号分隔)")
	flag.StringVar(&excludeDirs, "e", "", "排除的目录(逗号分隔)")

	// 重建索引参数
	var rebuildIndex bool
	flag.BoolVar(&rebuildIndex, "rebuild-index", false, "重建文件索引")
	flag.BoolVar(&rebuildIndex, "r", false, "重建文件索引")

	// 日志参数
	var enableLog bool
	flag.BoolVar(&enableLog, "log", false, "是否记录日志")
	flag.BoolVar(&enableLog, "l", false, "是否记录日志")

	// 输出参数
	var outputPath string
	flag.StringVar(&outputPath, "o", "", "输出结果到指定文件")
	flag.StringVar(&outputPath, "output", "", "输出结果到指定文件")
	var outputFormat string
	flag.StringVar(&outputFormat, "of", "txt", "输出格式: txt, json, csv")
	flag.StringVar(&outputFormat, "format", "txt", "输出格式: txt, json, csv")
	flag.StringVar(&outputFormat, "f", "txt", "输出格式: txt, json, csv")

	// 权限参数
	var permType string
	flag.StringVar(&permType, "perm", "", "按权限搜索: r/w/rw")
	flag.StringVar(&permType, "p", "", "按权限搜索: r/w/rw")

	// 检查是否需要显示完整帮助信息（在flag.Parse之前检查）
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" || arg == "-help" {
			utils.PrintBanner()
			fmt.Println(fullUsage)
			return
		}
	}

	// 解析参数
	flag.Parse()

	// 如果没有参数，显示简化帮助
	if len(os.Args) == 1 {
		utils.PrintSimpleBanner()
		fmt.Println(simpleUsage)
		return
	}

	// 打印艺术字横幅
	utils.PrintBanner()

	// 初始化日志系统
	if err := utils.InitLogger(enableLog); err != nil {
		utils.PrintError("初始化日志失败: %v", err)
		os.Exit(1)
	}
	defer utils.CloseLogger()

	// 初始化输出管理器
	if err := utils.InitOutputManager(outputPath, outputFormat); err != nil {
		utils.PrintError("初始化输出管理器失败: %v", err)
		os.Exit(1)
	}
	defer utils.GlobalOutputManager.Close()

	// 检查是否有任何有效的搜索参数
	if keyword == "" && permType == "" && timeLimit == "" && !rebuildIndex {
		utils.PrintError("请至少指定一个搜索条件（-k/-keyword、-p/-perm、-t/-time 或 -r/-rebuild-index）")
		fmt.Println("\n可用的命令行选项：")
		flag.Usage()
		return
	}

	// 处理文件类型和排除目录
	if fileTypes != "" {
		config.FileTypes = strings.Split(fileTypes, ",")
	}
	if excludeDirs != "" {
		config.ExcludeDirs = strings.Split(excludeDirs, ",")
	}

	// 添加默认排除的系统目录
	config.ExcludeDirs = append(config.ExcludeDirs,
		"$Recycle.Bin", "$RECYCLE.BIN", "System Volume Information")

	// 如果开启全局搜索，设置起始目录为根目录
	if config.GlobalSearch {
		if runtime.GOOS == "windows" {
			drives := getWindowsDrives()
			if len(drives) > 0 {
				utils.PrintInfo("Windows系统，发现 %d 个驱动器", len(drives))
				allResults := make(map[string]finder.FileInfo)
				for _, drive := range drives {
					utils.PrintInfo("正在搜索驱动器: %s", drive)
					config.StartDir = drive
					results, err := executeSearchForPath(&keyword, &permType, &timeLimit, config)
					if err != nil {
						utils.PrintWarning("搜索驱动器 %s 时出错: %v", drive, err)
						continue
					}
					for k, v := range results {
						allResults[k] = v
					}
				}
				// 转换并保存结果
				searchResults := convertToSearchResults(allResults, keyword)
				for _, result := range searchResults {
					utils.GlobalOutputManager.AddResult(result)
				}
				utils.GlobalOutputManager.PrintResults(os.Stdout)
				return
			}
		} else {
			config.StartDir = "/"
		}
	}

	// 在参数解析添加
	if rebuildIndex {
		config.GlobalSearch = true // 重建索引时默认全局搜索
		indexer := finder.GetIndexer()
		utils.PrintInfo("开始重建文件索引...")
		if err := indexer.BuildIndex(config.StartDir, config); err != nil {
			utils.PrintError("重建索引时出错: %v", err)
			os.Exit(1)
		}
		utils.PrintSuccess("索引重建完成")
		return
	}

	results, err := executeSearchForPath(&keyword, &permType, &timeLimit, config)
	if err != nil {
		utils.PrintError("搜索出错: %v", err)
		os.Exit(1)
	}

	// 转换并保存结果
	searchResults := convertToSearchResults(results, keyword)
	for _, result := range searchResults {
		utils.GlobalOutputManager.AddResult(result)
	}

	// 打印结果到终端
	utils.GlobalOutputManager.PrintResults(os.Stdout)

	// 如果指定了输出文件，提示用户
	if outputPath != "" {
		utils.PrintSuccess("结果已保存到: %s", outputPath)
	}
}
