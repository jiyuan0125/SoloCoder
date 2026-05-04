#ifndef MD5_STREAM_H
#define MD5_STREAM_H

#include "md5_core.h"
#include <stddef.h>

void md5_stream_init(MD5_CTX *ctx);
void md5_stream_update(MD5_CTX *ctx, const uint8_t *data, size_t len);
void md5_stream_final(MD5_CTX *ctx, uint8_t digest[MD5_DIGEST_LENGTH]);
void md5_stream_final_string(MD5_CTX *ctx, char str[MD5_STRING_LENGTH]);

void md5_memory(const uint8_t *data, size_t len, uint8_t digest[MD5_DIGEST_LENGTH]);
void md5_memory_string(const uint8_t *data, size_t len, char str[MD5_STRING_LENGTH]);

#endif
