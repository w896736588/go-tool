package common

import (
	"reflect"
	"testing"
)

// TestMemoryFragmentSearchTokens 验证多关键词搜索会按空格拆分并去重。
func TestMemoryFragmentSearchTokens(t *testing.T) {
	handler := &CSqlite{}
	result := handler.memoryFragmentSearchTokens(" Git   冲突  git ")
	expect := []string{"git", "冲突"}
	if !reflect.DeepEqual(result, expect) {
		t.Fatalf("memoryFragmentSearchTokens() = %v, want %v", result, expect)
	}
}

// TestMemoryFragmentSearchScoreMultiKeyword 验证多关键词需要同时命中才算匹配。
func TestMemoryFragmentSearchScoreMultiKeyword(t *testing.T) {
	handler := &CSqlite{}
	query := handler.memoryFragmentNormalizeSearchQuery(" git   冲突 ")
	tokens := handler.memoryFragmentSearchTokens(query)

	matched, score := handler.memoryFragmentSearchScore(
		"keyword",
		query,
		tokens,
		"这是处理冲突时常用的 git rebase 记录",
		[]string{"版本控制"},
		"git 操作备忘",
	)
	if !matched {
		t.Fatalf("memoryFragmentSearchScore() matched = false, want true")
	}
	if score <= 0 {
		t.Fatalf("memoryFragmentSearchScore() score = %d, want > 0", score)
	}

	matched, _ = handler.memoryFragmentSearchScore(
		"keyword",
		query,
		tokens,
		"这里只有 git，没有另一个关键词",
		[]string{"版本控制"},
		"git 操作备忘",
	)
	if matched {
		t.Fatalf("memoryFragmentSearchScore() matched = true, want false")
	}
}
