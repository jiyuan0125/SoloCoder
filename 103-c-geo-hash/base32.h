#ifndef BASE32_H
#define BASE32_H

#include <stdint.h>

#define GEOHASH_BASE32_CHARS "0123456789bcdefghjkmnpqrstuvwxyz"
#define BASE32_PRECISION_MAX 12

int base32_char_to_index(char c);
char base32_index_to_char(int index);
uint64_t base32_decode(const char *hash, int precision);
void base32_encode(uint64_t value, int precision, char *result);

#endif
