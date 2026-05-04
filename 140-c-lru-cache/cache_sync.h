#ifndef CACHE_SYNC_H
#define CACHE_SYNC_H

#include "cache.h"
#include "cache_core.h"
#include <pthread.h>
#include <sys/types.h>

#define CACHE_SEGMENTS 16

typedef struct {
    CacheCore *core;
    pthread_rwlock_t rwlock;
} CacheSegment;

struct Cache {
    CacheSegment segments[CACHE_SEGMENTS];
    size_t capacity;
    size_t segment_capacity;
    pthread_mutex_t stats_mutex;
    CacheStats stats;
    pthread_t cleanup_thread;
    volatile int cleanup_running;
};

static inline unsigned int cache_key_hash(const char *key) {
    unsigned int hash = 5381;
    int c;
    while ((c = *key++)) {
        hash = ((hash << 5) + hash) + c;
    }
    return hash;
}

static inline int cache_get_segment_idx(const char *key) {
    return cache_key_hash(key) % CACHE_SEGMENTS;
}

int cache_sync_init(Cache *cache, size_t capacity);
void cache_sync_destroy(Cache *cache);

int cache_sync_read_lock(Cache *cache, int segment_idx);
int cache_sync_read_unlock(Cache *cache, int segment_idx);
int cache_sync_write_lock(Cache *cache, int segment_idx);
int cache_sync_write_unlock(Cache *cache, int segment_idx);

int cache_sync_write_lock_all(Cache *cache);
int cache_sync_write_unlock_all(Cache *cache);

void cache_sync_stats_lock(Cache *cache);
void cache_sync_stats_unlock(Cache *cache);

void cache_sync_increment_total(Cache *cache);
void cache_sync_increment_hits(Cache *cache);
void cache_sync_increment_misses(Cache *cache);
void cache_sync_increment_evictions(Cache *cache, size_t count);
void cache_sync_update_size(Cache *cache, ssize_t delta);
void cache_sync_update_entry_count(Cache *cache, ssize_t delta);

#endif
