#ifndef BASE64_ENCODE_H
#define BASE64_ENCODE_H

#include "base64.h"

#ifdef __cplusplus
extern "C" {
#endif

size_t base64_encode_size(size_t input_len);

int base64_encode(const uint8_t *input, size_t input_len,
                  char *output, size_t output_size,
                  size_t *output_len);

#ifdef __cplusplus
}
#endif

#endif
