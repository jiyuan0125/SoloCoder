#include "concurrent_buffer.h"
#include <stdlib.h>

concurrent_buffer_t* concurrent_buffer_create(size_t capacity) {
    concurrent_buffer_t* cb = (concurrent_buffer_t*)malloc(sizeof(concurrent_buffer_t));
    if (!cb) return NULL;
    
    cb->buffer = circular_buffer_create(capacity);
    if (!cb->buffer) {
        free(cb);
        return NULL;
    }
    
    if (pthread_rwlock_init(&cb->rwlock, NULL) != 0) {
        circular_buffer_destroy(cb->buffer);
        free(cb);
        return NULL;
    }
    
    return cb;
}

void concurrent_buffer_destroy(concurrent_buffer_t* cb) {
    if (!cb) return;
    
    pthread_rwlock_destroy(&cb->rwlock);
    circular_buffer_destroy(cb->buffer);
    free(cb);
}

int concurrent_buffer_write(concurrent_buffer_t* cb, const sensor_data_t* data) {
    if (!cb) return -1;
    
    if (pthread_rwlock_wrlock(&cb->rwlock) != 0) {
        return -1;
    }
    
    int result = circular_buffer_write(cb->buffer, data);
    
    pthread_rwlock_unlock(&cb->rwlock);
    return result;
}

int concurrent_buffer_read_latest(const concurrent_buffer_t* cb, sensor_data_t* out) {
    if (!cb) return -1;
    
    if (pthread_rwlock_rdlock(&((concurrent_buffer_t*)cb)->rwlock) != 0) {
        return -1;
    }
    
    int result = circular_buffer_read_latest(cb->buffer, out);
    
    pthread_rwlock_unlock(&((concurrent_buffer_t*)cb)->rwlock);
    return result;
}

size_t concurrent_buffer_read_recent(const concurrent_buffer_t* cb, sensor_data_t* out, size_t n) {
    if (!cb) return 0;
    
    if (pthread_rwlock_rdlock(&((concurrent_buffer_t*)cb)->rwlock) != 0) {
        return 0;
    }
    
    size_t result = circular_buffer_read_recent(cb->buffer, out, n);
    
    pthread_rwlock_unlock(&((concurrent_buffer_t*)cb)->rwlock);
    return result;
}

size_t concurrent_buffer_read_range(const concurrent_buffer_t* cb, sensor_data_t* out, 
                                      timestamp_t start, timestamp_t end, size_t max_count) {
    if (!cb) return 0;
    
    if (pthread_rwlock_rdlock(&((concurrent_buffer_t*)cb)->rwlock) != 0) {
        return 0;
    }
    
    size_t result = circular_buffer_read_range(cb->buffer, out, start, end, max_count);
    
    pthread_rwlock_unlock(&((concurrent_buffer_t*)cb)->rwlock);
    return result;
}

size_t concurrent_buffer_size(const concurrent_buffer_t* cb) {
    if (!cb) return 0;
    
    if (pthread_rwlock_rdlock(&((concurrent_buffer_t*)cb)->rwlock) != 0) {
        return 0;
    }
    
    size_t result = circular_buffer_size(cb->buffer);
    
    pthread_rwlock_unlock(&((concurrent_buffer_t*)cb)->rwlock);
    return result;
}

size_t concurrent_buffer_capacity(const concurrent_buffer_t* cb) {
    if (!cb) return 0;
    return circular_buffer_capacity(cb->buffer);
}
