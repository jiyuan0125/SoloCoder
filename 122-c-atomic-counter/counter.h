#ifndef RATE_LIMITER_COUNTER_H
#define RATE_LIMITER_COUNTER_H

#include "common.h"
#include <pthread.h>

typedef struct bucket_counter {
    volatile uint32_t requests;
    volatile uint32_t rejected;
    time_t timestamp;
} bucket_counter_t;

typedef struct key_counter {
    char key[MAX_KEY_LEN];
    struct key_counter *next;
    
    volatile uint64_t total_allowed;
    volatile uint64_t total_rejected;
    volatile time_t last_activity;
    
    pthread_spinlock_t spinlock;
    
    union {
        struct {
            volatile uint32_t requests;
            time_t window_start;
        } fixed;
        struct {
            bucket_counter_t buckets[SLIDING_WINDOW_BUCKETS];
        } sliding;
    } window;
} key_counter_t;

typedef struct hash_table {
    key_counter_t *buckets[HASH_TABLE_SIZE];
    pthread_rwlock_t bucket_locks[HASH_TABLE_SIZE];
} hash_table_t;

typedef struct rate_limiter {
    hash_table_t hash_table;
    rate_limit_mode_t mode;
    uint32_t threshold;
    
    volatile uint64_t global_total_requests;
    volatile uint64_t global_rejected_requests;
    volatile uint32_t active_keys_count;
    
    pthread_rwlock_t global_lock;
} rate_limiter_t;

rate_limiter_t* rate_limiter_create(rate_limit_mode_t mode, uint32_t threshold);
void rate_limiter_destroy(rate_limiter_t *limiter);

key_counter_t* counter_get_or_create(rate_limiter_t *limiter, const char *key);
key_counter_t* counter_find(rate_limiter_t *limiter, const char *key);
void counter_remove(rate_limiter_t *limiter, const char *key);
void counter_traverse(rate_limiter_t *limiter, 
                      void (*callback)(key_counter_t*, void*), 
                      void *user_data);

static inline uint32_t hash_string(const char *str) {
    uint32_t hash = 5381;
    int c;
    while ((c = (unsigned char)*str++))
        hash = ((hash << 5) + hash) + c;
    return hash % HASH_TABLE_SIZE;
}

#endif
