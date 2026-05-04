#ifndef RATE_LIMITER_SLIDING_WINDOW_H
#define RATE_LIMITER_SLIDING_WINDOW_H

#include "counter.h"

bool rate_limit_check(rate_limiter_t *limiter, const char *key);
bool rate_limit_check_and_increment(rate_limiter_t *limiter, const char *key);

static inline time_t get_aligned_bucket_time(time_t now) {
    return now - (now % BUCKET_SECONDS);
}

static inline uint32_t get_bucket_index(time_t aligned_time) {
    return (aligned_time / BUCKET_SECONDS) % SLIDING_WINDOW_BUCKETS;
}

bool fixed_window_check_and_inc(rate_limiter_t *limiter, key_counter_t *counter, time_t now);
bool sliding_window_check_and_inc(rate_limiter_t *limiter, key_counter_t *counter, time_t now);

void get_global_stats(rate_limiter_t *limiter, global_stats_t *stats);
bool get_key_stats(rate_limiter_t *limiter, const char *key, key_stats_t *stats);

#endif
