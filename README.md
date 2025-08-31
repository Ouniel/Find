# 🔍 FindFlag 文件搜索工具

<p align="center">
  <img alt="Go版本" src="https://img.shields.io/badge/Go-1.20%2B-blue">
  <img alt="多平台支持" src="https://img.shields.io/badge/平台-Windows%2FLinux-green">
  <img alt="开源协议" src="https://img.shields.io/badge/许可-MIT-orange">
</p>

> 强大的文件搜索工具，支持文件名和内容全文搜索，快速定位本地文件系统中的目标文件

FindFlag 是一个高效的命令行文件搜索工具，专为开发者和系统管理员设计。支持文件名搜索和文件内容全文搜索，提供灵活的搜索条件和实时预览功能，让文件查找变得简单直观。

<br/>

## ✨ 核心功能

- **🔍 智能搜索**：支持文件名和内容搜索，包括无扩展名文件
- **⚡ 高性能**：使用Boyer-Moore算法和并发技术提升搜索速度
- **🎯 多种搜索模式**：filename（仅文件名）、content（仅内容）、both（文件名+内容）
- **📝 上下文预览**：显示匹配内容的上下文行，便于快速定位
- **🔤 编码支持**：自动检测和转换多种文本编码（UTF-8、GBK、GB18030等）
- **📊 详细信息**：显示文件大小、修改时间、权限和匹配信息
- **🌐 跨平台支持**：完美兼容Windows/Linux/macOS系统
- **🚀 性能优化**：智能文件类型检测、大文件处理优化、并发搜索

<br/>

## 🚀 快速开始

### 安装步骤
```bash
# 克隆仓库
git clone https://github.com/your-repo/FindFlag.git

# 进入项目目录
cd FindFlag

# 编译项目
go build -o file-finder

# 验证安装
./file-finder -h
```

### 基础使用
```bash
# 搜索文件名包含"flag"的文件
./file-finder -keyword flag

# 搜索文件内容包含"flag"的文件
./file-finder -keyword flag -mode content

# 同时搜索文件名和内容
./file-finder -keyword flag -mode both

# 全局搜索（所有驱动器）
./file-finder -keyword flag -mode both -global
```

<br/>

## 🛠️ 参数详解

### 基本搜索参数
| 参数 | 说明 | 示例 |
|------|------|------|
| `-keyword` | 搜索关键字 | `-keyword flag` |
| `-mode` | 搜索模式：filename/content/both | `-mode content` |
| `-time` | 修改时间过滤 (YYYY-MM-DD) | `-time "2024-01-01"` |
| `-perm` | 文件权限过滤 (r/w/rw) | `-perm rw` |

### 内容搜索参数
| 参数 | 说明 | 示例 |
|------|------|------|
| `-content` | 启用内容搜索 | `-content` |
| `-context` | 上下文行数 | `-context 3` |
| `-max-content-size` | 最大搜索文件大小(字节) | `-max-content-size 1048576` |
| `-case-sensitive` | 大小写敏感 | `-case-sensitive` |

### 过滤和范围参数
| 参数 | 说明 | 示例 |
|------|------|------|
| `-dir` | 指定搜索目录 | `-dir "/var/log"` |
| `-global` | 全局搜索 | `-global` |
| `-types` | 文件类型过滤 | `-types "txt,log,conf"` |
| `-size` | 文件大小限制 | `-size 1048576` |
| `-exclude` | 排除目录 | `-exclude "tmp,cache"` |
| `-depth` | 搜索深度 | `-depth 3` |

### 性能参数
| 参数 | 说明 | 示例 |
|------|------|------|
| `-concurrent` | 启用并发搜索 | `-concurrent` |
| `-workers` | 并发工作协程数 | `-workers 8` |
| `-rebuild-index` | 重建文件索引 | `-rebuild-index` |
| `-log` | 启用日志记录 | `-log` |

<br/>

## 📊 使用示例

### 1. 基础搜索示例
```bash
# 搜索文件名包含"config"的文件
./file-finder -keyword config

# 搜索文件内容包含"password"的文件
./file-finder -keyword password -mode content

# 同时搜索文件名和内容包含"database"的文件
./file-finder -keyword database -mode both
```

### 2. 高级搜索示例
```bash
# 搜索配置文件中包含"database"的内容，显示3行上下文
./file-finder -keyword database -mode content -types "conf,cfg,ini" -context 3

# 区分大小写搜索"Flag"
./file-finder -keyword Flag -case-sensitive -mode both

# 限制内容搜索文件大小为1MB
./file-finder -keyword error -mode content -max-content-size 1048576

# 在指定目录搜索日志文件内容
./file-finder -keyword "ERROR" -dir "/var/log" -types "log" -mode content
```

### 3. 性能优化示例
```bash
# 使用8个工作协程进行并发搜索
./file-finder -keyword flag -mode both -workers 8

# 首次使用建立索引
./file-finder -rebuild-index -global

# 排除临时目录，提高搜索效率
./file-finder -keyword config -exclude "tmp,cache,node_modules"
```

<br/>

## 📋 输出示例

```bash
找到 3 个匹配文件:

                       文件路径                         文件大小           修改时间               权限               内容
+----------------------------------------------------+----------+---------------------+-------------+-------------------------------------+
| config/database.conf                               |    1.2 KB | 2024-03-20 15:37:01 | -rw-r--r--  | [内容:2处] host = localhost            |
+====================================================+==========+=====================+=============+=====================================+
| logs/app.log                                       |    5.4 MB | 2024-03-21 09:15:22 | -rw-rw-rw-  | [内容:15处] [ERROR] Database connection |
+====================================================+==========+=====================+=============+=====================================+
| scripts/backup_flag.sh                             |      856 B | 2024-03-21 08:30:15 | -rwxr-xr-x  | [文件名+内容:3处] #!/bin/bash          |
+----------------------------------------------------+----------+---------------------+-------------+-------------------------------------+
```

**输出说明：**
- `[内容:2处]` - 表示在文件内容中找到2处匹配
- `[文件名]` - 表示仅在文件名中匹配
- `[文件名+内容:3处]` - 表示文件名和内容都匹配，内容中有3处匹配

<br/>

## 🔧 技术特性

### 搜索算法
- **Boyer-Moore算法**：高效的字符串搜索算法，特别适合长文本搜索
- **并发处理**：使用Goroutines和Channel实现并发文件处理
- **智能编码检测**：自动检测UTF-8、GBK、GB18030等编码格式

### 文件类型支持
- **文本文件**：txt, log, conf, cfg, ini, json, xml, yaml, yml, md
- **代码文件**：go, py, js, html, css, sql, sh, bat, ps1
- **无扩展名文件**：自动检测文件内容类型，支持如tesdrf等无扩展名文件

### 性能优化
- **流式读取**：大文件分块处理，避免内存溢出
- **智能过滤**：提前过滤二进制文件和系统文件
- **索引缓存**：文件索引缓存，提高重复搜索效率
- **并发控制**：可配置工作协程数，平衡性能和资源占用

<br/>

## 📌 注意事项

### 搜索模式说明
- **filename模式**：仅搜索文件名，速度最快
- **content模式**：仅搜索文件内容，适合已知文件类型的情况
- **both模式**：同时搜索文件名和内容，最全面但速度较慢

### 性能建议
1. **首次使用**：建议先运行 `-rebuild-index -global` 建立索引
2. **大范围搜索**：使用 `-types` 参数限制文件类型
3. **内容搜索**：设置合理的 `-max-content-size` 避免处理过大文件
4. **并发优化**：根据硬件配置调整 `-workers` 参数

### 权限要求
- **Windows**：普通用户权限即可
- **Linux/macOS**：全局搜索可能需要sudo权限
- **网络驱动器**：可能需要额外的网络权限

<br/>

## ⚠️ 免责声明

**使用本工具前请务必阅读并同意以下条款：**

1. **合法用途**：仅限搜索您拥有合法访问权限的文件和目录
2. **隐私尊重**：不得用于侵犯他人隐私或获取未授权数据
3. **系统安全**：避免在生产环境关键路径执行深度扫描
4. **性能影响**：大范围内容搜索可能影响系统性能
5. **风险自担**：因使用本工具导致的任何问题由用户自行承担

<br/>

## 🤝 贡献指南

欢迎提交Issue和Pull Request来改进这个项目！

### 开发环境
- Go 1.20+
- 支持的操作系统：Windows 10+, Linux, macOS

### 项目结构
```
FindFlag/
├── main.go                    # 主程序入口
├── internal/
│   ├── finder/               # 核心搜索逻辑
│   │   ├── config.go        # 配置结构
│   │   ├── keyword_finder.go # 关键字搜索
│   │   ├── indexer.go       # 文件索引
│   │   └── ...
│   ├── parser/              # 文件解析器
│   │   └── text_parser.go   # 文本文件解析
│   ├── search/              # 搜索算法
│   │   └── boyer_moore.go   # Boyer-Moore算法
│   └── utils/               # 工具函数
└── README.md
```

---

**高效定位每一份文件** - 让搜索不再成为瓶颈 🚀