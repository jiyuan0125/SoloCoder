#include "decompress.h"
#include "frequency.h"
#include <string.h>

#define BUFFER_SIZE (1024 * 1024)

int bit_reader_init(BitReader *reader, FILE *file) {
    if (reader == NULL || file == NULL) {
        return -1;
    }
    reader->file = file;
    reader->current_byte = 0;
    reader->bit_count = 0;
    return 0;
}

int bit_reader_read_bit(BitReader *reader, uint8_t *bit) {
    if (reader == NULL || bit == NULL) {
        return -1;
    }
    
    if (reader->bit_count == 0) {
        if (fread(&reader->current_byte, 1, 1, reader->file) != 1) {
            return -1;
        }
        reader->bit_count = 8;
    }
    
    reader->bit_count--;
    *bit = (reader->current_byte >> reader->bit_count) & 0x01;
    
    return 0;
}

int bit_reader_read_bits(BitReader *reader, uint8_t *bits, uint8_t length) {
    if (reader == NULL || bits == NULL || length == 0) {
        return -1;
    }
    
    for (uint8_t i = 0; i < length; i++) {
        if (bit_reader_read_bit(reader, &bits[i]) != 0) {
            return -1;
        }
    }
    
    return 0;
}

static uint64_t read_varint(FILE *in) {
    uint64_t value = 0;
    uint8_t shift = 0;
    uint8_t byte;
    
    do {
        if (fread(&byte, 1, 1, in) != 1) {
            return 0;
        }
        value |= (uint64_t)(byte & 0x7F) << shift;
        shift += 7;
    } while (byte & 0x80);
    
    return value;
}

static uint64_t read_varint_from_buffer(const uint8_t *buffer, size_t *offset) {
    uint64_t value = 0;
    uint8_t shift = 0;
    uint8_t byte;
    
    do {
        byte = buffer[*offset];
        (*offset)++;
        value |= (uint64_t)(byte & 0x7F) << shift;
        shift += 7;
    } while (byte & 0x80);
    
    return value;
}

int read_file_header(FILE *in, FrequencyTable *table) {
    if (in == NULL || table == NULL) {
        return -1;
    }
    
    frequency_init(table);
    
    if (fread(&table->non_zero_count, 1, 1, in) != 1) {
        return -1;
    }
    
    if (table->non_zero_count == 0) {
        table->total_bits = 0;
        table->total_bytes = 0;
        return 0;
    }
    
    for (int i = 0; i < table->non_zero_count; i++) {
        uint8_t byte;
        if (fread(&byte, 1, 1, in) != 1) {
            return -1;
        }
        table->frequencies[byte] = read_varint(in);
        table->total_bytes += table->frequencies[byte];
    }
    
    table->total_bits = read_varint(in);
    
    return 0;
}

int decompress_file(const char *input_filename, const char *output_filename) {
    if (input_filename == NULL || output_filename == NULL) {
        return -1;
    }
    
    FILE *in = fopen(input_filename, "rb");
    if (in == NULL) {
        return -1;
    }
    
    FrequencyTable table;
    if (read_file_header(in, &table) != 0) {
        fclose(in);
        return -1;
    }
    
    if (table.non_zero_count == 0) {
        FILE *out = fopen(output_filename, "wb");
        if (out == NULL) {
            fclose(in);
            return -1;
        }
        fclose(out);
        fclose(in);
        return 0;
    }
    
    HuffmanNode *root = huffman_build_tree(&table);
    if (root == NULL) {
        fclose(in);
        return -1;
    }
    
    FILE *out = fopen(output_filename, "wb");
    if (out == NULL) {
        huffman_free_tree(root);
        fclose(in);
        return -1;
    }
    
    if (table.non_zero_count == 1) {
        uint8_t single_byte = 0;
        for (int i = 0; i < BYTE_COUNT; i++) {
            if (table.frequencies[i] > 0) {
                single_byte = (uint8_t)i;
                break;
            }
        }
        
        uint8_t *buffer = (uint8_t *)malloc(BUFFER_SIZE);
        if (buffer == NULL) {
            fclose(in);
            fclose(out);
            huffman_free_tree(root);
            return -1;
        }
        
        memset(buffer, single_byte, BUFFER_SIZE);
        
        uint64_t remaining = table.total_bytes;
        while (remaining > 0) {
            size_t to_write = (remaining > BUFFER_SIZE) ? BUFFER_SIZE : (size_t)remaining;
            if (fwrite(buffer, 1, to_write, out) != to_write) {
                free(buffer);
                fclose(in);
                fclose(out);
                huffman_free_tree(root);
                return -1;
            }
            remaining -= to_write;
        }
        
        free(buffer);
    } else {
        BitReader bit_reader;
        if (bit_reader_init(&bit_reader, in) != 0) {
            fclose(in);
            fclose(out);
            huffman_free_tree(root);
            return -1;
        }
        
        uint8_t *output_buffer = (uint8_t *)malloc(BUFFER_SIZE);
        if (output_buffer == NULL) {
            fclose(in);
            fclose(out);
            huffman_free_tree(root);
            return -1;
        }
        
        size_t buffer_offset = 0;
        HuffmanNode *current = root;
        uint64_t bits_read = 0;
        
        while (bits_read < table.total_bits) {
            uint8_t bit;
            if (bit_reader_read_bit(&bit_reader, &bit) != 0) {
                free(output_buffer);
                fclose(in);
                fclose(out);
                huffman_free_tree(root);
                return -1;
            }
            bits_read++;
            
            if (bit == 0) {
                current = current->left;
            } else {
                current = current->right;
            }
            
            if (current == NULL) {
                free(output_buffer);
                fclose(in);
                fclose(out);
                huffman_free_tree(root);
                return -1;
            }
            
            if (current->left == NULL && current->right == NULL) {
                output_buffer[buffer_offset++] = current->byte;
                current = root;
                
                if (buffer_offset >= BUFFER_SIZE) {
                    if (fwrite(output_buffer, 1, BUFFER_SIZE, out) != BUFFER_SIZE) {
                        free(output_buffer);
                        fclose(in);
                        fclose(out);
                        huffman_free_tree(root);
                        return -1;
                    }
                    buffer_offset = 0;
                }
            }
        }
        
        if (buffer_offset > 0) {
            if (fwrite(output_buffer, 1, buffer_offset, out) != buffer_offset) {
                free(output_buffer);
                fclose(in);
                fclose(out);
                huffman_free_tree(root);
                return -1;
            }
        }
        
        free(output_buffer);
    }
    
    fclose(in);
    fclose(out);
    huffman_free_tree(root);
    
    return 0;
}

int decompress_buffer(const uint8_t *input_buffer, size_t input_size, 
                      uint8_t **output_buffer, size_t *output_size) {
    if (input_buffer == NULL || output_buffer == NULL || output_size == NULL || input_size == 0) {
        return -1;
    }
    
    size_t offset = 0;
    FrequencyTable table;
    frequency_init(&table);
    
    table.non_zero_count = input_buffer[offset++];
    
    if (table.non_zero_count == 0) {
        *output_buffer = NULL;
        *output_size = 0;
        return 0;
    }
    
    for (int i = 0; i < table.non_zero_count; i++) {
        uint8_t byte = input_buffer[offset++];
        table.frequencies[byte] = read_varint_from_buffer(input_buffer, &offset);
        table.total_bytes += table.frequencies[byte];
    }
    
    table.total_bits = read_varint_from_buffer(input_buffer, &offset);
    
    HuffmanNode *root = huffman_build_tree(&table);
    if (root == NULL) {
        return -1;
    }
    
    *output_buffer = (uint8_t *)malloc(table.total_bytes);
    if (*output_buffer == NULL) {
        huffman_free_tree(root);
        return -1;
    }
    *output_size = table.total_bytes;
    
    if (table.non_zero_count == 1) {
        uint8_t single_byte = 0;
        for (int i = 0; i < BYTE_COUNT; i++) {
            if (table.frequencies[i] > 0) {
                single_byte = (uint8_t)i;
                break;
            }
        }
        memset(*output_buffer, single_byte, table.total_bytes);
    } else {
        HuffmanNode *current = root;
        size_t output_offset = 0;
        uint64_t bits_read = 0;
        
        uint8_t current_byte = 0;
        uint8_t bit_count = 0;
        size_t byte_offset = offset;
        
        while (bits_read < table.total_bits && output_offset < table.total_bytes) {
            if (bit_count == 0) {
                if (byte_offset >= input_size) {
                    free(*output_buffer);
                    *output_buffer = NULL;
                    huffman_free_tree(root);
                    return -1;
                }
                current_byte = input_buffer[byte_offset++];
                bit_count = 8;
            }
            
            bit_count--;
            uint8_t bit = (current_byte >> bit_count) & 0x01;
            bits_read++;
            
            if (bit == 0) {
                current = current->left;
            } else {
                current = current->right;
            }
            
            if (current == NULL) {
                free(*output_buffer);
                *output_buffer = NULL;
                huffman_free_tree(root);
                return -1;
            }
            
            if (current->left == NULL && current->right == NULL) {
                (*output_buffer)[output_offset++] = current->byte;
                current = root;
            }
        }
    }
    
    huffman_free_tree(root);
    return 0;
}
