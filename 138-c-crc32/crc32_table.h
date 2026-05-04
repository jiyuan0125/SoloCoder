#ifndef CRC32_TABLE_H
#define CRC32_TABLE_H

#include <stdint.h>
#include <stddef.h>

#define CRC32_POLYNOMIAL 0xEDB88320U

extern const uint32_t *crc32_table;

void crc32_table_init(void);
int crc32_table_is_initialized(void);

#endif
