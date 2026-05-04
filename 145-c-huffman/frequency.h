#ifndef FREQUENCY_H
#define FREQUENCY_H

#include "huffman.h"

int frequency_init(FrequencyTable *table);
int frequency_count_file(FrequencyTable *table, const char *filename);
int frequency_count_buffer(FrequencyTable *table, const uint8_t *buffer, size_t size);
HuffmanNode *huffman_build_tree(const FrequencyTable *table);
void huffman_build_codes(const HuffmanNode *root, HuffmanCode codes[BYTE_COUNT], 
                         uint8_t *current_code, uint8_t current_length);
void huffman_init_codes(HuffmanCode codes[BYTE_COUNT]);

#endif
