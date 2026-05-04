#ifndef CRC32_H
#define CRC32_H

#include <stdint.h>
#include <stddef.h>

#define CRC32_INIT_VALUE 0xFFFFFFFFU
#define CRC32_FINAL_XOR 0xFFFFFFFFU
#define CRC32_HEX_STRING_LENGTH 9

uint32_t crc32_compute(const uint8_t *data, size_t len);
void crc32_compute_hex(const uint8_t *data, size_t len, char *hex_str);

int crc32_verify(const uint8_t *data, size_t len, uint32_t expected);
int crc32_verify_hex(const uint8_t *data, size_t len, const char *expected_hex);

uint32_t crc32_parse_hex(const char *hex_str);
void crc32_to_hex(uint32_t crc, char *hex_str);

#endif
