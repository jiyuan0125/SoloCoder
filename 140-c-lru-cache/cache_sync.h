#ifndef CACHE_SYNC_H
#define CACHE_SYNC_H

#include "cache.h"
#include "cache_core.h"
#include <pthread.h>
#include <sys/types.h>

struct Cache {
    CacheCore *core;
    pthread_rwlock_t rwlock;
    size_t capacity;
    pthread_mutex_t stats_mutex;
    CacheStats stats;
    pthread_t cleanup_thread;
    volatile int cleanup_running;
};

int cache_sync_init(Cache *cache, size_t capacity);
void cache_sync_destroy(Cache *cache);

int cache_sync_read_lock(Cache *cache);
int cache_sync_read_unlock(Cache *cache);
int cache_sync_write_lock(Cache *cache);
int cache_sync_write_unlock(Cache *cache);

void cache_sync_stats_lock(Cache *cache);
void cache_sync_stats_unlock(Cache *cache);

void cache_sync_increment_total(Cache *cache);
void cache_sync_increment_hits(Cache *cache);
void cache_sync_increment_misses(Cache *cache);
void cache_sync_increment_evictions(Cache *cache, size_t count);
void cache_sync_update_size(Cache *cache, ssize_t delta);
void cache_sync_update_entry_count(Cache *cache, ssize_t delta);

#endif
