#include "url_codec.h"
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

static int is_safe_url_char(char c) {
    return isalnum((unsigned char)c) || c == '-' || c == '_' || c == '.' || c == '~';
}

static char hex_to_char(char c) {
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    return -1;
}

size_t url_encode(const char *src, char *dst, size_t dst_size) {
    const char *p = src;
    size_t needed = 0;
    
    while (*p) {
        if (is_safe_url_char(*p)) {
            needed++;
        } else {
            needed += 3;
        }
        p++;
    }
    
    if (!dst || dst_size == 0) {
        return needed;
    }
    
    p = src;
    size_t written = 0;
    
    while (*p && written < dst_size - 1) {
        if (is_safe_url_char(*p)) {
            dst[written++] = *p;
        } else {
            if (written + 3 > dst_size - 1) {
                break;
            }
            dst[written++] = '%';
            dst[written++] = "0123456789ABCDEF"[(unsigned char)(*p) >> 4];
            dst[written++] = "0123456789ABCDEF"[(unsigned char)(*p) & 0x0F];
        }
        p++;
    }
    
    if (dst_size > 0) {
        dst[written] = '\0';
    }
    
    return needed;
}

size_t url_decode(const char *src, char *dst, size_t dst_size) {
    const char *p = src;
    size_t needed = 0;
    
    while (*p) {
        if (*p == '%' && p[1] && p[2] && 
            isxdigit((unsigned char)p[1]) && isxdigit((unsigned char)p[2])) {
            needed++;
            p += 3;
        } else {
            needed++;
            p++;
        }
    }
    
    if (!dst || dst_size == 0) {
        return needed;
    }
    
    p = src;
    size_t written = 0;
    
    while (*p && written < dst_size - 1) {
        if (*p == '%' && p[1] && p[2] && 
            isxdigit((unsigned char)p[1]) && isxdigit((unsigned char)p[2])) {
            char c1 = hex_to_char(p[1]);
            char c2 = hex_to_char(p[2]);
            if (c1 >= 0 && c2 >= 0) {
                dst[written++] = (c1 << 4) | c2;
            } else {
                dst[written++] = *p;
            }
            p += 3;
        } else {
            dst[written++] = *p;
            p++;
        }
    }
    
    if (dst_size > 0) {
        dst[written] = '\0';
    }
    
    return needed;
}

char *url_encode_alloc(const char *src) {
    size_t needed = url_encode(src, NULL, 0) + 1;
    char *result = (char *)malloc(needed);
    if (result) {
        url_encode(src, result, needed);
    }
    return result;
}

char *url_decode_alloc(const char *src) {
    size_t needed = url_decode(src, NULL, 0) + 1;
    char *result = (char *)malloc(needed);
    if (result) {
        url_decode(src, result, needed);
    }
    return result;
}
