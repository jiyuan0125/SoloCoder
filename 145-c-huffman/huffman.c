#include "huffman.h"

HuffmanNode *huffman_create_node(uint8_t byte, uint64_t frequency, 
                                  HuffmanNode *left, HuffmanNode *right) {
    HuffmanNode *node = (HuffmanNode *)malloc(sizeof(HuffmanNode));
    if (node == NULL) {
        return NULL;
    }
    node->byte = byte;
    node->frequency = frequency;
    node->left = left;
    node->right = right;
    return node;
}

void huffman_free_tree(HuffmanNode *root) {
    if (root == NULL) {
        return;
    }
    huffman_free_tree(root->left);
    huffman_free_tree(root->right);
    free(root);
}
