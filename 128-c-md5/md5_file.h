#ifndef MD5_FILE_H
#define MD5_FILE_H

#include "md5_core.h"
#include <stddef.h>

#define MD5_FILE_CHUNK_SIZE (64 * 1024)

int md5_file(const char *filepath, uint8_t digest[MD5_DIGEST_LENGTH]);
int md5_file_string(const char *filepath, char str[MD5_STRING_LENGTH]);

int md5_file_verify(const char *filepath, const char *expected_hash);

#endif
