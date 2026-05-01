//go:build amd64 && cgo && !purego

#include "dot_product_amd64.h"
#include <immintrin.h>
#include <stdint.h>

// implementation for avx2
// __attribute__((target("avx2,fma")))
// float dotProductAVX2FMA(const float* vector1, const float* vector2, int64_t size) {
//     __m256 sumV = _mm256_setzero_ps();
//     int64_t i = 0;

//     for (; i <= size - 8; i += 8) {
//         __m256 v1 = _mm256_loadu_ps(&vector1[i]);
//         __m256 v2 = _mm256_loadu_ps(&vector2[i]);
//         sumV = _mm256_fmadd_ps(v1, v2, sumV);
//     }

//     __m128 vlow = _mm256_castps256_ps128(sumV);
//     __m128 vhigh = _mm256_extractf128_ps(sumV, 1);
//     __m128 vsum = _mm_add_ps(vlow, vhigh);

//     vsum = _mm_add_ps(vsum, _mm_movehl_ps(vsum, vsum));
//     vsum = _mm_add_ss(vsum, _mm_shuffle_ps(vsum, vsum, 0x55));

//     float result = _mm_cvtss_f32(vsum);

//     for (; i < size; i++) {
//         result += vector1[i] * vector2[i];
//     }

//     return result;
// }

// implementation for avx512
__attribute__((target("avx512f,avx512dq,avx512bw")))
float dotProductAVX2FMA(const float* v1, const float* v2, int64_t size) {
    if (size <= 0) return 0.0f;

    __m512 sum1 = _mm512_setzero_ps();
    __m512 sum2 = _mm512_setzero_ps();
    __m512 sum3 = _mm512_setzero_ps();
    __m512 sum4 = _mm512_setzero_ps();

    int64_t i = 0;
    for (; i <= size - 64; i += 64) {
        sum1 = _mm512_fmadd_ps(_mm512_loadu_ps(&v1[i]),    _mm512_loadu_ps(&v2[i]),    sum1);
        sum2 = _mm512_fmadd_ps(_mm512_loadu_ps(&v1[i+16]), _mm512_loadu_ps(&v2[i+16]), sum2);
        sum3 = _mm512_fmadd_ps(_mm512_loadu_ps(&v1[i+32]), _mm512_loadu_ps(&v2[i+32]), sum3);
        sum4 = _mm512_fmadd_ps(_mm512_loadu_ps(&v1[i+48]), _mm512_loadu_ps(&v2[i+48]), sum4);
    }

    __m512 finalSum = _mm512_add_ps(_mm512_add_ps(sum1, sum2), _mm512_add_ps(sum3, sum4));

    for (; i <= size - 16; i += 16) {
        finalSum = _mm512_fmadd_ps(_mm512_loadu_ps(&v1[i]), _mm512_loadu_ps(&v2[i]), finalSum);
    }

    int64_t remaining = size - i;
    if (remaining > 0) {
        __mmask16 mask = (__mmask16)((1U << remaining) - 1);
        __m512 a = _mm512_maskz_loadu_ps(mask, &v1[i]);
        __m512 b = _mm512_maskz_loadu_ps(mask, &v2[i]);
        finalSum = _mm512_fmadd_ps(a, b, finalSum);
    }

    return _mm512_reduce_add_ps(finalSum);
}