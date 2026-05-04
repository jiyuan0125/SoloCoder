#include "sliding_window.h"
#include <string.h>

bool rate_limit_check(rate_limiter_t *limiter, const char *key) {
    if (!limiter || !key) return false;
    
    key_counter_t *counter = counter_find(limiter, key);
    if (!counter) return true;
    
    time_t now = time(NULL);
    
    if (limiter->mode == RATE_LIMIT_FIXED_WINDOW) {
        time_t current_window_start = now - (now % WINDOW_SECONDS);
        if (current_window_start != counter->window.fixed.window_start) {
            return true;
        }
        return counter->window.fixed.requests < limiter->threshold;
    } else {
        time_t aligned = get_aligned_bucket_time(now);
        uint32_t total = 0;
        
        for (int i = 0; i < SLIDING_WINDOW_BUCKETS; i++) {
            if (counter->window.sliding.buckets[i].timestamp >= aligned - WINDOW_SECONDS + 1) {
                total += counter->window.sliding.buckets[i].requests;
            }
        }
        return total < limiter->threshold;
    }
}

bool fixed_window_check_and_inc(rate_limiter_t *limiter, key_counter_t *counter, time_t now) {
    bool allowed = false;
    
    pthread_spin_lock(&counter->spinlock);
    
    time_t current_window_start = now - (now % WINDOW_SECONDS);
    
    if (current_window_start != counter->window.fixed.window_start) {
        counter->window.fixed.window_start = current_window_start;
        counter->window.fixed.requests = 0;
    }
    
    if (counter->window.fixed.requests < limiter->threshold) {
        counter->window.fixed.requests++;
        counter->total_allowed++;
        counter->last_activity = now;
        allowed = true;
    } else {
        counter->total_rejected++;
    }
    
    pthread_spin_unlock(&counter->spinlock);
    
    if (allowed) {
        atomic_inc(&limiter->global_total_requests);
    } else {
        atomic_inc(&limiter->global_rejected_requests);
    }
    
    return allowed;
}

bool sliding_window_check_and_inc(rate_limiter_t *limiter, key_counter_t *counter, time_t now) {
    bool allowed = false;
    time_t aligned = get_aligned_bucket_time(now);
    uint32_t idx = get_bucket_index(aligned);
    
    pthread_spin_lock(&counter->spinlock);
    
    if (counter->window.sliding.buckets[idx].timestamp != aligned) {
        counter->window.sliding.buckets[idx].timestamp = aligned;
        counter->window.sliding.buckets[idx].requests = 0;
        counter->window.sliding.buckets[idx].rejected = 0;
    }
    
    uint32_t total = 0;
    for (int i = 0; i < SLIDING_WINDOW_BUCKETS; i++) {
        if (counter->window.sliding.buckets[i].timestamp >= aligned - WINDOW_SECONDS + 1) {
            total += counter->window.sliding.buckets[i].requests;
        }
    }
    
    if (total < limiter->threshold) {
        counter->window.sliding.buckets[idx].requests++;
        counter->total_allowed++;
        counter->last_activity = now;
        allowed = true;
    } else {
        counter->window.sliding.buckets[idx].rejected++;
        counter->total_rejected++;
    }
    
    pthread_spin_unlock(&counter->spinlock);
    
    if (allowed) {
        atomic_inc(&limiter->global_total_requests);
    } else {
        atomic_inc(&limiter->global_rejected_requests);
    }
    
    return allowed;
}

bool rate_limit_check_and_increment(rate_limiter_t *limiter, const char *key) {
    if (!limiter || !key) return false;
    
    key_counter_t *counter = counter_get_or_create(limiter, key);
    if (!counter) return false;
    
    time_t now = time(NULL);
    bool allowed;
    
    if (limiter->mode == RATE_LIMIT_FIXED_WINDOW) {
        allowed = fixed_window_check_and_inc(limiter, counter, now);
    } else {
        allowed = sliding_window_check_and_inc(limiter, counter, now);
    }
    
    return allowed;
}

void get_global_stats(rate_limiter_t *limiter, global_stats_t *stats) {
    if (!limiter || !stats) return;
    
    stats->total_requests = limiter->global_total_requests;
    stats->rejected_requests = limiter->global_rejected_requests;
    stats->active_keys = limiter->active_keys_count;
}

bool get_key_stats(rate_limiter_t *limiter, const char *key, key_stats_t *stats) {
    if (!limiter || !key || !stats) return false;
    
    key_counter_t *counter = counter_find(limiter, key);
    if (!counter) return false;
    
    time_t now = time(NULL);
    
    stats->total_allowed = counter->total_allowed;
    stats->total_rejected = counter->total_rejected;
    stats->last_activity = counter->last_activity;
    
    if (limiter->mode == RATE_LIMIT_FIXED_WINDOW) {
        time_t current_window_start = now - (now % WINDOW_SECONDS);
        if (current_window_start == counter->window.fixed.window_start) {
            stats->requests_last_minute = counter->window.fixed.requests;
            stats->rejected_last_minute = 0;
        } else {
            stats->requests_last_minute = 0;
            stats->rejected_last_minute = 0;
        }
    } else {
        time_t aligned = get_aligned_bucket_time(now);
        stats->requests_last_minute = 0;
        stats->rejected_last_minute = 0;
        
        for (int i = 0; i < SLIDING_WINDOW_BUCKETS; i++) {
            if (counter->window.sliding.buckets[i].timestamp >= aligned - WINDOW_SECONDS + 1) {
                stats->requests_last_minute += counter->window.sliding.buckets[i].requests;
                stats->rejected_last_minute += counter->window.sliding.buckets[i].rejected;
            }
        }
    }
    
    return true;
}
