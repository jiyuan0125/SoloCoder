#ifndef CIRCULAR_BUFFER_H
#define CIRCULAR_BUFFER_H

#include "sensor_data.h"
#include <stddef.h>

typedef struct circular_buffer circular_buffer_t;

circular_buffer_t* circular_buffer_create(size_t capacity);
void circular_buffer_destroy(circular_buffer_t* cb);

int circular_buffer_write(circular_buffer_t* cb, const sensor_data_t* data);
int circular_buffer_read_latest(const circular_buffer_t* cb, sensor_data_t* out);
size_t circular_buffer_read_recent(const circular_buffer_t* cb, sensor_data_t* out, size_t n);
size_t circular_buffer_read_range(const circular_buffer_t* cb, sensor_data_t* out, 
                                    timestamp_t start, timestamp_t end, size_t max_count);

size_t circular_buffer_size(const circular_buffer_t* cb);
size_t circular_buffer_capacity(const circular_buffer_t* cb);
bool circular_buffer_is_empty(const circular_buffer_t* cb);
bool circular_buffer_is_full(const circular_buffer_t* cb);

void circular_buffer_sort_by_timestamp(circular_buffer_t* cb);

#endif
