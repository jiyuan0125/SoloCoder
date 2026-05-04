#ifndef HUFFMAN_H
#define HUFFMAN_H

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#define MAX_CODE_LENGTH 256
#define BYTE_COUNT 256

typedef struct HuffmanNode {
    uint8_t byte;
    uint64_t frequency;
    struct HuffmanNode *left;
    struct HuffmanNode *right;
} HuffmanNode;

typedef struct {
    uint8_t code[MAX_CODE_LENGTH];
    uint8_t length;
} HuffmanCode;

typedef struct {
    uint64_t frequencies[BYTE_COUNT];
    uint64_t total_bytes;
    uint64_t total_bits;
    uint8_t non_zero_count;
} FrequencyTable;

typedef struct {
    uint64_t total_bits;
    uint8_t non_zero_count;
} FileHeader;

HuffmanNode *huffman_create_node(uint8_t byte, uint64_t frequency, 
                                  HuffmanNode *left, HuffmanNode *right);
void huffman_free_tree(HuffmanNode *root);

#endif
