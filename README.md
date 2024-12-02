# File Finder

File Finder 是一个强大的文件搜索工具，支持多种搜索条件，可以快速在本地文件系统中查找文件。

## 功能特点

- 支持按文件名关键字搜索
- 支持按文件权限搜索（读/写权限）
- 支持按修改时间搜索
- 支持全局搜索或指定目录搜索
- 支持文件类型过滤
- 支持文件大小限制
- 支持排除指定目录
- 支持并发搜索
- 美观的表格式输出
- 可选的日志记录功能

## 安装
```bash
克隆仓库
git clone https://github.com/yourusername/file-finder.git
进入项目目录
cd file-finder
编译
go build
运行
./file-finder [选项]
```
```bash
在当前目录搜索
file-finder -keyword flag
全局搜索
file-finder -keyword flag -global
```
```bash
搜索可读文件
file-finder -perm r
搜索可写文件
file-finder -perm w
搜索读写文件
file-finder -perm rw
```
``` bash
搜索特定日期后修改的文件
file-finder -time "2024-03-20"
```
``` bash
指定搜索目录
file-finder -keyword flag -dir /path/to/search
限制搜索深度
file-finder -keyword flag -depth 3
```
``` bash
按文件类型过滤
file-finder -keyword flag -types "txt,log,conf"
限制文件大小
file-finder -keyword flag -size 1048576 # 1MB
排除目录
file-finder -keyword flag -exclude "tmp,cache"
```
``` bash
调整并发数
file-finder -keyword flag -workers 8
```
```bash
启用日志
file-finder -keyword flag -log
找到 3 个匹配文件:
文件路径 文件大小 修改时间 权限 内容
+--------------------------------------------------+----------+---------------------+-----------+-----------------------------------+
| D:\Projects\test\flag.txt | 485 B | 2024-03-20 15:37:01 | -r--r--r- | # Test flag content |
+==================================================+==========+=====================+===========+===================================+
| D:\Projects\test\flag.exe | 2.0KB | 2024-03-20 14:20:15 | -rw-rw-rw | [二进制] |
+--------------------------------------------------+----------+---------------------+-----------+-----------------------------------+
```

## 注意事项

1. 全局搜索
   - Windows：会搜索所有可用驱动器
   - Linux/Unix：从根目录开始搜索，可能需要 root 权限

2. 性能优化
   - 使用 -types 限制文件类型
   - 使用 -size 限制文件大小
   - 使用 -depth 限制搜索深度
   - 适当调整 -workers 参数

3. 默认排除的系统目录
   - $Recycle.Bin
   - System Volume Information
   - 其他系统目录

4. 日志文件
   - 位置：程序运行目录下的 file_finder.log
   - 使用 -log 参数启用日志记录

## 开发环境

- Go 1.20 或更高版本
- 支持 Windows/Linux/macOS

