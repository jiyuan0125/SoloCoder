#define _POSIX_C_SOURCE 200809L
#define _XOPEN_SOURCE 700

#include "fingerprint.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>

#define ROTL32(x, n) (((x) << (n)) | ((x) >> (32 - (n))))
#define ROTR32(x, n) (((x) >> (n)) | ((x) << (32 - (n))))

#define F1(x, y, z) ((x & y) | (~x & z))
#define F2(x, y, z) ((x & y) | (x & z) | (y & z))
#define F3(x, y, z) (x ^ y ^ z)

typedef struct {
    uint32_t state[4];
    uint32_t count[2];
    unsigned char buffer[64];
} MD5Context;

static const uint32_t md5_s[64] = {
    7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
    5,  9, 14, 20, 5,  9, 14, 20, 5,  9, 14, 20, 5,  9, 14, 20,
    4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
    6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21
};

static const uint32_t md5_k[64] = {
    0xd76aa478, 0xe8c7b756, 0x242070db, 0xc1bdceee,
    0xf57c0faf, 0x4787c62a, 0xa8304613, 0xfd469501,
    0x698098d8, 0x8b44f7af, 0xffff5bb1, 0x895cd7be,
    0x6b901122, 0xfd987193, 0xa679438e, 0x49b40821,
    0xf61e2562, 0xc040b340, 0x265e5a51, 0xe9b6c7aa,
    0xd62f105d, 0x02441453, 0xd8a1e681, 0xe7d3fbc8,
    0x21e1cde6, 0xc33707d6, 0xf4d50d87, 0x455a14ed,
    0xa9e3e905, 0xfcefa3f8, 0x676f02d9, 0x8d2a4c8a,
    0xfffa3942, 0x8771f681, 0x6d9d6122, 0xfde5380c,
    0xa4beea44, 0x4bdecfa9, 0xf6bb4b60, 0xbebfbc70,
    0x289b7ec6, 0xeaa127fa, 0xd4ef3085, 0x04881d05,
    0xd9d4d039, 0xe6db99e5, 0x1fa27cf8, 0xc4ac5665,
    0xf4292244, 0x432aff97, 0xab9423a7, 0xfc93a039,
    0x655b59c3, 0x8f0ccc92, 0xffeff47d, 0x85845dd1,
    0x6fa87e4f, 0xfe2ce6e0, 0xa3014314, 0x4e0811a1,
    0xf7537e82, 0xbd3af235, 0x2ad7d2bb, 0xeb86d391
};

static void md5_init(MD5Context *ctx) {
    ctx->count[0] = 0;
    ctx->count[1] = 0;
    ctx->state[0] = 0x67452301;
    ctx->state[1] = 0xefcdab89;
    ctx->state[2] = 0x98badcfe;
    ctx->state[3] = 0x10325476;
}

static void md5_transform(MD5Context *ctx, const unsigned char block[64]) {
    uint32_t a = ctx->state[0];
    uint32_t b = ctx->state[1];
    uint32_t c = ctx->state[2];
    uint32_t d = ctx->state[3];
    uint32_t x[16];

    for (int i = 0; i < 16; i++) {
        x[i] = ((uint32_t)block[i * 4]) |
               ((uint32_t)block[i * 4 + 1] << 8) |
               ((uint32_t)block[i * 4 + 2] << 16) |
               ((uint32_t)block[i * 4 + 3] << 24);
    }

    for (int i = 0; i < 64; i++) {
        uint32_t f, g;
        if (i < 16) {
            f = F1(b, c, d);
            g = i;
        } else if (i < 32) {
            f = F3(b, c, d);
            g = (5 * i + 1) % 16;
        } else if (i < 48) {
            f = F2(b, c, d);
            g = (3 * i + 5) % 16;
        } else {
            f = F3(b, c, d);
            g = (7 * i) % 16;
        }
        uint32_t temp = d;
        d = c;
        c = b;
        b = b + ROTL32((a + f + md5_k[i] + x[g]), md5_s[i]);
        a = temp;
    }

    ctx->state[0] += a;
    ctx->state[1] += b;
    ctx->state[2] += c;
    ctx->state[3] += d;
}

static void md5_update(MD5Context *ctx, const unsigned char *data, size_t len) {
    uint32_t index = (uint32_t)((ctx->count[0] >> 3) & 0x3F);
    if ((ctx->count[0] += (uint32_t)(len << 3)) < (uint32_t)(len << 3)) {
        ctx->count[1]++;
    }
    ctx->count[1] += (uint32_t)(len >> 29);
    
    uint32_t part_len = 64 - index;
    uint32_t i = 0;

    if (len >= part_len) {
        memcpy(&ctx->buffer[index], data, part_len);
        md5_transform(ctx, ctx->buffer);
        for (i = part_len; i + 63 < len; i += 64) {
            md5_transform(ctx, &data[i]);
        }
        index = 0;
    }
    memcpy(&ctx->buffer[index], &data[i], len - i);
}

static void md5_final(unsigned char digest[16], MD5Context *ctx) {
    unsigned char bits[8];
    for (int i = 0; i < 8; i++) {
        bits[i] = (unsigned char)((ctx->count[i >> 2] >> ((i & 3) << 3)) & 0xFF);
    }

    uint32_t index = (uint32_t)((ctx->count[0] >> 3) & 0x3f);
    uint32_t pad_len = (index < 56) ? (56 - index) : (120 - index);
    unsigned char padding[64] = {0x80};
    md5_update(ctx, padding, pad_len);
    md5_update(ctx, bits, 8);

    for (int i = 0; i < 4; i++) {
        for (int j = 0; j < 4; j++) {
            digest[i * 4 + j] = (unsigned char)((ctx->state[i] >> (j << 3)) & 0xff);
        }
    }
}

static void hash_to_hex(const unsigned char *hash, int hash_len, char *output, size_t output_len) {
    const char *hex = "0123456789abcdef";
    size_t max_bytes = (output_len - 1) / 2;
    if (max_bytes > (size_t)hash_len) max_bytes = hash_len;
    
    for (size_t i = 0; i < max_bytes; i++) {
        output[i * 2] = hex[(hash[i] >> 4) & 0x0F];
        output[i * 2 + 1] = hex[hash[i] & 0x0F];
    }
    output[max_bytes * 2] = '\0';
}

static FingerprintResult read_bytes_at_offset(const char *path, off_t offset, 
                                                unsigned char *buffer, size_t bytes_to_read) {
    int fd = open(path, O_RDONLY);
    if (fd < 0) return FP_ERROR_OPEN;

    if (offset != 0 && lseek(fd, offset, SEEK_SET) < 0) {
        close(fd);
        return FP_ERROR_SEEK;
    }

    ssize_t bytes_read = 0;
    size_t total_read = 0;
    
    while (total_read < bytes_to_read) {
        bytes_read = read(fd, buffer + total_read, bytes_to_read - total_read);
        if (bytes_read < 0) {
            close(fd);
            return FP_ERROR_READ;
        }
        if (bytes_read == 0) break;
        total_read += bytes_read;
    }

    close(fd);
    return FP_SUCCESS;
}

FingerprintResult compute_light_hash(const char *path, off_t size, char *hash_out, size_t hash_len) {
    if (!path || !hash_out || hash_len < 2) return FP_ERROR_MEMORY;
    
    MD5Context ctx;
    md5_init(&ctx);
    
    md5_update(&ctx, (const unsigned char *)&size, sizeof(off_t));
    
    off_t read_size = (size < LIGHT_FINGERPRINT_BYTES * 2) ? size : LIGHT_FINGERPRINT_BYTES;
    
    if (read_size > 0) {
        unsigned char *buffer = malloc(read_size);
        if (!buffer) return FP_ERROR_MEMORY;
        
        FingerprintResult result = read_bytes_at_offset(path, 0, buffer, read_size);
        if (result != FP_SUCCESS) {
            free(buffer);
            return result;
        }
        md5_update(&ctx, buffer, read_size);
        
        if (size > LIGHT_FINGERPRINT_BYTES * 2) {
            off_t tail_offset = size - LIGHT_FINGERPRINT_BYTES;
            result = read_bytes_at_offset(path, tail_offset, buffer, LIGHT_FINGERPRINT_BYTES);
            if (result != FP_SUCCESS) {
                free(buffer);
                return result;
            }
            md5_update(&ctx, buffer, LIGHT_FINGERPRINT_BYTES);
        }
        
        free(buffer);
    }
    
    unsigned char final_hash[16];
    md5_final(final_hash, &ctx);
    hash_to_hex(final_hash, 16, hash_out, hash_len);
    
    return FP_SUCCESS;
}

FingerprintResult compute_full_hash(const char *path, char *hash_out, size_t hash_len) {
    if (!path || !hash_out || hash_len < 2) return FP_ERROR_MEMORY;
    
    int fd = open(path, O_RDONLY);
    if (fd < 0) return FP_ERROR_OPEN;
    
    MD5Context ctx;
    md5_init(&ctx);
    
    unsigned char *buffer = malloc(STREAM_BUFFER_SIZE);
    if (!buffer) {
        close(fd);
        return FP_ERROR_MEMORY;
    }
    
    ssize_t bytes_read;
    while ((bytes_read = read(fd, buffer, STREAM_BUFFER_SIZE)) > 0) {
        md5_update(&ctx, buffer, bytes_read);
    }
    
    free(buffer);
    close(fd);
    
    if (bytes_read < 0) return FP_ERROR_READ;
    
    unsigned char final_hash[16];
    md5_final(final_hash, &ctx);
    hash_to_hex(final_hash, 16, hash_out, hash_len);
    
    return FP_SUCCESS;
}

FingerprintResult compute_file_hashes(FileInfo *file) {
    if (!file) return FP_ERROR_MEMORY;
    
    FingerprintResult result = compute_light_hash(file->path, file->size, 
                                                   file->light_hash, sizeof(file->light_hash));
    if (result != FP_SUCCESS) return result;
    
    return FP_SUCCESS;
}

void compute_hashes_for_list(FileList *files, int max_open_files) {
    (void)max_open_files;
    if (!files) return;
    
    for (size_t i = 0; i < files->count; i++) {
        compute_file_hashes(&files->files[i]);
    }
}

const char *fingerprint_error_string(FingerprintResult result) {
    switch (result) {
        case FP_SUCCESS: return "Success";
        case FP_ERROR_OPEN: return "Failed to open file";
        case FP_ERROR_READ: return "Failed to read file";
        case FP_ERROR_SEEK: return "Failed to seek in file";
        case FP_ERROR_MEMORY: return "Memory allocation failed";
        default: return "Unknown error";
    }
}
