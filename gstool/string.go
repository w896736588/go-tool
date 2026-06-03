package gstool

import (
	"regexp"
	"strings"
	"unicode"
)

// StringTo_str_str 转为 _str_str
func StringTo_str_str(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for key, char := range charList {
			if key == 0 {
				dataByte = append(dataByte, CharToLower(char))
			} else {
				dataByte = append(dataByte, char)
			}
		}
		return string(dataByte)
	})
	return strings.Join(stringList, `_`)
}

// StringTo_StrStr 转为 StrStr
func StringTo_StrStr(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for key, char := range charList {
			if key == 0 {
				dataByte = append(dataByte, CharToUpper(char))
			} else {
				dataByte = append(dataByte, CharToLower(char))
			}
		}
		return string(dataByte)
	})
	return strings.Join(stringList, ``)
}

// StringTo_Str_Str 转为 _Str_Str
func StringTo_Str_Str(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for key, char := range charList {
			if key == 0 {
				dataByte = append(dataByte, CharToUpper(char))
			} else {
				dataByte = append(dataByte, char)
			}
		}
		return string(dataByte)
	})
	return strings.Join(stringList, `_`)
}

// StringTo_STRSTR 转为 STRSTR
func StringTo_STRSTR(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for _, char := range charList {
			dataByte = append(dataByte, CharToUpper(char))
		}
		return string(dataByte)
	})
	return strings.Join(stringList, ``)
}

// StringTo_STR_STR 转为 STR_STR
func StringTo_STR_STR(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for _, char := range charList {
			dataByte = append(dataByte, CharToUpper(char))
		}
		return string(dataByte)
	})
	return strings.Join(stringList, `_`)
}

// StringTo_strstr 转为 strstr
func StringTo_strstr(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for key, char := range charList {
			if key == 0 {
				dataByte = append(dataByte, CharToLower(char))
			} else {
				dataByte = append(dataByte, char)
			}
		}
		return string(dataByte)
	})
	return strings.Join(stringList, ``)
}

// StringTo_strStr 转为 strStr
func StringTo_strStr(source string) string {
	stringList := StringToWordList(source)
	ArrayWalk(&stringList, func(index int, charList string) string {
		dataByte := make([]int32, 0)
		for key, char := range charList {
			if key == 0 {
				if index == 0 {
					dataByte = append(dataByte, CharToLower(char))
				} else {
					dataByte = append(dataByte, CharToUpper(char))
				}
			} else {
				dataByte = append(dataByte, char)
			}
		}
		return string(dataByte)
	})
	return strings.Join(stringList, ``)
}

// StringToWordList 格式化字符串，返回单词列表 根据 空格 | - _ 分割
func StringToWordList(source string) []string {
	stringList := []string{source}
	stringList = splitSpace(stringList)
	stringList = splitUnderline(stringList)
	stringList = splitDash(stringList)
	stringList = splitVerticalBar(stringList)
	return stringList
}

// 根据空格分割字符串
func splitSpace(sourceList []string) []string {
	stringList := make([]string, 0)
	regex := regexp.MustCompile(`\s+`)
	for _, v := range sourceList {
		stringListTemp := regex.Split(v, -1)
		stringListTemp = ArrayFilterEmpty(&stringListTemp)
		stringList = append(stringList, stringListTemp...)
	}
	return stringList
}

// 根据竖杠分割字符串
func splitVerticalBar(sourceList []string) []string {
	stringList := make([]string, 0)
	for _, v := range sourceList {
		stringListTemp := strings.Split(v, `|`)
		stringListTemp = ArrayFilterEmpty(&stringListTemp)
		stringList = append(stringList, stringListTemp...)
	}
	return stringList
}

// 根据下划线分割字符串
func splitUnderline(sourceList []string) []string {
	stringList := make([]string, 0)
	for _, v := range sourceList {
		stringListTemp := strings.Split(v, `_`)
		stringListTemp = ArrayFilterEmpty(&stringListTemp)
		stringList = append(stringList, stringListTemp...)
	}
	return stringList
}

// 根据横杠分割字符串
func splitDash(sourceList []string) []string {
	stringList := make([]string, 0)
	for _, v := range sourceList {
		stringListTemp := strings.Split(v, `-`)
		stringListTemp = ArrayFilterEmpty(&stringListTemp)
		stringList = append(stringList, stringListTemp...)
	}
	return stringList
}

// StringIsCommonUse 判断是否是常用字符串 只包含英文（大小写）、数字、_和-的字符串
func StringIsCommonUse(str string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return re.MatchString(str)
}

// StringToAsciiIntList 字符串转为ascii拼接的字符串
func StringToAsciiIntList(str string) []int {
	if !StringIsAscii(str) {
		return make([]int, 0)
	}
	asciiIntList := make([]int, 0, len(str))
	for _, ch := range str {
		asciiIntList = append(asciiIntList, int(ch))
	}
	return asciiIntList
}

// StringIsAscii 判断字符串是否是ascii字符串
func StringIsAscii(str string) bool {
	for _, r := range str {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// StringToAsciiInt 字符串转ascii后累加为整数
func StringToAsciiInt(str string) int {
	asciiIntList := StringToAsciiIntList(str)
	asciiInt := 0
	for _, asciiIntVal := range asciiIntList {
		asciiInt += asciiIntVal
	}
	return asciiInt
}

// StringEndByStr 获取字符串以规定字符串结尾的字符串下标
func StringEndByStr(str, endStr string) int {
	if str == `` {
		return -1
	}
	return strings.Index(str, endStr)
}

// StringEndRemoveStr 移除字符串以规定字符串结尾的字符串
func StringEndRemoveStr(str, endStr string) string {
	endIndex := StringEndByStr(str, endStr)
	if endIndex == -1 {
		return str
	}
	return str[0:endIndex]
}

// StringEndRemoveStrList 移除字符串以规定字符串结尾的字符串
func StringEndRemoveStrList(str string, endStrList []string) string {
	for _, endStr := range endStrList {
		endIndex := StringEndByStr(str, endStr)
		if endIndex >= 0 {
			return StringEndRemoveStr(str, endStr)
		}
	}
	return str
}

// SReplaces 多次替换字符串
func SReplaces(str string, replaceList map[string]string) string {
	for key, value := range replaceList {
		str = strings.Replace(str, key, value, -1)
	}
	return str
}

// StringBytesLength 字符串字节长度
func StringBytesLength(str string) int {
	return len([]byte(str))
}

// SContains 判断字符串是否包含列表中任意一个字符
func SContains(str string, includeList []string) bool {
	for _, val := range includeList {
		if strings.Contains(str, val) {
			return true
		}
	}
	return false
}

// StringCountSubstr 计算字符串中子字符串出现的次数
func StringCountSubstr(s, sub string) int {
	count := 0
	subLen := len(sub)
	for i := 0; ; i += subLen {
		if idx := strings.Index(s[i:], sub); idx == -1 {
			break
		} else {
			count++
			i += idx // 调整索引位置以便下一次搜索
		}
	}
	return count
}

// StringCountSubstrRegex 计算字符串中子正则出现的次数
func StringCountSubstrRegex(s, regex string) int {
	re := regexp.MustCompile(regex)
	matches := re.FindAllString(s, -1)
	return len(matches)
}

// StringReplaceRegex 替换正则
func StringReplaceRegex(sql, regex, value string) string {
	re := regexp.MustCompile(regex)
	return re.ReplaceAllString(sql, value)
}

func StringSubstr(s string, start, num int) string {
	runes := []rune(s)
	if len(runes) > start+num {
		return string(runes[start:num])
	}
	return s
}

func SChunks(s string, n int) []string {
	var chunks []string
	current := ""
	count := 0

	for _, r := range s {
		current += string(r)
		count++
		if count == n {
			chunks = append(chunks, current)
			current = ""
			count = 0
		}
	}
	if current != "" {
		chunks = append(chunks, current)
	}
	return chunks
}

// StringFilterANSI 移除ANSI字符
func StringFilterANSI(in string) string {
	return regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`).ReplaceAllString(in, "")
}
