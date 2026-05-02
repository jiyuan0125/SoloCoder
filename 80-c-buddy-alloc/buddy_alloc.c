#include "buddy_alloc.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#define LEVEL_TO_SIZE(level)    ((size_t)MIN_BLOCK_SIZE << (level))

#define MIN_BLOCKS_PER_MAX      (MAX_BLOCK_SIZE / MIN_BLOCK_SIZE)
#define TOTAL_MIN_BLOCKS        (TOTAL_MEMORY / MIN_BLOCK_SIZE)

typedef struct block {
    struct block* next;
    struct block* prev;
} block_t;

struct buddy_allocator {
    uint8_t* memory_base;
    block_t* free_lists[NUM_LEVELS];
    uint8_t* block_level;
    uint8_t* block_allocated;
    size_t allocated_blocks;
    size_t used_memory;
    size_t free_blocks_count[NUM_LEVELS];
};

static size_t round_up_to_power_of_two(size_t size) {
    if (size == 0) return 0;
    size--;
    size |= size >> 1;
    size |= size >> 2;
    size |= size >> 4;
    size |= size >> 8;
    size |= size >> 16;
#if defined(__x86_64__) || defined(__aarch64__)
    size |= size >> 32;
#endif
    size++;
    return size;
}

static size_t get_required_level(size_t size) {
    if (size == 0) return NUM_LEVELS;
    
    size_t rounded = round_up_to_power_of_two(size);
    if (rounded < MIN_BLOCK_SIZE) {
        rounded = MIN_BLOCK_SIZE;
    }
    if (rounded > MAX_BLOCK_SIZE) {
        return NUM_LEVELS;
    }
    
    size_t level = 0;
    size_t current = MIN_BLOCK_SIZE;
    while (current < rounded) {
        current <<= 1;
        level++;
    }
    
    return level;
}

static void remove_from_free_list(buddy_allocator_t* allocator, block_t* block, size_t level) {
    if (block->prev != NULL) {
        block->prev->next = block->next;
    } else {
        allocator->free_lists[level] = block->next;
    }
    if (block->next != NULL) {
        block->next->prev = block->prev;
    }
    
    block->next = NULL;
    block->prev = NULL;
    allocator->free_blocks_count[level]--;
}

static void add_to_free_list(buddy_allocator_t* allocator, block_t* block, size_t level) {
    block->next = allocator->free_lists[level];
    block->prev = NULL;
    if (allocator->free_lists[level] != NULL) {
        allocator->free_lists[level]->prev = block;
    }
    allocator->free_lists[level] = block;
    allocator->free_blocks_count[level]++;
}

static block_t* index_to_block(buddy_allocator_t* allocator, size_t min_block_index) {
    return (block_t*)(allocator->memory_base + min_block_index * MIN_BLOCK_SIZE);
}

static size_t block_to_index(buddy_allocator_t* allocator, block_t* block) {
    return ((uintptr_t)block - (uintptr_t)allocator->memory_base) / MIN_BLOCK_SIZE;
}

static size_t get_buddy_index(size_t index, size_t level) {
    size_t blocks_per_level = ((size_t)1) << level;
    return index ^ blocks_per_level;
}

static void split_block(buddy_allocator_t* allocator, size_t index, 
                         size_t current_level, size_t target_level) {
    while (current_level > target_level) {
        size_t lower_level = current_level - 1;
        size_t blocks_per_lower = ((size_t)1) << lower_level;
        size_t buddy_index = index + blocks_per_lower;
        
        block_t* block = index_to_block(allocator, index);
        block_t* buddy = index_to_block(allocator, buddy_index);
        
        remove_from_free_list(allocator, block, current_level);
        
        allocator->block_level[index] = (uint8_t)lower_level;
        allocator->block_level[buddy_index] = (uint8_t)lower_level;
        allocator->block_allocated[index] = 0;
        allocator->block_allocated[buddy_index] = 0;
        
        add_to_free_list(allocator, block, lower_level);
        add_to_free_list(allocator, buddy, lower_level);
        
        current_level = lower_level;
    }
}

buddy_allocator_t* buddy_init(void) {
    buddy_allocator_t* allocator = (buddy_allocator_t*)malloc(sizeof(buddy_allocator_t));
    if (allocator == NULL) {
        return NULL;
    }
    
    if (posix_memalign((void**)&allocator->memory_base, MAX_BLOCK_SIZE, TOTAL_MEMORY) != 0) {
        free(allocator);
        return NULL;
    }
    
    allocator->block_level = (uint8_t*)malloc(TOTAL_MIN_BLOCKS * sizeof(uint8_t));
    if (allocator->block_level == NULL) {
        free(allocator->memory_base);
        free(allocator);
        return NULL;
    }
    
    allocator->block_allocated = (uint8_t*)malloc(TOTAL_MIN_BLOCKS * sizeof(uint8_t));
    if (allocator->block_allocated == NULL) {
        free(allocator->block_level);
        free(allocator->memory_base);
        free(allocator);
        return NULL;
    }
    
    memset(allocator->memory_base, 0, TOTAL_MEMORY);
    memset(allocator->free_lists, 0, sizeof(allocator->free_lists));
    memset(allocator->free_blocks_count, 0, sizeof(allocator->free_blocks_count));
    memset(allocator->block_level, 0, TOTAL_MIN_BLOCKS * sizeof(uint8_t));
    memset(allocator->block_allocated, 0, TOTAL_MIN_BLOCKS * sizeof(uint8_t));
    
    allocator->allocated_blocks = 0;
    allocator->used_memory = 0;
    
    size_t max_level = NUM_LEVELS - 1;
    size_t blocks_per_max = ((size_t)1) << max_level;
    size_t num_max_blocks = TOTAL_MIN_BLOCKS / blocks_per_max;
    
    for (size_t i = 0; i < num_max_blocks; i++) {
        size_t index = i * blocks_per_max;
        block_t* block = index_to_block(allocator, index);
        allocator->block_level[index] = (uint8_t)max_level;
        allocator->block_allocated[index] = 0;
        add_to_free_list(allocator, block, max_level);
    }
    
    return allocator;
}

void buddy_destroy(buddy_allocator_t* allocator) {
    if (allocator == NULL) return;
    
    if (allocator->block_allocated != NULL) {
        free(allocator->block_allocated);
    }
    if (allocator->block_level != NULL) {
        free(allocator->block_level);
    }
    if (allocator->memory_base != NULL) {
        free(allocator->memory_base);
    }
    free(allocator);
}

void* buddy_alloc(buddy_allocator_t* allocator, size_t size) {
    if (allocator == NULL || size == 0) {
        return NULL;
    }
    
    size_t target_level = get_required_level(size);
    if (target_level >= NUM_LEVELS) {
        return NULL;
    }
    
    size_t current_level = target_level;
    block_t* block = NULL;
    size_t found_index = 0;
    
    while (current_level < NUM_LEVELS) {
        if (allocator->free_lists[current_level] != NULL) {
            block = allocator->free_lists[current_level];
            found_index = block_to_index(allocator, block);
            break;
        }
        current_level++;
    }
    
    if (block == NULL) {
        return NULL;
    }
    
    if (current_level > target_level) {
        split_block(allocator, found_index, current_level, target_level);
        block = index_to_block(allocator, found_index);
    }
    
    remove_from_free_list(allocator, block, target_level);
    allocator->block_level[found_index] = (uint8_t)target_level;
    allocator->block_allocated[found_index] = 1;
    
    allocator->allocated_blocks++;
    allocator->used_memory += LEVEL_TO_SIZE(target_level);
    
    return (void*)block;
}

buddy_error_t buddy_free(buddy_allocator_t* allocator, void* ptr) {
    if (allocator == NULL || ptr == NULL) {
        return BUDDY_ERROR_INVALID_PTR;
    }
    
    uintptr_t ptr_addr = (uintptr_t)ptr;
    uintptr_t base_addr = (uintptr_t)allocator->memory_base;
    
    if (ptr_addr < base_addr || ptr_addr >= base_addr + TOTAL_MEMORY) {
        return BUDDY_ERROR_INVALID_PTR;
    }
    
    size_t index = block_to_index(allocator, (block_t*)ptr);
    
    if (index >= TOTAL_MIN_BLOCKS) {
        return BUDDY_ERROR_NOT_ALIGNED;
    }
    
    if (allocator->block_allocated[index] == 0) {
        return BUDDY_ERROR_ALREADY_FREE;
    }
    
    size_t level = allocator->block_level[index];
    if (level >= NUM_LEVELS) {
        return BUDDY_ERROR_NOT_ALIGNED;
    }
    
    size_t block_size = LEVEL_TO_SIZE(level);
    if ((ptr_addr & (block_size - 1)) != 0) {
        return BUDDY_ERROR_NOT_ALIGNED;
    }
    
    size_t current_level = level;
    size_t current_index = index;
    
    while (current_level < NUM_LEVELS - 1) {
        size_t buddy_index = get_buddy_index(current_index, current_level);
        
        if (buddy_index >= TOTAL_MIN_BLOCKS) {
            break;
        }
        
        if (allocator->block_allocated[buddy_index] != 0) {
            break;
        }
        
        if (allocator->block_level[buddy_index] != current_level) {
            break;
        }
        
        block_t* buddy_block = index_to_block(allocator, buddy_index);
        remove_from_free_list(allocator, buddy_block, current_level);
        
        if (buddy_index < current_index) {
            current_index = buddy_index;
        }
        
        current_level++;
    }
    
    block_t* block = index_to_block(allocator, current_index);
    allocator->block_level[current_index] = (uint8_t)current_level;
    allocator->block_allocated[current_index] = 0;
    add_to_free_list(allocator, block, current_level);
    
    allocator->allocated_blocks--;
    allocator->used_memory -= LEVEL_TO_SIZE(level);
    
    return BUDDY_OK;
}

void buddy_stats(buddy_allocator_t* allocator) {
    if (allocator == NULL) {
        printf("Allocator is NULL\n");
        return;
    }
    
    size_t free_memory = 0;
    for (size_t i = 0; i < NUM_LEVELS; i++) {
        free_memory += allocator->free_blocks_count[i] * LEVEL_TO_SIZE(i);
    }
    
    double fragmentation_rate = 0.0;
    if (TOTAL_MEMORY > 0) {
        fragmentation_rate = ((double)allocator->allocated_blocks * MIN_BLOCK_SIZE * 100.0) / TOTAL_MEMORY;
    }
    
    printf("========================================\n");
    printf("         Buddy Allocator Stats          \n");
    printf("========================================\n");
    printf("Total Memory:      %zu bytes (%.2f MB)\n", TOTAL_MEMORY, (double)TOTAL_MEMORY / (1024 * 1024));
    printf("Used Memory:       %zu bytes (%.2f MB)\n", allocator->used_memory, (double)allocator->used_memory / (1024 * 1024));
    printf("Free Memory:       %zu bytes (%.2f MB)\n", free_memory, (double)free_memory / (1024 * 1024));
    printf("Fragmentation Rate: %.2f%%\n", fragmentation_rate);
    printf("Allocated Blocks:  %zu\n", allocator->allocated_blocks);
    printf("\nFree Blocks by Level:\n");
    printf("Level   Block Size   Free Count\n");
    printf("--------------------------------\n");
    
    for (size_t i = 0; i < NUM_LEVELS; i++) {
        size_t size = LEVEL_TO_SIZE(i);
        const char* unit;
        double display_size;
        
        if (size >= 1024 * 1024) {
            display_size = (double)size / (1024 * 1024);
            unit = "MB";
        } else if (size >= 1024) {
            display_size = (double)size / 1024;
            unit = "KB";
        } else {
            display_size = (double)size;
            unit = "B";
        }
        
        printf("%2zu      %6.0f %-2s       %zu\n", i, display_size, unit, allocator->free_blocks_count[i]);
    }
    printf("========================================\n");
}
