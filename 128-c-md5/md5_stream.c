#include "md5_stream.h"

void md5_stream_init(MD5_CTX *ctx) {
    md5_init(ctx);
}

void md5_stream_update(MD5_CTX *ctx, const uint8_t *data, size_t len) {
    md5_update(ctx, data, len);
}

void md5_stream_final(MD5_CTX *ctx, uint8_t digest[MD5_DIGEST_LENGTH]) {
    md5_final(ctx, digest);
}

void md5_stream_final_string(MD5_CTX *ctx, char str[MD5_STRING_LENGTH]) {
    uint8_t digest[MD5_DIGEST_LENGTH];
    md5_final(ctx, digest);
    md5_digest_to_string(digest, str);
}

void md5_memory(const uint8_t *data, size_t len, uint8_t digest[MD5_DIGEST_LENGTH]) {
    MD5_CTX ctx;
    md5_init(&ctx);
    md5_update(&ctx, data, len);
    md5_final(&ctx, digest);
}

void md5_memory_string(const uint8_t *data, size_t len, char str[MD5_STRING_LENGTH]) {
    uint8_t digest[MD5_DIGEST_LENGTH];
    md5_memory(data, len, digest);
    md5_digest_to_string(digest, str);
}
