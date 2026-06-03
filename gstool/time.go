package gstool

import (
	"time"
)

// TimeNowMilliInt64 获取当前时间毫秒
func TimeNowMilliInt64() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// TimeNowUnixToString 获取当前时间格式化
func TimeNowUnixToString(format string) string {
	now := time.Now()
	return TimeUnixToString(now, format)
}

// TimeUnixToString 时间格式化
func TimeUnixToString(t time.Time, format string) string {
	if format == `` {
		format = `Y-m-d H:i:s`
	}
	formatDate := SReplaces(format, map[string]string{
		`Y`: `2006`,
		`m`: `01`,
		`d`: `02`,
		`H`: `15`,
		`i`: `04`,
		`s`: `05`,
	})
	return t.Format(formatDate)
}

// TimeStringToUnix 时间字符串转时间戳
func TimeStringToUnix(timeStr, format string) (time.Time, error) {
	formatDate := TimeGetFormatDate(format)
	t, err := time.ParseInLocation(formatDate, timeStr, time.Local)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func TimeGetFormatDate(format string) string {
	formatDate := ``
	switch format {
	case `Y-m-d`:
		formatDate = `2006-01-02`
	case `Ymd`:
		formatDate = `20060102`
	case `Y/m/d`:
		formatDate = `2006/01/02`
	case `Y-m-d H:i`:
		formatDate = `2006-01-02 15:04`
	case `Ymd H:i`:
		formatDate = `20060102 15:04`
	case `Y/m/d H:i`:
		formatDate = `2006/01/02 15:04`
	case `Y-m-d H:i:s`:
		formatDate = `2006-01-02 15:04:05`
	case `Ymd H:i:s`:
		formatDate = `20060102 15:04:05`
	case `YmdHis`:
		formatDate = `20060102150405`
	case `Y/m/d H:i:s`:
		formatDate = `2006/01/02 15:04:05`
	default:
		formatDate = `2006-01-02 15:04:05`
	}
	return formatDate
}

func TimeAsiaShangHai() {
	loc, locErr := time.LoadLocation("Asia/Shanghai")
	if locErr != nil {
		loc = time.FixedZone("CST", 8*3600) // 手动设置东八区
	}
	time.Local = loc
}
