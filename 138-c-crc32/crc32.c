#include "crc32.h"
#include "crc32_table.h"
#include <stdio.h>
#include <string.h>
#include <ctype.h>

static uint32_t crc32_update(uint32_t crc, const uint8_t *data, size_t len)
{
    size_t i;
    
    if (!crc32_table_is_initialized()) {
        crc32_table_init();
    }
    
    for (i = 0; i < len; i++) {
        uint8_t index = (uint8_t)(crc ^ data[i]);
        crc = (crc >> 8) ^ crc32_table[index];
    }
    
    return crc;
}

uint32_t crc32_compute(const uint8_t *data, size_t len)
{
    uint32_t crc;
    
    if (data == NULL || len == 0) {
        return 0;
    }
    
    crc = CRC32_INIT_VALUE;
    crc = crc32_update(crc, data, len);
    return crc ^ CRC32_FINAL_XOR;
}

void crc32_compute_hex(const uint8_t *data, size_t len, char *hex_str)
{
    uint32_t crc;
    
    if (hex_str == NULL) {
        return;
    }
    
    crc = crc32_compute(data, len);
    crc32_to_hex(crc, hex_str);
}

int crc32_verify(const uint8_t *data, size_t len, uint32_t expected)
{
    uint32_t computed;
    
    computed = crc32_compute(data, len);
    return (computed == expected) ? 1 : 0;
}

int crc32_verify_hex(const uint8_t *data, size_t len, const char *expected_hex)
{
    uint32_t expected;
    
    if (expected_hex == NULL) {
        return 0;
    }
    
    expected = crc32_parse_hex(expected_hex);
    return crc32_verify(data, len, expected);
}

uint32_t crc32_parse_hex(const char *hex_str)
{
    uint32_t result = 0;
    int i;
    
    if (hex_str == NULL) {
        return 0;
    }
    
    for (i = 0; i < 8 && hex_str[i] != '\0'; i++) {
        char c = toupper(hex_str[i]);
        result <<= 4;
        if (c >= '0' && c <= '9') {
            result |= (c - '0');
        } else if (c >= 'A' && c <= 'F') {
            result |= (c - 'A' + 10);
        } else {
            return 0;
        }
    }
    
    return result;
}

void crc32_to_hex(uint32_t crc, char *hex_str)
{
    static const char hex_digits[] = "0123456789ABCDEF";
    int i;
    
    if (hex_str == NULL) {
        return;
    }
    
    for (i = 7; i >= 0; i--) {
        hex_str[7 - i] = hex_digits[(crc >> (i * 4)) & 0x0FU];
    }
    hex_str[8] = '\0';
}
