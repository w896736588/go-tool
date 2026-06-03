package gstool

// CharisUpper 判断字符是否是大写字母
func CharisUpper(char int32) bool {
	return char >= 'A' && char <= 'Z'
}

// CharisLower 判断字符是否是小写字母
func CharisLower(char int32) bool {
	return char >= 'a' && char <= 'z'
}

// CharIsChar 判断字符是否是字母
func CharIsChar(char int32) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

// CharToUpper 字符转大写
func CharToUpper(char int32) int32 {
	if CharisLower(char) {
		return char - 32
	}
	return char
}

// CharToLower 字符转小写
func CharToLower(char int32) int32 {
	if CharisUpper(char) {
		return char + 32
	}
	return char
}
