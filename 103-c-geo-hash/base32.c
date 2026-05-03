#include "base32.h"
#include <string.h>
#include <ctype.h>

static const char base32_chars[] = GEOHASH_BASE32_CHARS;

int base32_char_to_index(char c)
{
    char lower = tolower(c);
    const char *pos = strchr(base32_chars, lower);
    if (pos == NULL) {
        return -1;
    }
    return (int)(pos - base32_chars);
}

char base32_index_to_char(int index)
{
    if (index < 0 || index >= 32) {
        return '\0';
    }
    return base32_chars[index];
}

uint64_t base32_decode(const char *hash, int precision)
{
    uint64_t result = 0;
    int len = (int)strlen(hash);
    int i;
    
    if (precision <= 0) {
        precision = len;
    }
    if (precision > BASE32_PRECISION_MAX) {
        precision = BASE32_PRECISION_MAX;
    }
    
    for (i = 0; i < precision && i < len; i++) {
        int idx = base32_char_to_index(hash[i]);
        if (idx == -1) {
            return 0;
        }
        result = (result << 5) | (uint64_t)idx;
    }
    
    while (i < precision) {
        result = (result << 5);
        i++;
    }
    
    return result;
}

void base32_encode(uint64_t value, int precision, char *result)
{
    int i;
    int shift;
    
    if (precision <= 0 || precision > BASE32_PRECISION_MAX) {
        result[0] = '\0';
        return;
    }
    
    shift = (precision - 1) * 5;
    
    for (i = 0; i < precision; i++) {
        uint64_t mask = (uint64_t)31 << shift;
        int idx = (int)((value & mask) >> shift);
        result[i] = base32_index_to_char(idx);
        shift -= 5;
    }
    
    result[precision] = '\0';
}
