#ifndef BASE64_STREAM_H
#define BASE64_STREAM_H

#include "base64.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    uint8_t leftover[2];
    size_t leftover_len;
    size_t column;
} base64_encode_stream_t;

typedef struct {
    uint8_t group[4];
    size_t group_idx;
} base64_decode_stream_t;

void base64_encode_stream_init(base64_encode_stream_t *stream);

int base64_encode_stream_update(base64_encode_stream_t *stream,
                                 const uint8_t *input, size_t input_len,
                                 char *output, size_t output_size,
                                 size_t *output_len);

int base64_encode_stream_final(base64_encode_stream_t *stream,
                                char *output, size_t output_size,
                                size_t *output_len);

void base64_decode_stream_init(base64_decode_stream_t *stream);

int base64_decode_stream_update(base64_decode_stream_t *stream,
                                 const char *input, size_t input_len,
                                 uint8_t *output, size_t output_size,
                                 size_t *output_len);

int base64_decode_stream_final(base64_decode_stream_t *stream,
                                uint8_t *output, size_t output_size,
                                size_t *output_len);

#ifdef __cplusplus
}
#endif

#endif
