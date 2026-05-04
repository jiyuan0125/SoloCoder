#include "cache.h"
#include "cache_sync.h"
#include "cache_core.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <unistd.h>

#define CLEANUP_INTERVAL_SECONDS 1

static void* cleanup_thread_func(void *arg) {
    Cache *cache = (Cache *)arg;
    
    while (cache->cleanup_running) {
        sleep(CLEANUP_INTERVAL_SECONDS);
        
        if (!cache->cleanup_running) break;
        
        for (int i = 0; i < CACHE_SEGMENTS; i++) {
            if (cache_sync_write_lock(cache, i) == 0) {
                size_t removed = 0;
                CacheCore *core = cache->segments[i].core;
                if (core) {
                    cache_core_remove_expired(core, &removed);
                }
                cache_sync_write_unlock(cache, i);
            }
        }
    }
    
    return NULL;
}

Cache* cache_create(size_t capacity) {
    if (capacity == 0) capacity = DEFAULT_CACHE_SIZE;
    
    Cache *cache = (Cache *)malloc(sizeof(Cache));
    if (!cache) return NULL;
    
    if (cache_sync_init(cache, capacity) != 0) {
        free(cache);
        return NULL;
    }
    
    cache->cleanup_running = 1;
    if (pthread_create(&cache->cleanup_thread, NULL, cleanup_thread_func, cache) != 0) {
        cache->cleanup_running = 0;
        cache_sync_destroy(cache);
        free(cache);
        return NULL;
    }
    
    return cache;
}

void cache_destroy(Cache *cache) {
    if (!cache) return;
    
    cache->cleanup_running = 0;
    pthread_join(cache->cleanup_thread, NULL);
    
    cache_sync_destroy(cache);
    free(cache);
}

int cache_set(Cache *cache, const char *key, const unsigned char *value,
              size_t value_len, time_t ttl_seconds) {
    if (!cache || !key || !value || value_len == 0) return -1;
    
    int segment_idx = cache_get_segment_idx(key);
    time_t expire_time = (ttl_seconds > 0) ? (time(NULL) + ttl_seconds) : 0;
    size_t evicted = 0;
    
    if (cache_sync_write_lock(cache, segment_idx) != 0) {
        return -1;
    }
    
    CacheCore *core = cache->segments[segment_idx].core;
    if (!core) {
        cache_sync_write_unlock(cache, segment_idx);
        return -1;
    }
    
    int ret = cache_core_set(core, key, value, value_len, expire_time, &evicted);
    
    cache_sync_write_unlock(cache, segment_idx);
    
    if (evicted > 0) {
        cache_sync_increment_evictions(cache, evicted);
    }
    
    return ret;
}

int cache_get(Cache *cache, const char *key, unsigned char **value,
              size_t *value_len) {
    if (!cache || !key || !value || !value_len) return -1;
    
    *value = NULL;
    *value_len = 0;
    
    cache_sync_increment_total(cache);
    
    int segment_idx = cache_get_segment_idx(key);
    
    if (cache_sync_write_lock(cache, segment_idx) != 0) {
        cache_sync_increment_misses(cache);
        return -1;
    }
    
    CacheCore *core = cache->segments[segment_idx].core;
    if (!core) {
        cache_sync_write_unlock(cache, segment_idx);
        cache_sync_increment_misses(cache);
        return -1;
    }
    
    int ret = cache_core_get(core, key, value, value_len);
    
    cache_sync_write_unlock(cache, segment_idx);
    
    if (ret == 1) {
        cache_sync_increment_hits(cache);
    } else if (ret == 0) {
        cache_sync_increment_misses(cache);
    }
    
    return ret;
}

int cache_delete(Cache *cache, const char *key) {
    if (!cache || !key) return -1;
    
    int segment_idx = cache_get_segment_idx(key);
    
    if (cache_sync_write_lock(cache, segment_idx) != 0) {
        return -1;
    }
    
    CacheCore *core = cache->segments[segment_idx].core;
    if (!core) {
        cache_sync_write_unlock(cache, segment_idx);
        return -1;
    }
    
    int ret = cache_core_delete(core, key);
    
    cache_sync_write_unlock(cache, segment_idx);
    
    return ret;
}

void cache_clear(Cache *cache) {
    if (!cache) return;
    
    cache_sync_write_lock_all(cache);
    
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        CacheCore *core = cache->segments[i].core;
        if (core) {
            cache_core_clear(core);
        }
    }
    
    cache_sync_stats_lock(cache);
    cache->stats.current_size = 0;
    cache->stats.entry_count = 0;
    cache_sync_stats_unlock(cache);
    
    cache_sync_write_unlock_all(cache);
}

BatchResult* cache_batch_get(Cache *cache, const char **keys, size_t key_count) {
    if (!cache || !keys || key_count == 0) return NULL;
    
    BatchResult *result = (BatchResult *)malloc(sizeof(BatchResult));
    if (!result) return NULL;
    
    result->items = (BatchResultItem *)malloc(sizeof(BatchResultItem) * key_count);
    if (!result->items) {
        free(result);
        return NULL;
    }
    result->total_count = key_count;
    result->hit_count = 0;
    result->miss_count = 0;
    
    int *segments_needed = (int *)calloc(CACHE_SEGMENTS, sizeof(int));
    if (!segments_needed) {
        free(result->items);
        free(result);
        return NULL;
    }
    
    for (size_t i = 0; i < key_count; i++) {
        int seg_idx = cache_get_segment_idx(keys[i]);
        segments_needed[seg_idx] = 1;
    }
    
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        if (segments_needed[i]) {
            cache_sync_write_lock(cache, i);
        }
    }
    
    for (size_t i = 0; i < key_count; i++) {
        cache_sync_increment_total(cache);
        
        const char *key = keys[i];
        int segment_idx = cache_get_segment_idx(key);
        
        result->items[i].key = key;
        result->items[i].is_hit = 0;
        result->items[i].value = NULL;
        result->items[i].value_len = 0;
        
        CacheCore *core = cache->segments[segment_idx].core;
        if (!core) {
            cache_sync_increment_misses(cache);
            result->miss_count++;
            continue;
        }
        
        unsigned char *value = NULL;
        size_t value_len = 0;
        int ret = cache_core_get(core, key, &value, &value_len);
        
        if (ret == 1 && value != NULL) {
            result->items[i].is_hit = 1;
            result->items[i].value = value;
            result->items[i].value_len = value_len;
            result->hit_count++;
            cache_sync_increment_hits(cache);
        } else {
            cache_sync_increment_misses(cache);
            result->miss_count++;
        }
    }
    
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        if (segments_needed[i]) {
            cache_sync_write_unlock(cache, i);
        }
    }
    
    result->missed_keys = NULL;
    if (result->miss_count > 0) {
        result->missed_keys = (char **)malloc(sizeof(char *) * result->miss_count);
        if (result->missed_keys) {
            size_t miss_idx = 0;
            for (size_t i = 0; i < key_count; i++) {
                if (!result->items[i].is_hit) {
                    result->missed_keys[miss_idx++] = (char *)result->items[i].key;
                }
            }
        }
    }
    
    free(segments_needed);
    
    return result;
}

void batch_result_free(BatchResult *result) {
    if (!result) return;
    
    if (result->items) {
        for (size_t i = 0; i < result->total_count; i++) {
            if (result->items[i].value) {
                free((void *)result->items[i].value);
            }
        }
        free(result->items);
    }
    
    if (result->missed_keys) {
        free(result->missed_keys);
    }
    
    free(result);
}

void cache_get_stats(Cache *cache, CacheStats *stats) {
    if (!cache || !stats) return;
    
    cache_sync_stats_lock(cache);
    *stats = cache->stats;
    cache_sync_stats_unlock(cache);
    
    stats->current_size = 0;
    stats->entry_count = 0;
    
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        if (cache_sync_read_lock(cache, i) == 0) {
            CacheCore *core = cache->segments[i].core;
            if (core) {
                stats->current_size += core->current_size;
                stats->entry_count += core->entry_count;
            }
            cache_sync_read_unlock(cache, i);
        }
    }
}

void cache_reset_stats(Cache *cache) {
    if (!cache) return;
    
    cache_sync_stats_lock(cache);
    cache->stats.total_queries = 0;
    cache->stats.hits = 0;
    cache->stats.misses = 0;
    cache->stats.evictions = 0;
    cache_sync_stats_unlock(cache);
}

double cache_get_hit_rate(Cache *cache) {
    if (!cache) return 0.0;
    
    cache_sync_stats_lock(cache);
    size_t total = cache->stats.total_queries;
    size_t hits = cache->stats.hits;
    cache_sync_stats_unlock(cache);
    
    if (total == 0) return 0.0;
    return (double)hits / (double)total;
}
