#ifndef MD5_CORE_H
#define MD5_CORE_H

#include <stdint.h>
#include <stddef.h>

#define MD5_DIGEST_LENGTH 16
#define MD5_STRING_LENGTH 33

typedef struct {
    uint32_t state[4];
    uint64_t count;
    uint8_t buffer[64];
} MD5_CTX;

void md5_init(MD5_CTX *ctx);
void md5_update(MD5_CTX *ctx, const uint8_t *data, size_t len);
void md5_final(MD5_CTX *ctx, uint8_t digest[MD5_DIGEST_LENGTH]);
void md5_digest_to_string(const uint8_t digest[MD5_DIGEST_LENGTH], char str[MD5_STRING_LENGTH]);

#endif
