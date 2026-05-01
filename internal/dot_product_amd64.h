//go:build amd64 && cgo && !purego

#ifndef DOT_SIMD_H
#define DOT_SIMD_H

#include <stdint.h>

float dotProductAVX2FMA(const float* v1, const float* v2, int64_t size);

#endif