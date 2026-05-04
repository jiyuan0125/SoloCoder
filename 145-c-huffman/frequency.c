#include "frequency.h"
#include <string.h>

#define BUFFER_SIZE (1024 * 1024)

int frequency_init(FrequencyTable *table) {
    if (table == NULL) {
        return -1;
    }
    memset(table->frequencies, 0, sizeof(table->frequencies));
    table->total_bytes = 0;
    table->total_bits = 0;
    table->non_zero_count = 0;
    return 0;
}

int frequency_count_file(FrequencyTable *table, const char *filename) {
    if (table == NULL || filename == NULL) {
        return -1;
    }
    
    FILE *file = fopen(filename, "rb");
    if (file == NULL) {
        return -1;
    }
    
    uint8_t *buffer = (uint8_t *)malloc(BUFFER_SIZE);
    if (buffer == NULL) {
        fclose(file);
        return -1;
    }
    
    size_t bytes_read;
    while ((bytes_read = fread(buffer, 1, BUFFER_SIZE, file)) > 0) {
        for (size_t i = 0; i < bytes_read; i++) {
            table->frequencies[buffer[i]]++;
        }
        table->total_bytes += bytes_read;
    }
    
    free(buffer);
    fclose(file);
    
    table->non_zero_count = 0;
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table->frequencies[i] > 0) {
            table->non_zero_count++;
        }
    }
    
    return 0;
}

int frequency_count_buffer(FrequencyTable *table, const uint8_t *buffer, size_t size) {
    if (table == NULL || buffer == NULL) {
        return -1;
    }
    
    frequency_init(table);
    
    for (size_t i = 0; i < size; i++) {
        table->frequencies[buffer[i]]++;
    }
    table->total_bytes = size;
    
    table->non_zero_count = 0;
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table->frequencies[i] > 0) {
            table->non_zero_count++;
        }
    }
    
    return 0;
}

typedef struct {
    HuffmanNode **nodes;
    int size;
    int capacity;
} MinHeap;

static MinHeap *min_heap_create(int capacity) {
    MinHeap *heap = (MinHeap *)malloc(sizeof(MinHeap));
    if (heap == NULL) {
        return NULL;
    }
    heap->nodes = (HuffmanNode **)malloc(sizeof(HuffmanNode *) * capacity);
    if (heap->nodes == NULL) {
        free(heap);
        return NULL;
    }
    heap->size = 0;
    heap->capacity = capacity;
    return heap;
}

static void min_heap_free(MinHeap *heap) {
    if (heap == NULL) {
        return;
    }
    free(heap->nodes);
    free(heap);
}

static void min_heap_swap(HuffmanNode **a, HuffmanNode **b) {
    HuffmanNode *temp = *a;
    *a = *b;
    *b = temp;
}

static void min_heap_heapify(MinHeap *heap, int index) {
    int smallest = index;
    int left = 2 * index + 1;
    int right = 2 * index + 2;
    
    if (left < heap->size && 
        heap->nodes[left]->frequency < heap->nodes[smallest]->frequency) {
        smallest = left;
    }
    
    if (right < heap->size && 
        heap->nodes[right]->frequency < heap->nodes[smallest]->frequency) {
        smallest = right;
    }
    
    if (smallest != index) {
        min_heap_swap(&heap->nodes[index], &heap->nodes[smallest]);
        min_heap_heapify(heap, smallest);
    }
}

static int min_heap_insert(MinHeap *heap, HuffmanNode *node) {
    if (heap == NULL || node == NULL) {
        return -1;
    }
    
    if (heap->size >= heap->capacity) {
        return -1;
    }
    
    int i = heap->size;
    heap->nodes[i] = node;
    heap->size++;
    
    while (i != 0 && 
           heap->nodes[(i - 1) / 2]->frequency > heap->nodes[i]->frequency) {
        min_heap_swap(&heap->nodes[i], &heap->nodes[(i - 1) / 2]);
        i = (i - 1) / 2;
    }
    
    return 0;
}

static HuffmanNode *min_heap_extract_min(MinHeap *heap) {
    if (heap == NULL || heap->size <= 0) {
        return NULL;
    }
    
    if (heap->size == 1) {
        heap->size--;
        return heap->nodes[0];
    }
    
    HuffmanNode *root = heap->nodes[0];
    heap->nodes[0] = heap->nodes[heap->size - 1];
    heap->size--;
    min_heap_heapify(heap, 0);
    
    return root;
}

HuffmanNode *huffman_build_tree(const FrequencyTable *table) {
    if (table == NULL) {
        return NULL;
    }
    
    if (table->non_zero_count == 0) {
        return NULL;
    }
    
    if (table->non_zero_count == 1) {
        uint8_t single_byte = 0;
        for (int i = 0; i < BYTE_COUNT; i++) {
            if (table->frequencies[i] > 0) {
                single_byte = (uint8_t)i;
                break;
            }
        }
        return huffman_create_node(single_byte, table->frequencies[single_byte], NULL, NULL);
    }
    
    MinHeap *heap = min_heap_create(table->non_zero_count);
    if (heap == NULL) {
        return NULL;
    }
    
    for (int i = 0; i < BYTE_COUNT; i++) {
        if (table->frequencies[i] > 0) {
            HuffmanNode *node = huffman_create_node((uint8_t)i, table->frequencies[i], NULL, NULL);
            if (node == NULL) {
                for (int j = 0; j < heap->size; j++) {
                    free(heap->nodes[j]);
                }
                min_heap_free(heap);
                return NULL;
            }
            min_heap_insert(heap, node);
        }
    }
    
    while (heap->size > 1) {
        HuffmanNode *left = min_heap_extract_min(heap);
        HuffmanNode *right = min_heap_extract_min(heap);
        
        if (left == NULL || right == NULL) {
            for (int j = 0; j < heap->size; j++) {
                free(heap->nodes[j]);
            }
            if (left != NULL) free(left);
            if (right != NULL) free(right);
            min_heap_free(heap);
            return NULL;
        }
        
        HuffmanNode *parent = huffman_create_node(0, left->frequency + right->frequency, left, right);
        if (parent == NULL) {
            for (int j = 0; j < heap->size; j++) {
                free(heap->nodes[j]);
            }
            free(left);
            free(right);
            min_heap_free(heap);
            return NULL;
        }
        
        min_heap_insert(heap, parent);
    }
    
    HuffmanNode *root = min_heap_extract_min(heap);
    min_heap_free(heap);
    
    return root;
}

void huffman_init_codes(HuffmanCode codes[BYTE_COUNT]) {
    if (codes == NULL) {
        return;
    }
    for (int i = 0; i < BYTE_COUNT; i++) {
        codes[i].length = 0;
        memset(codes[i].code, 0, MAX_CODE_LENGTH);
    }
}

void huffman_build_codes(const HuffmanNode *root, HuffmanCode codes[BYTE_COUNT], 
                         uint8_t *current_code, uint8_t current_length) {
    if (root == NULL || codes == NULL) {
        return;
    }
    
    if (root->left == NULL && root->right == NULL) {
        codes[root->byte].length = current_length;
        memcpy(codes[root->byte].code, current_code, current_length);
        return;
    }
    
    if (root->left != NULL) {
        current_code[current_length] = 0;
        huffman_build_codes(root->left, codes, current_code, current_length + 1);
    }
    
    if (root->right != NULL) {
        current_code[current_length] = 1;
        huffman_build_codes(root->right, codes, current_code, current_length + 1);
    }
}
