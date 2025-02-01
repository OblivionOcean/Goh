//go:build go1.20
// +build go1.20

package Goh

import (
	"unsafe"
)

func String2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func Byte2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
