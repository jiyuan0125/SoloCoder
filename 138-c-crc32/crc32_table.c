#include "crc32_table.h"
#include <stdlib.h>

static uint32_t g_crc32_table[256];
static volatile int g_table_initialized = 0;

const uint32_t *crc32_table = (const uint32_t *)g_crc32_table;

void crc32_table_init(void)
{
    int i, j;
    uint32_t crc;
    
    if (g_table_initialized) {
        return;
    }
    
    for (i = 0; i < 256; i++) {
        crc = (uint32_t)i;
        for (j = 0; j < 8; j++) {
            if (crc & 1U) {
                crc = (crc >> 1) ^ CRC32_POLYNOMIAL;
            } else {
                crc >>= 1;
            }
        }
        g_crc32_table[i] = crc;
    }
    
    g_table_initialized = 1;
}

int crc32_table_is_initialized(void)
{
    return g_table_initialized;
}
