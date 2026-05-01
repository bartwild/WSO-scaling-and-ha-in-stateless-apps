//go:build !amd64 || purego

package internal

func DotProductAVX2FMA(vector1 []float32, vector2 []float32) float32 {
	if len(vector1) == 0 || len(vector1) != len(vector2) {
		return 0
	}

	var sum float32

	_ = vector2[len(vector1)-1]
	for i, v1 := range vector1 {
		sum += v1 * vector2[i]
	}

	return sum
}
