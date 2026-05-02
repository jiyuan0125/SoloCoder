#include "hashmap.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define LOAD_FACTOR_MAX 0.70
#define LOAD_FACTOR_MIN 0.20
#define MIN_CAPACITY 7

typedef enum {
    ENTRY_EMPTY,
    ENTRY_OCCUPIED,
    ENTRY_TOMBSTONE
} entry_state_t;

typedef struct {
    char* key;
    char* value;
    size_t probe_distance;
    entry_state_t state;
} entry_t;

struct hashmap {
    entry_t* buckets;
    size_t capacity;
    size_t size;
    size_t tombstone_count;
};

static int is_prime(size_t n) {
    if (n < 2) return 0;
    if (n == 2) return 1;
    if (n % 2 == 0) return 0;
    for (size_t i = 3; i * i <= n; i += 2) {
        if (n % i == 0) return 0;
    }
    return 1;
}

static size_t next_prime(size_t n) {
    if (n < 2) return 2;
    size_t candidate = n;
    while (1) {
        if (is_prime(candidate)) return candidate;
        candidate++;
    }
}

static size_t hash_func(const char* key) {
    size_t hash = 5381;
    int c;
    while ((c = *key++)) {
        hash = ((hash << 5) + hash) + c;
    }
    return hash;
}

static void entry_init(entry_t* entry) {
    entry->key = NULL;
    entry->value = NULL;
    entry->probe_distance = 0;
    entry->state = ENTRY_EMPTY;
}

static void entry_clear(entry_t* entry) {
    if (entry->key) {
        free(entry->key);
        entry->key = NULL;
    }
    if (entry->value) {
        free(entry->value);
        entry->value = NULL;
    }
    entry->probe_distance = 0;
    entry->state = ENTRY_EMPTY;
}

static void entry_set_tombstone(entry_t* entry) {
    if (entry->key) {
        free(entry->key);
        entry->key = NULL;
    }
    if (entry->value) {
        free(entry->value);
        entry->value = NULL;
    }
    entry->probe_distance = 0;
    entry->state = ENTRY_TOMBSTONE;
}

static void entry_swap(entry_t* a, entry_t* b) {
    entry_t temp = *a;
    *a = *b;
    *b = temp;
}

hashmap_t* hm_create(size_t capacity) {
    hashmap_t* map = (hashmap_t*)malloc(sizeof(hashmap_t));
    if (!map) return NULL;
    
    if (capacity < MIN_CAPACITY) {
        capacity = MIN_CAPACITY;
    }
    capacity = next_prime(capacity);
    
    map->buckets = (entry_t*)calloc(capacity, sizeof(entry_t));
    if (!map->buckets) {
        free(map);
        return NULL;
    }
    
    for (size_t i = 0; i < capacity; i++) {
        entry_init(&map->buckets[i]);
    }
    
    map->capacity = capacity;
    map->size = 0;
    map->tombstone_count = 0;
    
    return map;
}

void hm_destroy(hashmap_t* map) {
    if (!map) return;
    
    for (size_t i = 0; i < map->capacity; i++) {
        entry_clear(&map->buckets[i]);
    }
    free(map->buckets);
    free(map);
}

static int hm_resize(hashmap_t* map, size_t new_capacity) {
    if (new_capacity < MIN_CAPACITY) {
        new_capacity = MIN_CAPACITY;
    }
    new_capacity = next_prime(new_capacity);
    
    if (new_capacity == map->capacity) {
        return 0;
    }
    
    entry_t* new_buckets = (entry_t*)calloc(new_capacity, sizeof(entry_t));
    if (!new_buckets) {
        return -1;
    }
    
    for (size_t i = 0; i < new_capacity; i++) {
        entry_init(&new_buckets[i]);
    }
    
    entry_t* old_buckets = map->buckets;
    size_t old_capacity = map->capacity;
    
    map->buckets = new_buckets;
    map->capacity = new_capacity;
    map->size = 0;
    map->tombstone_count = 0;
    
    for (size_t i = 0; i < old_capacity; i++) {
        entry_t* entry = &old_buckets[i];
        if (entry->state == ENTRY_OCCUPIED) {
            size_t hash = hash_func(entry->key);
            size_t idx = hash % map->capacity;
            size_t dist = 0;
            
            while (1) {
                entry_t* curr = &map->buckets[idx];
                
                if (curr->state == ENTRY_EMPTY || curr->state == ENTRY_TOMBSTONE) {
                    curr->key = entry->key;
                    curr->value = entry->value;
                    curr->probe_distance = dist;
                    curr->state = ENTRY_OCCUPIED;
                    map->size++;
                    
                    entry->key = NULL;
                    entry->value = NULL;
                    break;
                }
                
                if (dist > curr->probe_distance) {
                    char* temp_key = entry->key;
                    char* temp_val = entry->value;
                    size_t temp_dist = dist;
                    
                    entry->key = curr->key;
                    entry->value = curr->value;
                    dist = curr->probe_distance;
                    
                    curr->key = temp_key;
                    curr->value = temp_val;
                    curr->probe_distance = temp_dist;
                }
                
                dist++;
                idx = (idx + 1) % map->capacity;
            }
        }
        entry_clear(&old_buckets[i]);
    }
    
    free(old_buckets);
    return 0;
}

int hm_put(hashmap_t* map, const char* key, const char* value) {
    if (!map || !key || !value) return -1;
    
    double load_factor = (double)(map->size + map->tombstone_count) / map->capacity;
    if (load_factor >= LOAD_FACTOR_MAX) {
        size_t new_cap = map->capacity * 2;
        if (hm_resize(map, new_cap) != 0) {
            return -1;
        }
    }
    
    size_t hash = hash_func(key);
    size_t idx = hash % map->capacity;
    
    char* new_key = strdup(key);
    char* new_value = strdup(value);
    if (!new_key || !new_value) {
        free(new_key);
        free(new_value);
        return -1;
    }
    
    entry_t inserting;
    inserting.key = new_key;
    inserting.value = new_value;
    inserting.probe_distance = 0;
    inserting.state = ENTRY_OCCUPIED;
    
    entry_t* current = &inserting;
    size_t current_dist = 0;
    size_t current_idx = idx;
    
    while (1) {
        entry_t* bucket = &map->buckets[current_idx];
        
        if (bucket->state == ENTRY_EMPTY) {
            bucket->key = current->key;
            bucket->value = current->value;
            bucket->probe_distance = current_dist;
            bucket->state = ENTRY_OCCUPIED;
            map->size++;
            if (current == &inserting) {
                current->key = NULL;
                current->value = NULL;
            }
            break;
        }
        
        if (bucket->state == ENTRY_TOMBSTONE) {
            bucket->key = current->key;
            bucket->value = current->value;
            bucket->probe_distance = current_dist;
            bucket->state = ENTRY_OCCUPIED;
            map->size++;
            map->tombstone_count--;
            if (current == &inserting) {
                current->key = NULL;
                current->value = NULL;
            }
            break;
        }
        
        if (bucket->state == ENTRY_OCCUPIED && strcmp(bucket->key, key) == 0) {
            free(bucket->value);
            bucket->value = strdup(value);
            free(current->key);
            free(current->value);
            return 0;
        }
        
        if (current_dist > bucket->probe_distance) {
            entry_swap(current, bucket);
            size_t temp_dist = current_dist;
            current_dist = bucket->probe_distance;
            bucket->probe_distance = temp_dist;
        }
        
        current_dist++;
        current_idx = (current_idx + 1) % map->capacity;
    }
    
    return 0;
}

int hm_get(const hashmap_t* map, const char* key, const char** out_value) {
    if (!map || !key) return -1;
    
    size_t hash = hash_func(key);
    size_t idx = hash % map->capacity;
    size_t dist = 0;
    
    while (dist < map->capacity) {
        entry_t* bucket = &map->buckets[idx];
        
        if (bucket->state == ENTRY_EMPTY) {
            break;
        }
        
        if (bucket->state == ENTRY_OCCUPIED) {
            if (strcmp(bucket->key, key) == 0) {
                if (out_value) {
                    *out_value = bucket->value;
                }
                return 0;
            }
        }
        
        dist++;
        idx = (idx + 1) % map->capacity;
    }
    
    return -1;
}

int hm_delete(hashmap_t* map, const char* key) {
    if (!map || !key) return -1;
    
    size_t hash = hash_func(key);
    size_t idx = hash % map->capacity;
    size_t dist = 0;
    
    while (dist < map->capacity) {
        entry_t* bucket = &map->buckets[idx];
        
        if (bucket->state == ENTRY_EMPTY) {
            break;
        }
        
        if (bucket->state == ENTRY_OCCUPIED && strcmp(bucket->key, key) == 0) {
            entry_set_tombstone(bucket);
            map->size--;
            map->tombstone_count++;
            
            if (map->capacity > MIN_CAPACITY) {
                double load_factor = (double)map->size / map->capacity;
                if (load_factor <= LOAD_FACTOR_MIN) {
                    size_t new_cap = map->capacity / 2;
                    hm_resize(map, new_cap);
                }
            }
            
            return 0;
        }
        
        dist++;
        idx = (idx + 1) % map->capacity;
    }
    
    return -1;
}

int hm_contains(const hashmap_t* map, const char* key) {
    return hm_get(map, key, NULL) == 0 ? 1 : 0;
}

size_t hm_size(const hashmap_t* map) {
    if (!map) return 0;
    return map->size;
}

size_t hm_capacity(const hashmap_t* map) {
    if (!map) return 0;
    return map->capacity;
}
