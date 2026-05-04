#include "crc32_stream.h"
#include "crc32_table.h"
#include "crc32.h"
#include <string.h>

#define CRC32_POLY 0xEDB88320U

static uint32_t crc32_zeros_naive(uint32_t crc, size_t len)
{
    size_t i;
    uint8_t zero = 0;
    
    for (i = 0; i < len; i++) {
        uint8_t index = (uint8_t)(crc ^ zero);
        crc = (crc >> 8) ^ crc32_table[index];
    }
    
    return crc;
}

typedef uint32_t gf2_matrix[32];

static void gf2_matrix_copy(gf2_matrix dst, const gf2_matrix src)
{
    int i;
    for (i = 0; i < 32; i++) {
        dst[i] = src[i];
    }
}

static uint32_t gf2_matrix_times(const gf2_matrix mat, uint32_t vec)
{
    uint32_t result = 0;
    int i;
    
    for (i = 0; vec != 0; i++, vec >>= 1) {
        if (vec & 1) {
            result ^= mat[i];
        }
    }
    
    return result;
}

static void gf2_matrix_square(gf2_matrix square, const gf2_matrix mat)
{
    int i;
    
    for (i = 0; i < 32; i++) {
        square[i] = gf2_matrix_times(mat, mat[i]);
    }
}

static void crc32_zeros_matrix(gf2_matrix mat)
{
    int i;
    uint32_t crc;
    uint32_t bit;
    
    for (i = 0; i < 32; i++) {
        bit = 1U << i;
        crc = crc32_zeros_naive(bit, 1);
        mat[i] = crc;
    }
}

static uint32_t crc32_apply_zeros(uint32_t crc, size_t len)
{
    gf2_matrix mat;
    gf2_matrix square;
    int i;
    
    if (len == 0) {
        return crc;
    }
    
    if (!crc32_table_is_initialized()) {
        crc32_table_init();
    }
    
    if (len <= 64) {
        return crc32_zeros_naive(crc, len);
    }
    
    crc32_zeros_matrix(mat);
    
    for (i = 0; len > 0; i++, len >>= 1) {
        if (len & 1) {
            crc = gf2_matrix_times(mat, crc);
        }
        gf2_matrix_square(square, mat);
        gf2_matrix_copy(mat, square);
    }
    
    return crc;
}

void crc32_stream_init(crc32_stream_t *stream)
{
    if (stream == NULL) {
        return;
    }
    stream->crc = CRC32_INIT_VALUE;
    stream->finalized = 0;
    
    if (!crc32_table_is_initialized()) {
        crc32_table_init();
    }
}

void crc32_stream_reset(crc32_stream_t *stream)
{
    crc32_stream_init(stream);
}

void crc32_stream_update(crc32_stream_t *stream, const uint8_t *data, size_t len)
{
    size_t i;
    
    if (stream == NULL || data == NULL || len == 0 || stream->finalized) {
        return;
    }
    
    for (i = 0; i < len; i++) {
        uint8_t index = (uint8_t)(stream->crc ^ data[i]);
        stream->crc = (stream->crc >> 8) ^ crc32_table[index];
    }
}

uint32_t crc32_stream_final(crc32_stream_t *stream)
{
    if (stream == NULL) {
        return 0;
    }
    
    if (!stream->finalized) {
        stream->crc ^= CRC32_FINAL_XOR;
        stream->finalized = 1;
    }
    
    return stream->crc;
}

void crc32_stream_final_hex(crc32_stream_t *stream, char *hex_str)
{
    uint32_t crc;
    
    if (hex_str == NULL) {
        return;
    }
    
    crc = crc32_stream_final(stream);
    crc32_to_hex(crc, hex_str);
}

uint32_t crc32_stream_raw(crc32_stream_t *stream)
{
    if (stream == NULL) {
        return 0;
    }
    return stream->crc;
}

uint32_t crc32_combine(uint32_t crc1, uint32_t crc2, size_t len2)
{
    uint32_t crc1_raw, crc2_raw;
    uint32_t zeros_after_crc1, zeros_after_init;
    uint32_t delta, combined_raw;
    
    if (len2 == 0) {
        return crc1;
    }
    
    if (!crc32_table_is_initialized()) {
        crc32_table_init();
    }
    
    crc1_raw = crc1 ^ CRC32_FINAL_XOR;
    crc2_raw = crc2 ^ CRC32_FINAL_XOR;
    
    zeros_after_crc1 = crc32_apply_zeros(crc1_raw, len2);
    zeros_after_init = crc32_apply_zeros(CRC32_INIT_VALUE, len2);
    
    delta = zeros_after_init ^ crc2_raw;
    combined_raw = zeros_after_crc1 ^ delta;
    
    return combined_raw ^ CRC32_FINAL_XOR;
}

uint32_t crc32_combine_hex(const char *crc1_hex, const char *crc2_hex, size_t len2, char *result_hex)
{
    uint32_t crc1, crc2, result;
    
    if (crc1_hex == NULL || crc2_hex == NULL) {
        if (result_hex != NULL) {
            result_hex[0] = '\0';
        }
        return 0;
    }
    
    crc1 = crc32_parse_hex(crc1_hex);
    crc2 = crc32_parse_hex(crc2_hex);
    
    result = crc32_combine(crc1, crc2, len2);
    
    if (result_hex != NULL) {
        crc32_to_hex(result, result_hex);
    }
    
    return result;
}

uint32_t crc32_combine_multi(const uint32_t *crcs, const size_t *lens, size_t count)
{
    size_t i;
    uint32_t result;
    
    if (crcs == NULL || lens == NULL || count == 0) {
        return 0;
    }
    
    if (count == 1) {
        return crcs[0];
    }
    
    result = crcs[0];
    for (i = 1; i < count; i++) {
        result = crc32_combine(result, crcs[i], lens[i]);
    }
    
    return result;
}
