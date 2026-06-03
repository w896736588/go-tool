package gstool

import (
	"errors"
	"strings"

	"github.com/tidwall/gjson"
)

type FlatItem struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	Desc string `json:"desc"`
}

func JsonFlatPaths(jsonStr string) ([]FlatItem, error) {
	if !gjson.Valid(jsonStr) {
		return nil, errors.New("invalid JSON")
	}

	var results []FlatItem
	extractAllPaths(gjson.Parse(jsonStr), "", &results, true)
	return results, nil
}

// extractAllPaths 使用gjson保持原始顺序
func extractAllPaths(data gjson.Result, currentPath string, results *[]FlatItem, addCurrent bool) {
	// 添加当前路径（如果是有效路径）
	if addCurrent && currentPath != "" {
		cleanPath := strings.TrimPrefix(currentPath, ".")
		appendFlatItem(results, cleanPath, getGJSONTypeDesc(data))
	}

	// 递归处理子元素
	if data.IsObject() {
		data.ForEach(func(key, val gjson.Result) bool {
			newPath := currentPath
			if currentPath == "" {
				newPath = key.String()
			} else {
				newPath = currentPath + "." + key.String()
			}
			extractAllPaths(val, newPath, results, true)
			return true
		})
	} else if data.IsArray() {
		arr := data.Array()
		if len(arr) > 0 {
			newPath := currentPath
			if currentPath != "" && !strings.HasSuffix(currentPath, "]") {
				newPath = currentPath + "[0]"
			}
			// 添加数组路径本身
			if newPath != currentPath && newPath != "" {
				cleanPath := strings.TrimPrefix(newPath, ".")
				appendFlatItem(results, cleanPath, `array`)
			}
			for _, item := range arr {
				extractAllPaths(item, newPath, results, false)
			}
		}
	}
}

func getGJSONTypeDesc(data gjson.Result) string {
	return strings.ToLower(data.Type.String())
}

func appendFlatItem(results *[]FlatItem, key, typ string) {
	for _, item := range *results {
		if item.Key == key {
			return
		}
	}
	*results = append(*results, FlatItem{
		Key:  key,
		Type: typ,
		Desc: ``,
	})
}
