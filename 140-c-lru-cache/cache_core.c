#include "cache_core.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define HASH_TABLE_SIZE 1024

static unsigned int hash_string(const char *str) {
    unsigned int hash = 5381;
    int c;
    while ((c = *str++)) {
        hash = ((hash << 5) + hash) + c;
    }
    return hash;
}

static int hash_table_init(HashTable *table, size_t size) {
    table->buckets = (CacheEntry **)calloc(size, sizeof(CacheEntry *));
    if (!table->buckets) return -1;
    table->size = size;
    table->count = 0;
    return 0;
}

static void hash_table_destroy(HashTable *table) {
    if (table->buckets) {
        free(table->buckets);
        table->buckets = NULL;
    }
    table->size = 0;
    table->count = 0;
}

static CacheEntry* hash_table_find(HashTable *table, const char *key) {
    unsigned int idx = hash_string(key) % table->size;
    CacheEntry *entry = table->buckets[idx];
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            return entry;
        }
        entry = entry->hash_next;
    }
    return NULL;
}

static void hash_table_insert(HashTable *table, CacheEntry *entry) {
    unsigned int idx = hash_string(entry->key) % table->size;
    entry->hash_next = table->buckets[idx];
    table->buckets[idx] = entry;
    table->count++;
}

static void hash_table_remove(HashTable *table, const char *key) {
    unsigned int idx = hash_string(key) % table->size;
    CacheEntry **prev = &table->buckets[idx];
    CacheEntry *entry = table->buckets[idx];
    
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            *prev = entry->hash_next;
            entry->hash_next = NULL;
            table->count--;
            return;
        }
        prev = &entry->hash_next;
        entry = entry->hash_next;
    }
}

CacheCore* cache_core_create(size_t capacity) {
    CacheCore *core = (CacheCore *)malloc(sizeof(CacheCore));
    if (!core) return NULL;
    
    if (hash_table_init(&core->table, HASH_TABLE_SIZE) != 0) {
        free(core);
        return NULL;
    }
    
    core->lru_head = NULL;
    core->lru_tail = NULL;
    core->capacity = capacity;
    core->current_size = 0;
    core->entry_count = 0;
    
    return core;
}

void cache_core_destroy(CacheCore *core) {
    if (!core) return;
    cache_core_clear(core);
    hash_table_destroy(&core->table);
    free(core);
}

static void cache_entry_free(CacheEntry *entry) {
    if (entry) {
        if (entry->value) {
            free(entry->value);
        }
        free(entry);
    }
}

void cache_core_move_to_head(CacheCore *core, CacheEntry *entry) {
    if (!core || !entry) return;
    
    if (entry == core->lru_head) return;
    
    if (entry->prev) entry->prev->next = entry->next;
    if (entry->next) entry->next->prev = entry->prev;
    
    if (entry == core->lru_tail) core->lru_tail = entry->prev;
    
    entry->prev = NULL;
    entry->next = core->lru_head;
    if (core->lru_head) core->lru_head->prev = entry;
    core->lru_head = entry;
    
    if (!core->lru_tail) core->lru_tail = entry;
}

void cache_core_remove_entry(CacheCore *core, CacheEntry *entry) {
    if (!core || !entry) return;
    
    if (entry->prev) entry->prev->next = entry->next;
    if (entry->next) entry->next->prev = entry->prev;
    
    if (entry == core->lru_head) core->lru_head = entry->next;
    if (entry == core->lru_tail) core->lru_tail = entry->prev;
    
    entry->prev = NULL;
    entry->next = NULL;
}

CacheEntry* cache_core_find(CacheCore *core, const char *key) {
    if (!core || !key) return NULL;
    return hash_table_find(&core->table, key);
}

int cache_core_entry_expired(CacheCore *core, CacheEntry *entry) {
    (void)core;
    if (!entry) return 0;
    if (entry->expire_time == 0) return 0;
    time_t now = time(NULL);
    return now >= entry->expire_time;
}

void cache_core_remove_expired(CacheCore *core, size_t *removed_count) {
    if (!core) return;
    if (removed_count) *removed_count = 0;
    
    CacheEntry *entry = core->lru_tail;
    while (entry) {
        CacheEntry *prev = entry->prev;
        if (cache_core_entry_expired(core, entry)) {
            hash_table_remove(&core->table, entry->key);
            cache_core_remove_entry(core, entry);
            core->current_size -= entry->size;
            core->entry_count--;
            if (removed_count) (*removed_count)++;
            cache_entry_free(entry);
        }
        entry = prev;
    }
}

int cache_core_evict_if_needed(CacheCore *core, size_t needed_size, size_t *evicted_count) {
    if (!core) return -1;
    if (evicted_count) *evicted_count = 0;
    
    if (core->current_size + needed_size <= core->capacity) return 0;
    
    while (core->current_size + needed_size > core->capacity && core->lru_tail != NULL) {
        CacheEntry *entry = core->lru_tail;
        
        hash_table_remove(&core->table, entry->key);
        cache_core_remove_entry(core, entry);
        
        core->current_size -= entry->size;
        core->entry_count--;
        if (evicted_count) (*evicted_count)++;
        
        cache_entry_free(entry);
    }
    
    return 0;
}

int cache_core_set(CacheCore *core, const char *key, const unsigned char *value,
                   size_t value_len, time_t expire_time, size_t *evicted_count) {
    if (!core || !key || !value || value_len == 0) return -1;
    if (evicted_count) *evicted_count = 0;
    
    size_t key_len = strlen(key);
    if (key_len > KEY_MAX_LEN) return -1;
    if (value_len > VALUE_MAX_LEN) return -1;
    
    size_t entry_size = key_len + 1 + value_len + sizeof(CacheEntry);
    size_t total_evicted = 0;
    
    CacheEntry *existing = cache_core_find(core, key);
    if (existing) {
        if (cache_core_entry_expired(core, existing)) {
            cache_core_delete(core, key);
        } else {
            hash_table_remove(&core->table, key);
            core->current_size -= existing->size;
            core->entry_count--;
            
            unsigned char *new_value = (unsigned char *)realloc(existing->value, value_len);
            if (!new_value) {
                hash_table_insert(&core->table, existing);
                core->current_size += existing->size;
                core->entry_count++;
                return -1;
            }
            
            size_t evicted = 0;
            if (entry_size > existing->size) {
                size_t additional_needed = entry_size - existing->size;
                if (cache_core_evict_if_needed(core, additional_needed, &evicted) != 0) {
                    hash_table_insert(&core->table, existing);
                    core->current_size += existing->size;
                    core->entry_count++;
                    existing->value = new_value;
                    return -1;
                }
                total_evicted += evicted;
            }
            
            memcpy(new_value, value, value_len);
            existing->value = new_value;
            existing->value_len = value_len;
            existing->expire_time = expire_time;
            existing->size = entry_size;
            
            hash_table_insert(&core->table, existing);
            core->current_size += entry_size;
            core->entry_count++;
            cache_core_move_to_head(core, existing);
            
            if (evicted_count) *evicted_count = total_evicted;
            return 0;
        }
    }
    
    if (entry_size > core->capacity) return -1;
    
    size_t evicted = 0;
    if (cache_core_evict_if_needed(core, entry_size, &evicted) != 0) {
        return -1;
    }
    total_evicted += evicted;
    
    CacheEntry *entry = (CacheEntry *)malloc(sizeof(CacheEntry));
    if (!entry) return -1;
    
    strncpy(entry->key, key, KEY_MAX_LEN);
    entry->key[KEY_MAX_LEN] = '\0';
    
    entry->value = (unsigned char *)malloc(value_len);
    if (!entry->value) {
        free(entry);
        return -1;
    }
    memcpy(entry->value, value, value_len);
    entry->value_len = value_len;
    entry->expire_time = expire_time;
    entry->size = entry_size;
    entry->next = NULL;
    entry->prev = NULL;
    entry->hash_next = NULL;
    
    hash_table_insert(&core->table, entry);
    core->current_size += entry_size;
    core->entry_count++;
    
    entry->next = core->lru_head;
    if (core->lru_head) core->lru_head->prev = entry;
    core->lru_head = entry;
    
    if (!core->lru_tail) core->lru_tail = entry;
    
    if (evicted_count) *evicted_count = total_evicted;
    return 0;
}

int cache_core_get(CacheCore *core, const char *key, unsigned char **value,
                   size_t *value_len) {
    if (!core || !key || !value || !value_len) return -1;
    
    CacheEntry *entry = cache_core_find(core, key);
    if (!entry) return 0;
    
    if (cache_core_entry_expired(core, entry)) {
        hash_table_remove(&core->table, key);
        cache_core_remove_entry(core, entry);
        core->current_size -= entry->size;
        core->entry_count--;
        cache_entry_free(entry);
        return 0;
    }
    
    cache_core_move_to_head(core, entry);
    
    *value = (unsigned char *)malloc(entry->value_len);
    if (!*value) return -1;
    memcpy(*value, entry->value, entry->value_len);
    *value_len = entry->value_len;
    
    return 1;
}

int cache_core_delete(CacheCore *core, const char *key) {
    if (!core || !key) return -1;
    
    CacheEntry *entry = hash_table_find(&core->table, key);
    if (!entry) return 0;
    
    hash_table_remove(&core->table, key);
    cache_core_remove_entry(core, entry);
    core->current_size -= entry->size;
    core->entry_count--;
    cache_entry_free(entry);
    
    return 1;
}

void cache_core_clear(CacheCore *core) {
    if (!core) return;
    
    CacheEntry *entry = core->lru_head;
    while (entry) {
        CacheEntry *next = entry->next;
        cache_entry_free(entry);
        entry = next;
    }
    
    memset(core->table.buckets, 0, core->table.size * sizeof(CacheEntry *));
    core->table.count = 0;
    core->lru_head = NULL;
    core->lru_tail = NULL;
    core->current_size = 0;
    core->entry_count = 0;
}
