package gstool

import (
	"strings"
	"time"
)

func DateCurrent() string {
	now := time.Now()
	return now.Format("2006-01-02 15:04:05")
}

func DataCurrentByInt(timeInt int64) string {
	t := time.Unix(timeInt, 0)
	return t.Format("2006-01-02 15:04:05")
}

func DateCurrentDate1() string {
	now := time.Now()
	return now.Format("2006-01-02")
}

func DateCurrentDate2() string {
	now := time.Now()
	return now.Format("20060102")
}

func DateFormat(format string, timeInt int64) string {
	format = formatFormat(format)
	t := time.Unix(timeInt, 0)
	return t.Format(format)
}

// DateCurrentFormat 格式化时间 例如y/m/d H:i
func DateCurrentFormat(format string) string {
	format = formatFormat(format)
	t := time.Now()
	return t.Format(format)
}

func formatFormat(format string) string {
	format = strings.Replace(format, `y`, `2006`, 1)
	format = strings.Replace(format, `Y`, `2006`, 1)
	format = strings.Replace(format, `m`, `01`, 1)
	format = strings.Replace(format, `d`, `02`, 1)
	format = strings.Replace(format, `H`, `15`, 1)
	format = strings.Replace(format, `i`, `04`, 1)
	format = strings.Replace(format, `s`, `05`, 1)
	return format
}

func DateCurrentStartStr() string {
	dateStr := DateCurrentFormat(`Y-m-d`)
	return dateStr + ` 00:00:00`
}

func DateCurrentStartUnix() int64 {
	ti, err := time.Parse(formatFormat(`Y-m-d H:i:s`), DateCurrentStartStr())
	if err != nil {
		return 0
	}
	return ti.Unix()
}

func DateCurrentEndStr() string {
	dateStr := DateCurrentFormat(`Y-m-d`)
	return dateStr + ` 23:59:59`
}

func DateCurrentEndUnix() int64 {
	ti, err := time.Parse(formatFormat(`Y-m-d H:i:s`), DateCurrentEndStr())
	if err != nil {
		return 0
	}
	return ti.Unix()
}
