#include "counter.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

rate_limiter_t* rate_limiter_create(rate_limit_mode_t mode, uint32_t threshold) {
    rate_limiter_t *limiter = (rate_limiter_t*)calloc(1, sizeof(rate_limiter_t));
    if (!limiter) return NULL;
    
    limiter->mode = mode;
    limiter->threshold = threshold;
    
    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        pthread_rwlock_init(&limiter->hash_table.bucket_locks[i], NULL);
    }
    
    pthread_rwlock_init(&limiter->global_lock, NULL);
    
    return limiter;
}

void rate_limiter_destroy(rate_limiter_t *limiter) {
    if (!limiter) return;
    
    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        pthread_rwlock_wrlock(&limiter->hash_table.bucket_locks[i]);
        key_counter_t *curr = limiter->hash_table.buckets[i];
        while (curr) {
            key_counter_t *next = curr->next;
            pthread_spin_destroy(&curr->spinlock);
            free(curr);
            curr = next;
        }
        limiter->hash_table.buckets[i] = NULL;
        pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[i]);
        pthread_rwlock_destroy(&limiter->hash_table.bucket_locks[i]);
    }
    
    pthread_rwlock_destroy(&limiter->global_lock);
    free(limiter);
}

key_counter_t* counter_find(rate_limiter_t *limiter, const char *key) {
    if (!limiter || !key) return NULL;
    
    uint32_t hash = hash_string(key);
    key_counter_t *found = NULL;
    
    pthread_rwlock_rdlock(&limiter->hash_table.bucket_locks[hash]);
    key_counter_t *curr = limiter->hash_table.buckets[hash];
    while (curr) {
        if (strcmp(curr->key, key) == 0) {
            found = curr;
            break;
        }
        curr = curr->next;
    }
    pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[hash]);
    
    return found;
}

key_counter_t* counter_get_or_create(rate_limiter_t *limiter, const char *key) {
    if (!limiter || !key || strlen(key) >= MAX_KEY_LEN) return NULL;
    
    uint32_t hash = hash_string(key);
    
    pthread_rwlock_rdlock(&limiter->hash_table.bucket_locks[hash]);
    key_counter_t *curr = limiter->hash_table.buckets[hash];
    while (curr) {
        if (strcmp(curr->key, key) == 0) {
            pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[hash]);
            return curr;
        }
        curr = curr->next;
    }
    pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[hash]);
    
    key_counter_t *new_counter = (key_counter_t*)calloc(1, sizeof(key_counter_t));
    if (!new_counter) return NULL;
    
    strncpy(new_counter->key, key, MAX_KEY_LEN - 1);
    new_counter->key[MAX_KEY_LEN - 1] = '\0';
    
    pthread_spin_init(&new_counter->spinlock, PTHREAD_PROCESS_PRIVATE);
    
    time_t now = time(NULL);
    new_counter->last_activity = now;
    new_counter->next = NULL;
    
    if (limiter->mode == RATE_LIMIT_FIXED_WINDOW) {
        new_counter->window.fixed.requests = 0;
        new_counter->window.fixed.window_start = now - (now % WINDOW_SECONDS);
    } else {
        time_t aligned = now - (now % BUCKET_SECONDS);
        for (int i = 0; i < SLIDING_WINDOW_BUCKETS; i++) {
            new_counter->window.sliding.buckets[i].timestamp = aligned - (i * BUCKET_SECONDS);
            new_counter->window.sliding.buckets[i].requests = 0;
            new_counter->window.sliding.buckets[i].rejected = 0;
        }
    }
    
    pthread_rwlock_wrlock(&limiter->hash_table.bucket_locks[hash]);
    curr = limiter->hash_table.buckets[hash];
    int exists = 0;
    while (curr) {
        if (strcmp(curr->key, key) == 0) {
            exists = 1;
            break;
        }
        curr = curr->next;
    }
    
    if (exists) {
        free(new_counter);
        pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[hash]);
        return counter_find(limiter, key);
    }
    
    new_counter->next = limiter->hash_table.buckets[hash];
    limiter->hash_table.buckets[hash] = new_counter;
    pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[hash]);
    
    atomic_inc(&limiter->active_keys_count);
    
    return new_counter;
}

void counter_remove(rate_limiter_t *limiter, const char *key) {
    if (!limiter || !key) return;
    
    uint32_t hash = hash_string(key);
    
    pthread_rwlock_wrlock(&limiter->hash_table.bucket_locks[hash]);
    
    key_counter_t *prev = NULL;
    key_counter_t *curr = limiter->hash_table.buckets[hash];
    
    while (curr) {
        if (strcmp(curr->key, key) == 0) {
            if (prev) {
                prev->next = curr->next;
            } else {
                limiter->hash_table.buckets[hash] = curr->next;
            }
            pthread_spin_destroy(&curr->spinlock);
            free(curr);
            atomic_dec(&limiter->active_keys_count);
            break;
        }
        prev = curr;
        curr = curr->next;
    }
    
    pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[hash]);
}

void counter_traverse(rate_limiter_t *limiter, 
                      void (*callback)(key_counter_t*, void*), 
                      void *user_data) {
    if (!limiter || !callback) return;
    
    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        pthread_rwlock_rdlock(&limiter->hash_table.bucket_locks[i]);
        key_counter_t *curr = limiter->hash_table.buckets[i];
        while (curr) {
            callback(curr, user_data);
            curr = curr->next;
        }
        pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[i]);
    }
}
