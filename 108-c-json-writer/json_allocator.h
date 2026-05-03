#ifndef JSON_ALLOCATOR_H
#define JSON_ALLOCATOR_H

#include <stddef.h>

typedef void *(*JsonAllocFn)(size_t size);
typedef void *(*JsonReallocFn)(void *ptr, size_t size);
typedef void (*JsonFreeFn)(void *ptr);

typedef struct {
    JsonAllocFn alloc;
    JsonReallocFn realloc;
    JsonFreeFn free;
} JsonAllocator;

void json_allocator_set_default(JsonAllocator *allocator);
JsonAllocator *json_allocator_get_default(void);

void *json_malloc(size_t size);
void *json_realloc(void *ptr, size_t size);
void json_free_ptr(void *ptr);

char *json_strdup(const char *s);

#endif
