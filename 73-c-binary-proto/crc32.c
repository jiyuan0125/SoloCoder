#include "proto.h"
#include <stdint.h>

#define CRC32_POLYNOMIAL 0xEDB88320

static uint32_t crc32_table[256];
static int table_initialized = 0;

static void crc32_init_table(void) {
    if (table_initialized) {
        return;
    }

    for (uint32_t i = 0; i < 256; i++) {
        uint32_t crc = i;
        for (uint32_t j = 0; j < 8; j++) {
            if (crc & 1) {
                crc = (crc >> 1) ^ CRC32_POLYNOMIAL;
            } else {
                crc >>= 1;
            }
        }
        crc32_table[i] = crc;
    }

    table_initialized = 1;
}

uint32_t crc32_ieee(const uint8_t *data, size_t len) {
    crc32_init_table();

    uint32_t crc = 0xFFFFFFFF;

    for (size_t i = 0; i < len; i++) {
        uint8_t byte = data[i];
        crc = crc32_table[(crc ^ byte) & 0xFF] ^ (crc >> 8);
    }

    return crc ^ 0xFFFFFFFF;
}
