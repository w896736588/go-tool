package gstool

import "testing"

func TestJsonFlatPathsUsesNonEmptyArraySample(t *testing.T) {
	jsonStr := `{"res":0,"msg":"ok","data":{"list":[{"id":"3","app_list":[]},{"id":"2","app_list":[{"id":212455,"app_name":"芝麻云帆"}]}]}}`

	items, err := JsonFlatPaths(jsonStr)
	if err != nil {
		t.Fatalf("JsonFlatPaths() error = %v", err)
	}
	FmtPrintlnLog("JsonFlatPaths() = %#v", items)
	keys := make(map[string]bool, len(items))
	for _, item := range items {
		keys[item.Key] = true
	}

	for _, key := range []string{
		"data.list[0].app_list[0]",
		"data.list[0].app_list[0].id",
		"data.list[0].app_list[0].app_name",
	} {
		if !keys[key] {
			t.Fatalf("JsonFlatPaths() missing key %q; got %#v", key, items)
		}
	}
}
