//go:build amd64 && cgo && !purego

package internal

/*
#cgo CFLAGS: -O3
#include "dot_product_amd64.h"
*/
import "C"
import "unsafe"

func DotProductAVX2FMA(vector1 []float32, vector2 []float32) float32 {
	if len(vector1) == 0 || len(vector1) != len(vector2) {
		return 0
	}

	return float32(C.dotProductAVX2FMA(
		(*C.float)(unsafe.Pointer(&vector1[0])),
		(*C.float)(unsafe.Pointer(&vector2[0])),
		C.int64_t(len(vector1)),
	))
}
