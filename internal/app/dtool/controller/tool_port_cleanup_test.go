package controller

import (
	"errors"
	"testing"
)

func TestPortCleanupPlanPrefersMatchingCommandsBeforeFallback(t *testing.T) {
	items := []toolPortProcessItem{
		{PID: 1001, Command: "dtool.exe", Address: "0.0.0.0:17170"},
		{PID: 2002, Command: "nginx.exe", Address: "0.0.0.0:17170"},
		{PID: 3003, Command: "DTOOL", Address: "0.0.0.0:17171"},
	}

	preferred, fallback := buildPortCleanupPlan(items, []string{"dtool"})

	if len(preferred) != 2 {
		t.Fatalf("preferred len = %d, want 2, got=%v", len(preferred), preferred)
	}
	if preferred[0] != 1001 || preferred[1] != 3003 {
		t.Fatalf("preferred = %v, want [1001 3003]", preferred)
	}
	if len(fallback) != 1 || fallback[0] != 2002 {
		t.Fatalf("fallback = %v, want [2002]", fallback)
	}
}

func TestCleanupPortsByPreferenceKillsPreferredThenFallback(t *testing.T) {
	queryCalls := 0
	killed := make([]int, 0)

	origQuery := queryToolPortProcessesFunc
	origKill := killToolProcessFunc
	defer func() {
		queryToolPortProcessesFunc = origQuery
		killToolProcessFunc = origKill
	}()

	queryToolPortProcessesFunc = func(port int) ([]toolPortProcessItem, error) {
		queryCalls++
		switch queryCalls {
		case 1:
			return []toolPortProcessItem{
				{PID: 111, Command: "dtool.exe", Address: "0.0.0.0:17170"},
				{PID: 222, Command: "nginx.exe", Address: "0.0.0.0:17170"},
			}, nil
		case 2:
			return []toolPortProcessItem{
				{PID: 222, Command: "nginx.exe", Address: "0.0.0.0:17170"},
			}, nil
		default:
			return nil, nil
		}
	}
	killToolProcessFunc = func(pid int) error {
		killed = append(killed, pid)
		return nil
	}

	if err := CleanupPortsByPreference([]string{"17170"}, []string{"dtool"}); err != nil {
		t.Fatalf("CleanupPortsByPreference() error = %v", err)
	}

	if len(killed) != 2 {
		t.Fatalf("killed len = %d, want 2, got=%v", len(killed), killed)
	}
	if killed[0] != 111 || killed[1] != 222 {
		t.Fatalf("killed = %v, want [111 222]", killed)
	}
}

func TestCleanupPortsByPreferenceReturnsKillError(t *testing.T) {
	origQuery := queryToolPortProcessesFunc
	origKill := killToolProcessFunc
	defer func() {
		queryToolPortProcessesFunc = origQuery
		killToolProcessFunc = origKill
	}()

	queryToolPortProcessesFunc = func(port int) ([]toolPortProcessItem, error) {
		return []toolPortProcessItem{
			{PID: 111, Command: "dtool.exe", Address: "0.0.0.0:17170"},
		}, nil
	}
	killToolProcessFunc = func(pid int) error {
		return errors.New("boom")
	}

	err := CleanupPortsByPreference([]string{"17170"}, []string{"dtool"})
	if err == nil {
		t.Fatal("CleanupPortsByPreference() error = nil, want error")
	}
	if err.Error() != "清理端口 17170 的进程 111 失败: boom" {
		t.Fatalf("CleanupPortsByPreference() error = %q", err.Error())
	}
}
