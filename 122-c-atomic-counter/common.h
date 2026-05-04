#ifndef RATE_LIMITER_COMMON_H
#define RATE_LIMITER_COMMON_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>
#include <time.h>

#define MAX_KEY_LEN          64
#define HASH_TABLE_SIZE      65536
#define SLIDING_WINDOW_BUCKETS  6
#define WINDOW_SECONDS       60
#define BUCKET_SECONDS       10
#define EXPIRE_SECONDS       300
#define CLEANUP_INTERVAL_SECONDS 60

typedef enum {
    RATE_LIMIT_FIXED_WINDOW,
    RATE_LIMIT_SLIDING_WINDOW
} rate_limit_mode_t;

typedef struct {
    uint64_t total_requests;
    uint64_t rejected_requests;
    uint32_t active_keys;
} global_stats_t;

typedef struct {
    uint64_t requests_last_minute;
    uint64_t rejected_last_minute;
    uint64_t total_allowed;
    uint64_t total_rejected;
    time_t last_activity;
} key_stats_t;

#define atomic_inc(ptr)      __sync_fetch_and_add((ptr), 1)
#define atomic_dec(ptr)      __sync_fetch_and_sub((ptr), 1)
#define atomic_add(ptr, val) __sync_fetch_and_add((ptr), (val))
#define atomic_sub(ptr, val) __sync_fetch_and_sub((ptr), (val))
#define atomic_xchg(ptr, val) __sync_lock_test_and_set((ptr), (val))
#define atomic_cmpxchg(ptr, old, new) __sync_val_compare_and_swap((ptr), (old), (new))

#endif
