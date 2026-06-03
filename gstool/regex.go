package gstool

import (
	"github.com/spf13/cast"
	"regexp"
)

// RegexNumberChinese 匹配数字和汉字，获取其中的数字
// 例如：1 二，43,如果 从其中提取出1和43
func RegexNumberChinese(source string) ([]string, []string) {
	matchDList := make([]int, 0)
	reg := regexp.MustCompile(`\d+[:：，, ]*[\x{4e00}-\x{9fa5}]+`)
	matchList := reg.FindAllString(source, -1)
	if len(matchList) == 0 {
		return Array2Str(&matchDList), matchList
	}
	regD := regexp.MustCompile(`\d+`)
	for _, match := range matchList {
		matchDListTemp := regD.FindAllString(match, -1)
		if len(matchDListTemp) == 0 {
			continue
		}
		matchDList = append(matchDList, cast.ToInt(matchDListTemp[0]))
	}
	return Array2Str(&matchDList), matchList
}

// RegexSearchString 搜索所有匹配的字符串
func RegexSearchString(sourceStr string, reg string) []string {
	re, err := regexp.Compile(reg)
	if err != nil {
		return make([]string, 0)
	}
	return re.FindAllString(sourceStr, -1)
}

// RegexMatchString 匹配字符串
func RegexMatchString(sourceStr string, reg string) bool {
	regex := regexp.MustCompile(reg)
	return regex.MatchString(sourceStr)
}

// RegexMatchSubString 匹配子字符串 括号
// 例如：`id%(\d+)` 匹配的结果为[]string{"id%1" : 1 , "id%100" : 100}
func RegexMatchSubString(sourceStr string, reg string) []string {
	regex := regexp.MustCompile(reg)
	return regex.FindStringSubmatch(sourceStr)
}
