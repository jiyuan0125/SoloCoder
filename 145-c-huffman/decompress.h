#ifndef DECOMPRESS_H
#define DECOMPRESS_H

#include "huffman.h"

typedef struct {
    FILE *file;
    uint8_t current_byte;
    uint8_t bit_count;
} BitReader;

int bit_reader_init(BitReader *reader, FILE *file);
int bit_reader_read_bit(BitReader *reader, uint8_t *bit);
int bit_reader_read_bits(BitReader *reader, uint8_t *bits, uint8_t length);

int read_file_header(FILE *in, FrequencyTable *table);
int decompress_file(const char *input_filename, const char *output_filename);
int decompress_buffer(const uint8_t *input_buffer, size_t input_size, 
                      uint8_t **output_buffer, size_t *output_size);

#endif
