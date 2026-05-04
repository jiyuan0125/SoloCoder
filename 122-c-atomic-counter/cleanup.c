#include "cleanup.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <unistd.h>

static void* cleanup_thread_func(void *arg) {
    cleanup_thread_t *cleanup = (cleanup_thread_t*)arg;
    rate_limiter_t *limiter = cleanup->limiter;
    
    while (cleanup->running) {
        sleep(CLEANUP_INTERVAL_SECONDS);
        
        if (!cleanup->running) break;
        
        time_t now = time(NULL);
        cleanup_expired_keys(limiter, now);
    }
    
    return NULL;
}

cleanup_thread_t* cleanup_start(rate_limiter_t *limiter) {
    if (!limiter) return NULL;
    
    cleanup_thread_t *cleanup = (cleanup_thread_t*)calloc(1, sizeof(cleanup_thread_t));
    if (!cleanup) return NULL;
    
    cleanup->limiter = limiter;
    cleanup->running = 1;
    
    int ret = pthread_create(&cleanup->thread, NULL, cleanup_thread_func, cleanup);
    if (ret != 0) {
        free(cleanup);
        return NULL;
    }
    
    return cleanup;
}

void cleanup_stop(cleanup_thread_t *cleanup) {
    if (!cleanup) return;
    
    cleanup->running = 0;
    pthread_join(cleanup->thread, NULL);
    free(cleanup);
}

typedef struct {
    char **keys;
    size_t count;
    size_t capacity;
} expired_keys_list_t;

static void add_expired_key(expired_keys_list_t *list, const char *key) {
    if (list->count >= list->capacity) {
        size_t new_cap = list->capacity == 0 ? 64 : list->capacity * 2;
        char **new_keys = (char**)realloc(list->keys, new_cap * sizeof(char*));
        if (!new_keys) return;
        list->keys = new_keys;
        list->capacity = new_cap;
    }
    
    list->keys[list->count] = strdup(key);
    if (list->keys[list->count]) {
        list->count++;
    }
}

static void free_expired_list(expired_keys_list_t *list) {
    for (size_t i = 0; i < list->count; i++) {
        free(list->keys[i]);
    }
    free(list->keys);
}

static void collect_expired_callback(key_counter_t *counter, void *user_data) {
    time_t *now_ptr = (time_t*)user_data;
    time_t now = *now_ptr;
    
    if (difftime(now, counter->last_activity) > EXPIRE_SECONDS) {
    }
}

void cleanup_expired_keys(rate_limiter_t *limiter, time_t now) {
    if (!limiter) return;
    
    expired_keys_list_t expired = {NULL, 0, 0};
    
    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        pthread_rwlock_rdlock(&limiter->hash_table.bucket_locks[i]);
        key_counter_t *curr = limiter->hash_table.buckets[i];
        while (curr) {
            if (difftime(now, curr->last_activity) > EXPIRE_SECONDS) {
                add_expired_key(&expired, curr->key);
            }
            curr = curr->next;
        }
        pthread_rwlock_unlock(&limiter->hash_table.bucket_locks[i]);
    }
    
    for (size_t i = 0; i < expired.count; i++) {
        counter_remove(limiter, expired.keys[i]);
    }
    
    free_expired_list(&expired);
}
