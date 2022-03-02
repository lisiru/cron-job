package utils

import (
	"delay-queue/common"
	"math/rand"
	"strconv"
)

func StringToInt(str string) int {
	ret, err := strconv.Atoi(str)
	if err != nil {
		ret = 0
	}
	return ret
}

func IntToString(str int) string {
	return strconv.Itoa(str)
}

func StringToInt64(str string) int64 {
	ret, err := strconv.Atoi(str)
	if err != nil {
		ret = 0
	}
	return int64(ret)
}

func StringToUint64(str string) uint64  {
	ret, err := strconv.Atoi(str)
	if err != nil {
		ret = 0
	}
	return uint64(ret)
}

func Int64ToString(num int64) string  {
	return strconv.FormatInt(num,10)
}

func RandomStr(n int) string  {
	strs:=common.STRS
	b:=make([]byte,n)
	for i:=range b{
		b[i]=strs[rand.Intn(len(strs))]
	}
	return string(b)
}
