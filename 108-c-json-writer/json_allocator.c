#include "json_allocator.h"
#include <stdlib.h>
#include <string.h>

static JsonAllocator default_allocator = {
    malloc,
    realloc,
    free
};

static JsonAllocator *current_allocator = &default_allocator;

void json_allocator_set_default(JsonAllocator *allocator) {
    if (allocator != NULL &&
        allocator->alloc != NULL &&
        allocator->realloc != NULL &&
        allocator->free != NULL) {
        current_allocator = allocator;
    }
}

JsonAllocator *json_allocator_get_default(void) {
    return current_allocator;
}

void *json_malloc(size_t size) {
    if (current_allocator == NULL || current_allocator->alloc == NULL) {
        return NULL;
    }
    return current_allocator->alloc(size);
}

void *json_realloc(void *ptr, size_t size) {
    if (current_allocator == NULL || current_allocator->realloc == NULL) {
        return NULL;
    }
    return current_allocator->realloc(ptr, size);
}

void json_free_ptr(void *ptr) {
    if (current_allocator == NULL || current_allocator->free == NULL) {
        return;
    }
    if (ptr != NULL) {
        current_allocator->free(ptr);
    }
}

char *json_strdup(const char *s) {
    if (s == NULL) return NULL;
    size_t len = strlen(s);
    char *dup = (char *)json_malloc(len + 1);
    if (dup == NULL) return NULL;
    strcpy(dup, s);
    return dup;
}
