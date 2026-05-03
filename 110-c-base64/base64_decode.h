#ifndef BASE64_DECODE_H
#define BASE64_DECODE_H

#include "base64.h"

#ifdef __cplusplus
extern "C" {
#endif

size_t base64_decode_size(size_t input_len);

int base64_decode(const char *input, size_t input_len,
                  uint8_t *output, size_t output_size,
                  size_t *output_len);

#ifdef __cplusplus
}
#endif

#endif
