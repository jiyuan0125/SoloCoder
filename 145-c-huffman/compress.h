#ifndef COMPRESS_H
#define COMPRESS_H

#include "huffman.h"

typedef struct {
    FILE *file;
    uint8_t current_byte;
    uint8_t bit_count;
} BitWriter;

int bit_writer_init(BitWriter *writer, FILE *file);
int bit_writer_write_bit(BitWriter *writer, uint8_t bit);
int bit_writer_write_bits(BitWriter *writer, const uint8_t *bits, uint8_t length);
int bit_writer_flush(BitWriter *writer);

int write_file_header(FILE *out, const FrequencyTable *table);
int compress_file(const char *input_filename, const char *output_filename);
int compress_buffer(const uint8_t *input_buffer, size_t input_size, 
                    uint8_t **output_buffer, size_t *output_size);

#endif
