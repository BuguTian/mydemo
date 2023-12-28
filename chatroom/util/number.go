package util

import (
	"fmt"
	"strconv"
	"os"
)

//string转成int
func Str2int(str string) int {
	res, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return res
}

//string转成int64
func Str2int64(str string) int64 {
	res, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return res
}

//int转成string
func Num2str(num int) string {
	return strconv.Itoa(num)
}

//int64转成string
func LNum2str(num int64) string {
	return strconv.FormatInt(num,10)
}

func GetSubPath(sid uint64) string {
	subf := sid & 0xFF;
	return fmt.Sprintf("%02X", subf)
}

func EnsureDir(path string) bool{
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return false
	}
	return true
}

func IsFileExist(path string) (bool, error) {
    fileInfo, err := os.Stat(path)
    if os.IsNotExist(err) {
        return false, nil
    }
    if fileInfo.Size() == 0 {
        return false, nil
    }
    if err == nil {
        return true, nil
    }
    return false, err
}

func AnyToInt(t1 any) int {
	var t2 int
	switch t1.(type) {
	case uint:
		t2 = int(t1.(uint))
		break
	case int8:
		t2 = int(t1.(int8))
		break
	case uint8:
		t2 = int(t1.(uint8))
		break
	case int16:
		t2 = int(t1.(int16))
		break
	case uint16:
		t2 = int(t1.(uint16))
		break
	case int32:
		t2 = int(t1.(int32))
		break
	case uint32:
		t2 = int(t1.(uint32))
		break
	case int64:
		t2 = int(t1.(int64))
		break
	case uint64:
		t2 = int(t1.(uint64))
		break
	case float32:
		t2 = int(t1.(float32))
		break
	case float64:
		t2 = int(t1.(float64))
		break
	case string:
		t2, _ = strconv.Atoi(t1.(string))
		break
	default:
		t2 = t1.(int)
		break
	}
	return t2
}
