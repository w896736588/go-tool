package lib_tool

//key是否存在
func MapExist(mapValue map[string]interface{}, keyValue string) bool {
	if _, ok := mapValue[keyValue]; ok {
		return true
	}
	return false
}

//复制map
func MapCopy(sourceMap map[string]interface{}) map[string]interface{} {
	copyMap := make(map[string]interface{})
	for k, v := range sourceMap {
		copyMap[k] = v
	}
	return copyMap
}
