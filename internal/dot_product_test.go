package internal

import (
	"fmt"
	"math"
	"testing"
)

func pureGoDotProduct(v1, v2 []float32) float32 {
	var sum float32
	for i := range v1 {
		sum += v1[i] * v2[i]
	}
	return sum
}

func TestDotProductAVX2FMA(t *testing.T) {
	tests := []struct {
		name    string
		v1      []float32
		v2      []float32
		wantRes float32
	}{
		{
			name:    "prosty iloczyn skalarny",
			v1:      []float32{1.0, 2.0, 3.0},
			v2:      []float32{4.0, 5.0, 6.0},
			wantRes: 32.0, // 1*4 + 2*5 + 3*6 = 4+10+18
		},
		{
			name:    "puste wektory",
			v1:      []float32{},
			v2:      []float32{},
			wantRes: 0,
		},
		{
			name:    "różne długości",
			v1:      []float32{1.0, 2.0},
			v2:      []float32{1.0, 2.0, 3.0},
			wantRes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes := DotProductAVX2FMA(tt.v1, tt.v2)

			if math.Abs(float64(gotRes-tt.wantRes)) > 1e-6 {
				t.Errorf("dotProductAVX2FMA() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func BenchmarkDotProduct(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		v1 := make([]float32, size)
		v2 := make([]float32, size)
		for i := range size {
			v1[i] = 1.1
			v2[i] = 2.2
		}
		b.Run(fmt.Sprintf("PureGo-Size%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pureGoDotProduct(v1, v2)
			}
		})
		b.Run(fmt.Sprintf("AVX2-Size%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = DotProductAVX2FMA(v1, v2)
			}
		})
	}
}
