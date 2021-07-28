package lib

import (
	"strings"
	"time"
)

const (
	TimeFormat0 = "YYYY-MM-DD HH:MM:SS"
	TimeFormat1 = "YYYY/MM/DD HH:MM:SS"
	TimeFormat2 = "MM-DD-YYYY HH:MM:SS"
	TimeFormat3 = "MM/DD/YYYY HH:MM:SS"
	TimeFormat4 = "YYYY-MM-DD"
	TimeFormat5 = "HH:MM:SS"
	TimeFormat6 = "HH:MM"
	TimeFormat7 = "YYYY-MM-DD-HHMMSS"
	TimeFormat8 = "YYYYMMDDHHMMSS"
)

//FormatDatatime 按照惯例格式参数格式化时间，如YYYY-MM-DD
func FormatDateTime(layout string, t time.Time) string {
	switch strings.ToUpper(layout) {
	case TimeFormat0:
		return t.Format("2006-01-02 15:04:05")
	case TimeFormat1:
		return t.Format("2006/01/02 15:04:05")
	case TimeFormat2:
		return t.Format("01-02-2006 15:04:05")
	case TimeFormat3:
		return t.Format("01/02/2006 15:04:05")
	case TimeFormat4:
		return t.Format("2006-01-02")
	case TimeFormat5:
		return t.Format("15:04:05")
	case TimeFormat6:
		return t.Format("15:04")
	case TimeFormat7:
		return t.Format("2006-01-02 150405")
	case TimeFormat8:
		return t.Format("20060102150405")
	default:
		return t.Format("2006-01-02 15:04:05")
	}

}
