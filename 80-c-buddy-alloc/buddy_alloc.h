#ifndef BUDDY_ALLOC_H
#define BUDDY_ALLOC_H

#include <stddef.h>

#define MIN_BLOCK_SIZE      ((size_t)64)
#define MAX_BLOCK_SIZE      ((size_t)4 * 1024 * 1024)
#define TOTAL_MEMORY        ((size_t)64 * 1024 * 1024)
#define NUM_LEVELS          17

typedef struct buddy_allocator buddy_allocator_t;

typedef enum {
    BUDDY_OK = 0,
    BUDDY_ERROR_INVALID_PTR = -1,
    BUDDY_ERROR_NOT_ALIGNED = -2,
    BUDDY_ERROR_OUT_OF_MEMORY = -3,
    BUDDY_ERROR_ALREADY_FREE = -4
} buddy_error_t;

buddy_allocator_t* buddy_init(void);
void buddy_destroy(buddy_allocator_t* allocator);

void* buddy_alloc(buddy_allocator_t* allocator, size_t size);
buddy_error_t buddy_free(buddy_allocator_t* allocator, void* ptr);

void buddy_stats(buddy_allocator_t* allocator);

#endif
