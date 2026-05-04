#include "cache_sync.h"
#include <stdlib.h>
#include <string.h>

int cache_sync_init(Cache *cache, size_t capacity) {
    if (!cache) return -1;
    
    cache->capacity = capacity;
    cache->segment_capacity = capacity / CACHE_SEGMENTS;
    if (cache->segment_capacity == 0) {
        cache->segment_capacity = 1024 * 1024;
    }
    
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        cache->segments[i].core = cache_core_create(cache->segment_capacity);
        if (!cache->segments[i].core) {
            for (int j = 0; j < i; j++) {
                cache_core_destroy(cache->segments[j].core);
                pthread_rwlock_destroy(&cache->segments[j].rwlock);
            }
            return -1;
        }
        
        if (pthread_rwlock_init(&cache->segments[i].rwlock, NULL) != 0) {
            for (int j = 0; j <= i; j++) {
                if (cache->segments[j].core) {
                    cache_core_destroy(cache->segments[j].core);
                }
                if (j < i) {
                    pthread_rwlock_destroy(&cache->segments[j].rwlock);
                }
            }
            return -1;
        }
    }
    
    if (pthread_mutex_init(&cache->stats_mutex, NULL) != 0) {
        for (int i = 0; i < CACHE_SEGMENTS; i++) {
            cache_core_destroy(cache->segments[i].core);
            pthread_rwlock_destroy(&cache->segments[i].rwlock);
        }
        return -1;
    }
    
    memset(&cache->stats, 0, sizeof(CacheStats));
    cache->cleanup_running = 0;
    
    return 0;
}

void cache_sync_destroy(Cache *cache) {
    if (!cache) return;
    
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        pthread_rwlock_wrlock(&cache->segments[i].rwlock);
        if (cache->segments[i].core) {
            cache_core_destroy(cache->segments[i].core);
            cache->segments[i].core = NULL;
        }
        pthread_rwlock_unlock(&cache->segments[i].rwlock);
        pthread_rwlock_destroy(&cache->segments[i].rwlock);
    }
    
    pthread_mutex_destroy(&cache->stats_mutex);
}

int cache_sync_read_lock(Cache *cache, int segment_idx) {
    if (!cache || segment_idx < 0 || segment_idx >= CACHE_SEGMENTS) {
        return -1;
    }
    return pthread_rwlock_rdlock(&cache->segments[segment_idx].rwlock);
}

int cache_sync_read_unlock(Cache *cache, int segment_idx) {
    if (!cache || segment_idx < 0 || segment_idx >= CACHE_SEGMENTS) {
        return -1;
    }
    return pthread_rwlock_unlock(&cache->segments[segment_idx].rwlock);
}

int cache_sync_write_lock(Cache *cache, int segment_idx) {
    if (!cache || segment_idx < 0 || segment_idx >= CACHE_SEGMENTS) {
        return -1;
    }
    return pthread_rwlock_wrlock(&cache->segments[segment_idx].rwlock);
}

int cache_sync_write_unlock(Cache *cache, int segment_idx) {
    if (!cache || segment_idx < 0 || segment_idx >= CACHE_SEGMENTS) {
        return -1;
    }
    return pthread_rwlock_unlock(&cache->segments[segment_idx].rwlock);
}

int cache_sync_write_lock_all(Cache *cache) {
    if (!cache) return -1;
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        if (pthread_rwlock_wrlock(&cache->segments[i].rwlock) != 0) {
            for (int j = 0; j < i; j++) {
                pthread_rwlock_unlock(&cache->segments[j].rwlock);
            }
            return -1;
        }
    }
    return 0;
}

int cache_sync_write_unlock_all(Cache *cache) {
    if (!cache) return -1;
    for (int i = 0; i < CACHE_SEGMENTS; i++) {
        pthread_rwlock_unlock(&cache->segments[i].rwlock);
    }
    return 0;
}

void cache_sync_stats_lock(Cache *cache) {
    if (cache) pthread_mutex_lock(&cache->stats_mutex);
}

void cache_sync_stats_unlock(Cache *cache) {
    if (cache) pthread_mutex_unlock(&cache->stats_mutex);
}

void cache_sync_increment_total(Cache *cache) {
    if (!cache) return;
    cache_sync_stats_lock(cache);
    cache->stats.total_queries++;
    cache_sync_stats_unlock(cache);
}

void cache_sync_increment_hits(Cache *cache) {
    if (!cache) return;
    cache_sync_stats_lock(cache);
    cache->stats.hits++;
    cache_sync_stats_unlock(cache);
}

void cache_sync_increment_misses(Cache *cache) {
    if (!cache) return;
    cache_sync_stats_lock(cache);
    cache->stats.misses++;
    cache_sync_stats_unlock(cache);
}

void cache_sync_increment_evictions(Cache *cache, size_t count) {
    if (!cache) return;
    cache_sync_stats_lock(cache);
    cache->stats.evictions += count;
    cache_sync_stats_unlock(cache);
}

void cache_sync_update_size(Cache *cache, ssize_t delta) {
    if (!cache) return;
    cache_sync_stats_lock(cache);
    if (delta >= 0) {
        cache->stats.current_size += (size_t)delta;
    } else {
        size_t abs_delta = (size_t)(-delta);
        if (cache->stats.current_size >= abs_delta) {
            cache->stats.current_size -= abs_delta;
        } else {
            cache->stats.current_size = 0;
        }
    }
    cache_sync_stats_unlock(cache);
}

void cache_sync_update_entry_count(Cache *cache, ssize_t delta) {
    if (!cache) return;
    cache_sync_stats_lock(cache);
    if (delta >= 0) {
        cache->stats.entry_count += (size_t)delta;
    } else {
        size_t abs_delta = (size_t)(-delta);
        if (cache->stats.entry_count >= abs_delta) {
            cache->stats.entry_count -= abs_delta;
        } else {
            cache->stats.entry_count = 0;
        }
    }
    cache_sync_stats_unlock(cache);
}
