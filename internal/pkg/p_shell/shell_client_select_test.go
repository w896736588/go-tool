package p_shell

import (
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"gitee.com/Sxiaobai/gs/v2/gsssh"
)

func TestFindIdleClientIndexSkipsBusyClient(t *testing.T) {
	c1 := &gsssh.SshTerminal{}
	c2 := &gsssh.SshTerminal{}
	pool := []*gsssh.SshTerminal{c1, c2}

	oldInspector := terminalBusyInspector
	defer func() { terminalBusyInspector = oldInspector }()
	terminalBusyInspector = func(client *gsssh.SshTerminal) bool {
		if client == c1 {
			return true
		}
		return false
	}

	got := findIdleClientIndex(pool, 0)
	if got != 1 {
		t.Fatalf("findIdleClientIndex() = %d, want 1", got)
	}
}

func TestFindIdleClientIndexReturnsMinusOneWhenAllBusy(t *testing.T) {
	c1 := &gsssh.SshTerminal{}
	c2 := &gsssh.SshTerminal{}
	pool := []*gsssh.SshTerminal{c1, c2}

	oldInspector := terminalBusyInspector
	defer func() { terminalBusyInspector = oldInspector }()
	terminalBusyInspector = func(client *gsssh.SshTerminal) bool {
		return true
	}

	got := findIdleClientIndex(pool, 0)
	if got != -1 {
		t.Fatalf("findIdleClientIndex() = %d, want -1", got)
	}
}

func TestIsTerminalBusyByLockState(t *testing.T) {
	cli := &gsssh.SshTerminal{}
	if isTerminalBusy(cli) {
		t.Fatalf("expected unlocked terminal to be idle")
	}

	val := reflect.ValueOf(cli).Elem().FieldByName("lockCommand")
	if !val.IsValid() || !val.CanAddr() {
		t.Fatalf("lockCommand field not accessible")
	}
	mtx := (*sync.Mutex)(unsafe.Pointer(val.UnsafeAddr()))
	mtx.Lock()
	defer mtx.Unlock()

	if !isTerminalBusy(cli) {
		t.Fatalf("expected locked terminal to be busy")
	}
}
