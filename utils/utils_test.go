package Goh

import "testing"

func TestCountByte(t *testing.T) {
	bk := "`"  // a single backtick
	dq := "\"" // a single double-quote
	tests := []struct {
		name string
		s    string
		sep  byte
		want int
	}{
		{"empty", "", '`', 0},
		{"none", "abcdef", '`', 0},
		{"one", "a" + bk + "b", '`', 1},
		{"two", "a" + bk + "b" + bk + "c", '`', 2},
		{"three", bk + "a" + bk + "b" + bk + "c", '`', 3},
		{"all", bk + bk + bk, '`', 3},
		{"doublequote", "a" + dq + "b" + dq + "c", dq[0], 2},
		{"mixed-sep-ignored", "a" + bk + "b" + dq + "c", dq[0], 1},
		{"unicode-ignored", "héllo" + bk + "wörld", '`', 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountByte(tt.s, tt.sep); got != tt.want {
				t.Fatalf("CountByte(%q, %q) = %d, want %d", tt.s, tt.sep, got, tt.want)
			}
		})
	}
}
