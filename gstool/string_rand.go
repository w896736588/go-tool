package gstool

import (
	"math/rand"
	"sync"
	"time"
)

const (
	// CharsetNum 纯数字字符集
	CharsetNum = "0123456789"
	// CharsetLower 小写字母字符集
	CharsetLower = "abcdefghijklmnopqrstuvwxyz"
	// CharsetUpper 大写字母字符集
	CharsetUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// CharsetAlpha 大小写字母字符集
	CharsetAlpha = CharsetLower + CharsetUpper
	// CharsetAll 数字+大小写字母字符集（默认）
	CharsetAll = CharsetNum + CharsetAlpha
)

// RandString 随机字符串生成器
type RandString struct {
	charset string     // 自定义字符集
	rng     *rand.Rand // 每个实例独立的随机数生成器
	mu      sync.Mutex // 保证并发安全
}

var defaultRandString *RandString

func init() {
	defaultRandString = NewRandString(CharsetAll)
}

// NewRandString 创建随机字符串生成器实例
// charsets: 可选参数，指定自定义字符集；不传则使用默认的数字+大小写字母
func NewRandString(charsets ...string) *RandString {
	charset := CharsetAll
	if len(charsets) > 0 && charsets[0] != "" {
		charset = charsets[0]
	}

	// 每个实例创建独立的随机源，避免全局竞争
	source := rand.NewSource(time.Now().UnixNano() + int64(rand.Int63n(1000000)))
	return &RandString{
		charset: charset,
		rng:     rand.New(source),
	}
}

// Generate 生成指定长度的随机字符串
// n: 字符串长度，必须大于0，否则返回空字符串
func (rs *RandString) Generate(n int) string {
	// 边界检查
	if n <= 0 || rs.charset == "" {
		return ""
	}

	charsetLen := len(rs.charset)
	result := make([]byte, n)

	rs.mu.Lock()
	defer rs.mu.Unlock()
	for i := 0; i < n; i++ {
		idx := rs.rng.Intn(charsetLen)
		result[i] = rs.charset[idx]
	}
	return string(result)
}

// RandStringNum 生成纯数字的随机字符串
// n: 长度，必须大于0
func RandStringNum(n int) string {
	return NewRandString(CharsetNum).Generate(n)
}

// RandStringLower 生成小写字母的随机字符串
// n: 长度，必须大于0
func RandStringLower(n int) string {
	return NewRandString(CharsetLower).Generate(n)
}

// RandStringUpper 生成大写字母的随机字符串
// n: 长度，必须大于0
func RandStringUpper(n int) string {
	return NewRandString(CharsetUpper).Generate(n)
}

// RandStringAlpha 生成大小写字母的随机字符串
// n: 长度，必须大于0
func RandStringAlpha(n int) string {
	return NewRandString(CharsetAlpha).Generate(n)
}

// RandStringAll 生成数字+大小写字母的随机字符串（默认）
// n: 长度，必须大于0
func RandStringAll(n int) string {
	return defaultRandString.Generate(n)
}
