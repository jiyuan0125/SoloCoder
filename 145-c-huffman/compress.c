#include "compress.h"
#include "frequency.h"
#include <string.h>

#define BUFFER_SIZE (1024 * 1024)

int bit_writer_init(BitWriter *writer, FILE *file) {
    if (writer == NULL || file == NULL) {
        return -1;
    }
    writer->file = file;
    writer->current_byte = 0;
    writer->bit_count = 0;
    return 0;
}

int bit_writer_write_bit(BitWriter *writer, uint8_t bit) {
    if (writer == NULL) {
        return -1;
    }
    
    writer->current_byte = (writer->current_byte << 1) | (bit & 0x01);
    writer->bit_count++;
    
    if (writer->bit_count == 8) {
        if (fwrite(&writer->current_byte, 1, 1, writer->file) != 1) {
            return -1;
        }
        writer->current_byte = 0;
        writer->bit_count = 0;
    }
    
    return 0;
}

int bit_writer_write_bits(BitWriter *writer, const uint8_t *bits, uint8_t length) {
    if (writer == NULL || bits == NULL || length == 0) {
        return -1;
    }
    
    for (uint8_t i = 0; i < length; i++) {
        if (bit_writer_write_bit(writer, bits[i]) != 0) {
            return -1;
        }
    }
    
    return 0;
}

int bit_writer_flush(BitWriter *writer) {
    if (writer == NULL) {
        return -1;
    }
    
    if (writer->bit_count > 0) {
        writer->current_byte <<= (8 - writer->bit_count);
        if (fwrite(&writer->current_byte, 1, 1, writer->file) != 1) {
            return -1;
        }
        writer->current_byte = 0;
        writer->bit_count = 0;
    }
    
    return 0;
}

static size_t write_varint(uint64_t value, uint8_t *buffer) {
    size_t size = 0;
    do {
        uint8_t byte = value & 0x7F;
        value >>= 7;
        if (value != 0) {
            byte |= 0x80;
        }
        buffer[size++] = byte;
    } while (value != 0);
    return size;
}

int write_file_header(FILE *out, const FrequencyTable *table) {
    if (out == NULL || table == NULL) {
        return -1;
    }
    
    uint8_t byte_val = table->non_zero_count;
    if (fwrite(&byte_val, 1, 1, out) != 1) {
        return -1;
    }
    
    uint8_t varint_buf[16];
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table->frequencies[i] > 0) {
            byte_val = (uint8_t)i;
            if (fwrite(&byte_val, 1, 1, out) != 1) {
                return -1;
            }
            size_t varint_len = write_varint(table->frequencies[i], varint_buf);
            if (fwrite(varint_buf, 1, varint_len, out) != varint_len) {
                return -1;
            }
        }
    }
    
    size_t varint_len = write_varint(table->total_bits, varint_buf);
    if (fwrite(varint_buf, 1, varint_len, out) != varint_len) {
        return -1;
    }
    
    return 0;
}

static uint64_t calculate_total_bits(const FrequencyTable *table, const HuffmanCode codes[BYTE_COUNT]) {
    if (table == NULL || codes == NULL) {
        return 0;
    }
    
    uint64_t total = 0;
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table->frequencies[i] > 0) {
            total += table->frequencies[i] * codes[i].length;
        }
    }
    return total;
}

int compress_file(const char *input_filename, const char *output_filename) {
    if (input_filename == NULL || output_filename == NULL) {
        return -1;
    }
    
    FrequencyTable table;
    if (frequency_init(&table) != 0) {
        return -1;
    }
    
    if (frequency_count_file(&table, input_filename) != 0) {
        return -1;
    }
    
    if (table.total_bytes == 0) {
        FILE *out = fopen(output_filename, "wb");
        if (out == NULL) {
            return -1;
        }
        uint8_t zero_header = 0;
        fwrite(&zero_header, 1, 1, out);
        fclose(out);
        return 0;
    }
    
    HuffmanNode *root = huffman_build_tree(&table);
    if (root == NULL) {
        return -1;
    }
    
    HuffmanCode codes[BYTE_COUNT];
    huffman_init_codes(codes);
    
    if (table.non_zero_count == 1) {
        for (int i = 0; i < BYTE_COUNT; i++) {
            if (table.frequencies[i] > 0) {
                codes[i].length = 0;
                break;
            }
        }
        table.total_bits = table.total_bytes;
    } else {
        uint8_t current_code[MAX_CODE_LENGTH];
        huffman_build_codes(root, codes, current_code, 0);
        table.total_bits = calculate_total_bits(&table, codes);
    }
    
    FILE *in = fopen(input_filename, "rb");
    if (in == NULL) {
        huffman_free_tree(root);
        return -1;
    }
    
    FILE *out = fopen(output_filename, "wb");
    if (out == NULL) {
        fclose(in);
        huffman_free_tree(root);
        return -1;
    }
    
    if (write_file_header(out, &table) != 0) {
        fclose(in);
        fclose(out);
        huffman_free_tree(root);
        return -1;
    }
    
    BitWriter bit_writer;
    if (bit_writer_init(&bit_writer, out) != 0) {
        fclose(in);
        fclose(out);
        huffman_free_tree(root);
        return -1;
    }
    
    uint8_t *buffer = (uint8_t *)malloc(BUFFER_SIZE);
    if (buffer == NULL) {
        fclose(in);
        fclose(out);
        huffman_free_tree(root);
        return -1;
    }
    
    size_t bytes_read;
    if (table.non_zero_count == 1) {
        uint8_t single_byte = 0;
        for (int i = 0; i < BYTE_COUNT; i++) {
            if (table.frequencies[i] > 0) {
                single_byte = (uint8_t)i;
                break;
            }
        }
        while ((bytes_read = fread(buffer, 1, BUFFER_SIZE, in)) > 0) {
            for (size_t i = 0; i < bytes_read; i++) {
                if (buffer[i] != single_byte) {
                    free(buffer);
                    fclose(in);
                    fclose(out);
                    huffman_free_tree(root);
                    return -1;
                }
            }
        }
    } else {
        while ((bytes_read = fread(buffer, 1, BUFFER_SIZE, in)) > 0) {
            for (size_t i = 0; i < bytes_read; i++) {
                uint8_t byte = buffer[i];
                if (codes[byte].length > 0) {
                    if (bit_writer_write_bits(&bit_writer, codes[byte].code, codes[byte].length) != 0) {
                        free(buffer);
                        fclose(in);
                        fclose(out);
                        huffman_free_tree(root);
                        return -1;
                    }
                }
            }
        }
    }
    
    if (bit_writer_flush(&bit_writer) != 0) {
        free(buffer);
        fclose(in);
        fclose(out);
        huffman_free_tree(root);
        return -1;
    }
    
    free(buffer);
    fclose(in);
    fclose(out);
    huffman_free_tree(root);
    
    return 0;
}

int compress_buffer(const uint8_t *input_buffer, size_t input_size, 
                    uint8_t **output_buffer, size_t *output_size) {
    if (input_buffer == NULL || output_buffer == NULL || output_size == NULL) {
        return -1;
    }
    
    FrequencyTable table;
    if (frequency_init(&table) != 0) {
        return -1;
    }
    
    if (frequency_count_buffer(&table, input_buffer, input_size) != 0) {
        return -1;
    }
    
    if (table.total_bytes == 0) {
        *output_buffer = (uint8_t *)malloc(1);
        if (*output_buffer == NULL) {
            return -1;
        }
        (*output_buffer)[0] = 0;
        *output_size = 1;
        return 0;
    }
    
    HuffmanNode *root = huffman_build_tree(&table);
    if (root == NULL) {
        return -1;
    }
    
    HuffmanCode codes[BYTE_COUNT];
    huffman_init_codes(codes);
    
    if (table.non_zero_count == 1) {
        for (int i = 0; i < BYTE_COUNT; i++) {
            if (table.frequencies[i] > 0) {
                codes[i].length = 0;
                break;
            }
        }
        table.total_bits = table.total_bytes;
    } else {
        uint8_t current_code[MAX_CODE_LENGTH];
        huffman_build_codes(root, codes, current_code, 0);
        table.total_bits = calculate_total_bits(&table, codes);
    }
    
    size_t header_size = 1;
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table.frequencies[i] > 0) {
            header_size += 1;
            uint64_t freq = table.frequencies[i];
            do {
                header_size++;
                freq >>= 7;
            } while (freq > 0);
        }
    }
    uint64_t tb = table.total_bits;
    do {
        header_size++;
        tb >>= 7;
    } while (tb > 0);
    
    size_t max_output_size = input_size + header_size + 8;
    *output_buffer = (uint8_t *)malloc(max_output_size);
    if (*output_buffer == NULL) {
        huffman_free_tree(root);
        return -1;
    }
    
    size_t offset = 0;
    
    (*output_buffer)[offset++] = table.non_zero_count;
    
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table.frequencies[i] > 0) {
            (*output_buffer)[offset++] = (uint8_t)i;
            offset += write_varint(table.frequencies[i], &(*output_buffer)[offset]);
        }
    }
    
    offset += write_varint(table.total_bits, &(*output_buffer)[offset]);
    
    if (table.non_zero_count != 1) {
        uint8_t current_byte = 0;
        uint8_t bit_count = 0;
        
        for (size_t i = 0; i < input_size; i++) {
            uint8_t byte = input_buffer[i];
            for (uint8_t j = 0; j < codes[byte].length; j++) {
                current_byte = (current_byte << 1) | codes[byte].code[j];
                bit_count++;
                
                if (bit_count == 8) {
                    (*output_buffer)[offset++] = current_byte;
                    current_byte = 0;
                    bit_count = 0;
                }
            }
        }
        
        if (bit_count > 0) {
            current_byte <<= (8 - bit_count);
            (*output_buffer)[offset++] = current_byte;
        }
    }
    
    *output_size = offset;
    huffman_free_tree(root);
    
    return 0;
}
