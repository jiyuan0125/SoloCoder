#include "code_generator.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#ifdef _WIN32
#include <windows.h>
#include <wincrypt.h>
#else
#include <fcntl.h>
#include <unistd.h>
#endif

struct CodeGenerator {
    int dummy;
};

static bool get_random_bytes(uint8_t* buffer, size_t len) {
#ifdef _WIN32
    HCRYPTPROV hProvider;
    if (!CryptAcquireContext(&hProvider, NULL, NULL, PROV_RSA_FULL, 
                              CRYPT_VERIFYCONTEXT)) {
        return false;
    }
    BOOL result = CryptGenRandom(hProvider, (DWORD)len, buffer);
    CryptReleaseContext(hProvider, 0);
    return result != FALSE;
#else
    int fd = open("/dev/urandom", O_RDONLY);
    if (fd < 0) return false;
    
    ssize_t total = 0;
    while (total < (ssize_t)len) {
        ssize_t n = read(fd, buffer + total, len - total);
        if (n <= 0) {
            close(fd);
            return false;
        }
        total += n;
    }
    close(fd);
    return true;
#endif
}

static uint32_t get_random_uint32(void) {
    uint8_t bytes[4];
    if (!get_random_bytes(bytes, 4)) {
        return (uint32_t)rand();
    }
    return (uint32_t)bytes[0] | 
           (uint32_t)bytes[1] << 8 |
           (uint32_t)bytes[2] << 16 |
           (uint32_t)bytes[3] << 24;
}

static int get_random_index(void) {
    uint32_t rand_val = get_random_uint32();
    return (int)(rand_val % BASE58_CHAR_COUNT);
}

CodeGenerator* code_generator_create(void) {
    CodeGenerator* gen = (CodeGenerator*)malloc(sizeof(CodeGenerator));
    if (gen) {
        gen->dummy = 0;
    }
    return gen;
}

void code_generator_destroy(CodeGenerator* gen) {
    if (gen) {
        free(gen);
    }
}

ReturnCode code_generator_generate_single(CodeGenerator* gen, char* out_code, size_t out_size) {
    (void)gen;
    
    if (!out_code || out_size < INVITE_CODE_LENGTH + 1) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    for (int i = 0; i < INVITE_CODE_LENGTH; i++) {
        int idx = get_random_index();
        out_code[i] = BASE58_ALPHABET[idx];
    }
    out_code[INVITE_CODE_LENGTH] = '\0';
    
    return RC_SUCCESS;
}

ReturnCode code_generator_generate_batch(CodeGenerator* gen, char** out_codes, int count) {
    if (!out_codes || count <= 0) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    for (int i = 0; i < count; i++) {
        out_codes[i] = (char*)malloc(INVITE_CODE_LENGTH + 1);
        if (!out_codes[i]) {
            code_generator_free_batch(out_codes, i);
            return RC_ERROR_STORAGE;
        }
        
        ReturnCode rc = code_generator_generate_single(gen, out_codes[i], INVITE_CODE_LENGTH + 1);
        if (rc != RC_SUCCESS) {
            code_generator_free_batch(out_codes, i + 1);
            return rc;
        }
    }
    
    return RC_SUCCESS;
}

void code_generator_free_batch(char** codes, int count) {
    if (!codes) return;
    
    for (int i = 0; i < count; i++) {
        if (codes[i]) {
            free(codes[i]);
            codes[i] = NULL;
        }
    }
}

bool code_generator_is_valid_char(char c) {
    return strchr(BASE58_ALPHABET, c) != NULL;
}

int code_generator_char_to_index(char c) {
    const char* pos = strchr(BASE58_ALPHABET, c);
    if (!pos) return -1;
    return (int)(pos - BASE58_ALPHABET);
}
