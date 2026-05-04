#ifndef FINGERPRINT_H
#define FINGERPRINT_H

#include "common.h"

#define LIGHT_FINGERPRINT_BYTES 1024
#define STREAM_BUFFER_SIZE 65536

typedef enum {
    FP_SUCCESS = 0,
    FP_ERROR_OPEN,
    FP_ERROR_READ,
    FP_ERROR_SEEK,
    FP_ERROR_MEMORY
} FingerprintResult;

FingerprintResult compute_light_hash(const char *path, off_t size, char *hash_out, size_t hash_len);
FingerprintResult compute_full_hash(const char *path, char *hash_out, size_t hash_len);
FingerprintResult compute_file_hashes(FileInfo *file);

void compute_hashes_for_list(FileList *files, int max_open_files);

const char *fingerprint_error_string(FingerprintResult result);

#endif
