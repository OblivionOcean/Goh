package Goh

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// CountByte counts the number of occurrences of sep in s. It delegates to
// strings.Count, which for a single-byte sep dispatches to the
// assembly-optimized bytealg.CountString (SSE/AVX2 on amd64/arm64), so it
// is far faster than a portable Go loop.
func CountByte(s string, sep byte) int {
	return strings.Count(s, string(sep))
}

var (
	escapedValues = []string{"&amp;", "&#39;", "&lt;", "&gt;", "&#34;"}
)

//go:inline
func EscapeHTML(html string, buffer *bytes.Buffer) {
	var i, l, k int
	htmlLength := len(html)
	buffer.Grow(htmlLength)
	for ; i < htmlLength; i++ {
		switch html[i] {
		case '&':
			k = 0
		case '\'':
			k = 1
		case '<':
			k = 2
		case '>':
			k = 3
		case '"':
			k = 4
		default:
			continue
		}
		buffer.WriteString(html[l:i])
		buffer.WriteString(escapedValues[k])
		l = i + 1
	}
	buffer.WriteString(html[l:])
}

func FormatFloat(n float64, buf *bytes.Buffer) {
	buf.WriteString(strconv.FormatFloat(n, 'f', -1, 64))
}

func FormatInt(n int64, buf *bytes.Buffer) {
	buf.WriteString(strconv.FormatInt(n, 10))
}

func FormatUint(n uint64, buf *bytes.Buffer) {
	buf.WriteString(strconv.FormatUint(n, 10))
}

func FormatBool(n bool, buf *bytes.Buffer) {
	if n {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}
}

func FormatAny(n any, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprint(n))
}
