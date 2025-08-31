package search

import (
	"strings"
)

// BoyerMoore Boyer-Moore字符串搜索算法实现
type BoyerMoore struct {
	pattern       string
	caseSensitive bool
	badCharTable  map[rune]int
}

// NewBoyerMoore 创建新的Boyer-Moore搜索器
func NewBoyerMoore(pattern string, caseSensitive bool) *BoyerMoore {
	bm := &BoyerMoore{
		pattern:       pattern,
		caseSensitive: caseSensitive,
		badCharTable:  make(map[rune]int),
	}

	if !caseSensitive {
		bm.pattern = strings.ToLower(pattern)
	}

	bm.buildBadCharTable()
	return bm
}

// buildBadCharTable 构建坏字符表
func (bm *BoyerMoore) buildBadCharTable() {
	patternRunes := []rune(bm.pattern)
	patternLen := len(patternRunes)

	// 初始化所有字符的跳跃距离为模式长度
	for i, char := range patternRunes {
		bm.badCharTable[char] = patternLen - i - 1
	}
}

// Search 在文本中搜索模式，返回所有匹配位置
func (bm *BoyerMoore) Search(text string) []int {
	var matches []int

	searchText := text
	if !bm.caseSensitive {
		searchText = strings.ToLower(text)
	}

	textRunes := []rune(searchText)
	patternRunes := []rune(bm.pattern)
	textLen := len(textRunes)
	patternLen := len(patternRunes)

	if patternLen == 0 || textLen < patternLen {
		return matches
	}

	i := patternLen - 1 // 从模式的最后一个字符开始

	for i < textLen {
		j := patternLen - 1 // 模式的最后一个字符
		k := i              // 文本的当前位置

		// 从右到左比较
		for j >= 0 && textRunes[k] == patternRunes[j] {
			j--
			k--
		}

		if j < 0 {
			// 找到匹配
			matches = append(matches, k+1)
			i += patternLen
		} else {
			// 使用坏字符规则计算跳跃距离
			badCharShift := bm.getBadCharShift(textRunes[k])
			i += max(1, badCharShift)
		}
	}

	return matches
}

// getBadCharShift 获取坏字符的跳跃距离
func (bm *BoyerMoore) getBadCharShift(char rune) int {
	if shift, exists := bm.badCharTable[char]; exists {
		return shift
	}
	return len([]rune(bm.pattern))
}

// SearchLines 在多行文本中搜索，返回匹配的行信息
func (bm *BoyerMoore) SearchLines(lines []string) []LineMatch {
	var matches []LineMatch

	for lineNum, line := range lines {
		positions := bm.Search(line)
		if len(positions) > 0 {
			matches = append(matches, LineMatch{
				LineNumber: lineNum + 1,
				Line:       line,
				Positions:  positions,
				Count:      len(positions),
			})
		}
	}

	return matches
}

// LineMatch 行匹配结果
type LineMatch struct {
	LineNumber int
	Line       string
	Positions  []int
	Count      int
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ContextSearch 带上下文的搜索
type ContextSearch struct {
	searcher     *BoyerMoore
	contextLines int
}

// NewContextSearch 创建带上下文的搜索器
func NewContextSearch(pattern string, caseSensitive bool, contextLines int) *ContextSearch {
	return &ContextSearch{
		searcher:     NewBoyerMoore(pattern, caseSensitive),
		contextLines: contextLines,
	}
}

// SearchWithContext 搜索并返回带上下文的结果
func (cs *ContextSearch) SearchWithContext(lines []string) []ContextMatch {
	matches := cs.searcher.SearchLines(lines)
	var contextMatches []ContextMatch

	for _, match := range matches {
		contextMatch := ContextMatch{
			LineMatch: match,
			Context:   cs.getContext(lines, match.LineNumber-1),
		}
		contextMatches = append(contextMatches, contextMatch)
	}

	return contextMatches
}

// getContext 获取指定行的上下文
func (cs *ContextSearch) getContext(lines []string, lineIndex int) []string {
	start := max(0, lineIndex-cs.contextLines)
	end := min(len(lines), lineIndex+cs.contextLines+1)

	context := make([]string, end-start)
	copy(context, lines[start:end])

	return context
}

// ContextMatch 带上下文的匹配结果
type ContextMatch struct {
	LineMatch LineMatch
	Context   []string
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
