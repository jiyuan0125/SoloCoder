#ifndef STATS_H
#define STATS_H

#include "concurrent_buffer.h"
#include <pthread.h>

typedef struct {
    concurrent_buffer_t* buffer;
    double prev_valid_value;
    bool has_prev_valid;
    
    stats_t cached_stats;
    bool stats_dirty;
    double sum_valid;
    size_t count_valid;
    
    pthread_rwlock_t stats_lock;
} sensor_buffer_t;

sensor_buffer_t* sensor_buffer_create(size_t capacity);
void sensor_buffer_destroy(sensor_buffer_t* sb);

int sensor_buffer_write(sensor_buffer_t* sb, timestamp_t timestamp, double value);

int sensor_buffer_read_latest(const sensor_buffer_t* sb, sensor_data_t* out);
size_t sensor_buffer_read_recent(const sensor_buffer_t* sb, sensor_data_t* out, size_t n);
size_t sensor_buffer_read_range(const sensor_buffer_t* sb, sensor_data_t* out, 
                                  timestamp_t start, timestamp_t end, size_t max_count);

int sensor_buffer_get_stats(const sensor_buffer_t* sb, stats_t* out);

size_t sensor_buffer_size(const sensor_buffer_t* sb);
size_t sensor_buffer_capacity(const sensor_buffer_t* sb);

#endif
