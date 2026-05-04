#include "md5_file.h"
#include "md5_stream.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

static int hex_char_to_value(char c) {
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    return -1;
}

static void string_to_digest(const char *str, uint8_t digest[MD5_DIGEST_LENGTH]) {
    int i;
    for (i = 0; i < MD5_DIGEST_LENGTH; i++) {
        int high = hex_char_to_value(str[i * 2]);
        int low = hex_char_to_value(str[i * 2 + 1]);
        digest[i] = (uint8_t)((high << 4) | low);
    }
}

int md5_file(const char *filepath, uint8_t digest[MD5_DIGEST_LENGTH]) {
    FILE *file;
    MD5_CTX ctx;
    uint8_t *buffer;
    size_t bytes_read;

    file = fopen(filepath, "rb");
    if (file == NULL) {
        return -1;
    }

    buffer = (uint8_t *)malloc(MD5_FILE_CHUNK_SIZE);
    if (buffer == NULL) {
        fclose(file);
        return -1;
    }

    md5_stream_init(&ctx);

    while ((bytes_read = fread(buffer, 1, MD5_FILE_CHUNK_SIZE, file)) > 0) {
        md5_stream_update(&ctx, buffer, bytes_read);
    }

    if (ferror(file)) {
        free(buffer);
        fclose(file);
        return -1;
    }

    md5_stream_final(&ctx, digest);

    free(buffer);
    fclose(file);
    return 0;
}

int md5_file_string(const char *filepath, char str[MD5_STRING_LENGTH]) {
    uint8_t digest[MD5_DIGEST_LENGTH];
    int result = md5_file(filepath, digest);
    if (result == 0) {
        md5_digest_to_string(digest, str);
    }
    return result;
}

int md5_file_verify(const char *filepath, const char *expected_hash) {
    uint8_t actual_digest[MD5_DIGEST_LENGTH];
    uint8_t expected_digest[MD5_DIGEST_LENGTH];
    int i;

    if (expected_hash == NULL || strlen(expected_hash) != 32) {
        return -1;
    }

    if (md5_file(filepath, actual_digest) != 0) {
        return -1;
    }

    string_to_digest(expected_hash, expected_digest);

    for (i = 0; i < MD5_DIGEST_LENGTH; i++) {
        if (actual_digest[i] != expected_digest[i]) {
            return 0;
        }
    }

    return 1;
}
