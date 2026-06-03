package gstool

// StructCopyDeep 深度复制结构体
func StructCopyDeep(v any, t any) error {
	err := JsonDecode(JsonEncode(v), t)
	if err != nil {
		return err
	}
	return nil
}
