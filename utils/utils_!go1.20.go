//go:build !go1.20
// +build !go1.20

package Goh

import (
	"reflect"
	"unsafe"
)

func String2Bytes(s string) []byte {
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	by := reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	}
	//在把by从sliceheader转为[]byte类型
	return *(*[]byte)(unsafe.Pointer(&by))
}

func Byte2String(b []byte) string {
	by := (*reflect.SliceHeader)(unsafe.Pointer(&b))

	str := reflect.StringHeader{
		Data: by.Data,
		Len:  by.Len,
	}
	return *(*string)(unsafe.Pointer(&str))
}
