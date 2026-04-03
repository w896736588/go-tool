package component

import "testing"

func TestComponentExposesRuntimeGlobals(t *testing.T) {
	if ShellOutClient != nil {
		t.Fatal("expected ShellOutClient to default to nil in package test")
	}
	if MemoryRuntime != nil {
		t.Fatal("expected MemoryRuntime to default to nil in package test")
	}
	if DbMain != nil {
		t.Fatal("expected DbMain to default to nil in package test")
	}
	if DbLog != nil {
		t.Fatal("expected DbLog to default to nil in package test")
	}
	if DataBaseUp != nil {
		t.Fatal("expected DataBaseUp to default to nil in package test")
	}
	if VariableClient != nil {
		t.Fatal("expected VariableClient to default to nil in package test")
	}
}
