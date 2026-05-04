#ifndef CONCURRENT_BUFFER_H
#define CONCURRENT_BUFFER_H

#include "circular_buffer.h"
#include <pthread.h>

typedef struct {
    circular_buffer_t* buffer;
    pthread_rwlock_t rwlock;
} concurrent_buffer_t;

concurrent_buffer_t* concurrent_buffer_create(size_t capacity);
void concurrent_buffer_destroy(concurrent_buffer_t* cb);

int concurrent_buffer_write(concurrent_buffer_t* cb, const sensor_data_t* data);
int concurrent_buffer_read_latest(const concurrent_buffer_t* cb, sensor_data_t* out);
size_t concurrent_buffer_read_recent(const concurrent_buffer_t* cb, sensor_data_t* out, size_t n);
size_t concurrent_buffer_read_range(const concurrent_buffer_t* cb, sensor_data_t* out, 
                                      timestamp_t start, timestamp_t end, size_t max_count);

size_t concurrent_buffer_size(const concurrent_buffer_t* cb);
size_t concurrent_buffer_capacity(const concurrent_buffer_t* cb);

#endif
