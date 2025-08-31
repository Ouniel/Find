package utils

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
)

type ProgressBar struct {
	total      int64
	current    int64
	currentDir string
	lastDir    string
}

func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{}
}

func (p *ProgressBar) SetCurrentDir(dir string) {
	volume := filepath.VolumeName(dir)
	if volume != "" && volume != p.lastDir {
		fmt.Printf("正在构建 %s 的文件索引...\n", volume)
		p.lastDir = volume
	}
}

func (p *ProgressBar) Start() {
	fmt.Println("开始构建文件索引...")
}

func (p *ProgressBar) Increment() {
	atomic.AddInt64(&p.current, 1)
}

func (p *ProgressBar) SetTotal(total int64) {
	atomic.StoreInt64(&p.total, total)
}

func (p *ProgressBar) Stop(completed bool) {
	if completed {
		fmt.Println("索引构建完成")
	}
}
