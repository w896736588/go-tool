package plw

import (
	"dev_tool/internal/app/dtool/define"
	"testing"
)

// TestSanitizeWindowsFilename 验证 Windows 非法文件名字符会被替换
func TestSanitizeWindowsFilename(t *testing.T) {
	got := sanitizeWindowsFilename(`报表:2026/03*04?.xlsx`)
	want := `报表_2026_03_04_.xlsx`
	if got != want {
		t.Fatalf("sanitizeWindowsFilename() = %q, want %q", got, want)
	}
}

func TestGetContextByIndexHonorsOpenType(t *testing.T) {
	previousList := list
	list = []*ContextPage{
		{
			UserDataIndex: 7,
			OpenType:      define.OpenTypeWebkitSilence,
			LinkId:        `link_id_1`,
		},
	}
	t.Cleanup(func() {
		list = previousList
	})

	contextList := NewContextList(nil)

	if got := contextList.GetContextByIndex(7, define.OpenTypeWebkitChrome); got != nil {
		t.Fatalf("GetContextByIndex() reused context with mismatched open type: %+v", got)
	}

	got := contextList.GetContextByIndex(7, define.OpenTypeWebkitSilence)
	if got == nil {
		t.Fatalf("GetContextByIndex() should return context when open type matches")
	}
}
