#ifndef HASHMAP_H
#define HASHMAP_H

#include <stddef.h>

typedef struct hashmap hashmap_t;

hashmap_t* hm_create(size_t capacity);
void hm_destroy(hashmap_t* map);

int hm_put(hashmap_t* map, const char* key, const char* value);
int hm_get(const hashmap_t* map, const char* key, const char** out_value);
int hm_delete(hashmap_t* map, const char* key);
int hm_contains(const hashmap_t* map, const char* key);

size_t hm_size(const hashmap_t* map);
size_t hm_capacity(const hashmap_t* map);

#endif
