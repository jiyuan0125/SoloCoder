#ifndef CACHE_H
#define CACHE_H

#include <stddef.h>
#include <time.h>

#define KEY_MAX_LEN         256
#define VALUE_MAX_LEN       (1024 * 1024)
#define DEFAULT_CACHE_SIZE  (100 * 1024 * 1024)

typedef struct Cache Cache;
typedef struct CacheEntry CacheEntry;

typedef struct {
    size_t total_queries;
    size_t hits;
    size_t misses;
    size_t evictions;
    size_t current_size;
    size_t entry_count;
} CacheStats;

typedef struct {
    int is_hit;
    const char *key;
    const unsigned char *value;
    size_t value_len;
} BatchResultItem;

typedef struct {
    BatchResultItem *items;
    size_t hit_count;
    size_t miss_count;
    char **missed_keys;
    size_t total_count;
} BatchResult;

Cache* cache_create(size_t capacity);
void cache_destroy(Cache *cache);

int cache_set(Cache *cache, const char *key, const unsigned char *value, 
              size_t value_len, time_t ttl_seconds);
int cache_get(Cache *cache, const char *key, unsigned char **value, 
              size_t *value_len);
int cache_delete(Cache *cache, const char *key);
void cache_clear(Cache *cache);

BatchResult* cache_batch_get(Cache *cache, const char **keys, size_t key_count);
void batch_result_free(BatchResult *result);

void cache_get_stats(Cache *cache, CacheStats *stats);
void cache_reset_stats(Cache *cache);
double cache_get_hit_rate(Cache *cache);

#endif
