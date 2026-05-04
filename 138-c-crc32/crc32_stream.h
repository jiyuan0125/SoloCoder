#ifndef CRC32_STREAM_H
#define CRC32_STREAM_H

#include <stdint.h>
#include <stddef.h>

typedef struct {
    uint32_t crc;
    int finalized;
} crc32_stream_t;

void crc32_stream_init(crc32_stream_t *stream);
void crc32_stream_reset(crc32_stream_t *stream);
void crc32_stream_update(crc32_stream_t *stream, const uint8_t *data, size_t len);
uint32_t crc32_stream_final(crc32_stream_t *stream);
void crc32_stream_final_hex(crc32_stream_t *stream, char *hex_str);

uint32_t crc32_combine(uint32_t crc1, uint32_t crc2, size_t len2);
uint32_t crc32_combine_hex(const char *crc1_hex, const char *crc2_hex, size_t len2, char *result_hex);

uint32_t crc32_combine_multi(const uint32_t *crcs, const size_t *lens, size_t count);
uint32_t crc32_stream_raw(crc32_stream_t *stream);

#endif
