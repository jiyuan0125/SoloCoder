#ifndef CACHE_CORE_H
#define CACHE_CORE_H

#include "cache.h"
#include <pthread.h>

struct CacheEntry {
    char key[KEY_MAX_LEN + 1];
    unsigned char *value;
    size_t value_len;
    time_t expire_time;
    size_t size;
    CacheEntry *next;
    CacheEntry *prev;
    CacheEntry *hash_next;
};

typedef struct {
    CacheEntry **buckets;
    size_t size;
    size_t count;
} HashTable;

struct CacheCore {
    HashTable table;
    CacheEntry *lru_head;
    CacheEntry *lru_tail;
    size_t capacity;
    size_t current_size;
    size_t entry_count;
};

typedef struct CacheCore CacheCore;

CacheCore* cache_core_create(size_t capacity);
void cache_core_destroy(CacheCore *core);

int cache_core_set(CacheCore *core, const char *key, const unsigned char *value,
                   size_t value_len, time_t expire_time);
int cache_core_get(CacheCore *core, const char *key, unsigned char **value,
                   size_t *value_len);
int cache_core_delete(CacheCore *core, const char *key);
void cache_core_clear(CacheCore *core);

int cache_core_evict_if_needed(CacheCore *core, size_t needed_size, size_t *evicted_count);
int cache_core_entry_expired(CacheCore *core, CacheEntry *entry);
void cache_core_remove_expired(CacheCore *core, size_t *removed_count);

CacheEntry* cache_core_find(CacheCore *core, const char *key);
void cache_core_move_to_head(CacheCore *core, CacheEntry *entry);
void cache_core_remove_entry(CacheCore *core, CacheEntry *entry);

#endif
