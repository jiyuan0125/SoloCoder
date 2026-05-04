#include "circular_buffer.h"
#include <stdlib.h>
#include <string.h>

struct circular_buffer {
    sensor_data_t* buffer;
    size_t capacity;
    size_t head;
    size_t tail;
    size_t count;
};

circular_buffer_t* circular_buffer_create(size_t capacity) {
    if (capacity == 0) return NULL;
    
    circular_buffer_t* cb = (circular_buffer_t*)malloc(sizeof(circular_buffer_t));
    if (!cb) return NULL;
    
    cb->buffer = (sensor_data_t*)malloc(capacity * sizeof(sensor_data_t));
    if (!cb->buffer) {
        free(cb);
        return NULL;
    }
    
    cb->capacity = capacity;
    cb->head = 0;
    cb->tail = 0;
    cb->count = 0;
    
    return cb;
}

void circular_buffer_destroy(circular_buffer_t* cb) {
    if (!cb) return;
    free(cb->buffer);
    free(cb);
}

int circular_buffer_write(circular_buffer_t* cb, const sensor_data_t* data) {
    if (!cb || !data) return -1;
    
    memcpy(&cb->buffer[cb->tail], data, sizeof(sensor_data_t));
    
    if (cb->count == cb->capacity) {
        cb->head = (cb->head + 1) % cb->capacity;
    } else {
        cb->count++;
    }
    
    cb->tail = (cb->tail + 1) % cb->capacity;
    return 0;
}

int circular_buffer_read_latest(const circular_buffer_t* cb, sensor_data_t* out) {
    if (!cb || !out || cb->count == 0) return -1;
    
    size_t latest_idx = (cb->tail == 0) ? (cb->capacity - 1) : (cb->tail - 1);
    memcpy(out, &cb->buffer[latest_idx], sizeof(sensor_data_t));
    return 0;
}

size_t circular_buffer_read_recent(const circular_buffer_t* cb, sensor_data_t* out, size_t n) {
    if (!cb || !out || n == 0 || cb->count == 0) return 0;
    
    size_t read_count = (n < cb->count) ? n : cb->count;
    size_t start_idx = (cb->tail - read_count + cb->capacity) % cb->capacity;
    
    for (size_t i = 0; i < read_count; i++) {
        size_t idx = (start_idx + i) % cb->capacity;
        memcpy(&out[i], &cb->buffer[idx], sizeof(sensor_data_t));
    }
    
    return read_count;
}

size_t circular_buffer_read_range(const circular_buffer_t* cb, sensor_data_t* out, 
                                    timestamp_t start, timestamp_t end, size_t max_count) {
    if (!cb || !out || cb->count == 0) return 0;
    
    size_t count = 0;
    size_t start_idx = cb->head;
    
    for (size_t i = 0; i < cb->count && count < max_count; i++) {
        size_t idx = (start_idx + i) % cb->capacity;
        const sensor_data_t* data = &cb->buffer[idx];
        
        if (data->timestamp >= start && data->timestamp <= end) {
            memcpy(&out[count], data, sizeof(sensor_data_t));
            count++;
        }
    }
    
    return count;
}

size_t circular_buffer_size(const circular_buffer_t* cb) {
    return cb ? cb->count : 0;
}

size_t circular_buffer_capacity(const circular_buffer_t* cb) {
    return cb ? cb->capacity : 0;
}

bool circular_buffer_is_empty(const circular_buffer_t* cb) {
    return !cb || cb->count == 0;
}

bool circular_buffer_is_full(const circular_buffer_t* cb) {
    return cb && cb->count == cb->capacity;
}

static int compare_timestamp(const void* a, const void* b) {
    const sensor_data_t* da = (const sensor_data_t*)a;
    const sensor_data_t* db = (const sensor_data_t*)b;
    
    if (da->timestamp < db->timestamp) return -1;
    if (da->timestamp > db->timestamp) return 1;
    return 0;
}

void circular_buffer_sort_by_timestamp(circular_buffer_t* cb) {
    if (!cb || cb->count <= 1) return;
    
    sensor_data_t* temp = (sensor_data_t*)malloc(cb->count * sizeof(sensor_data_t));
    if (!temp) return;
    
    for (size_t i = 0; i < cb->count; i++) {
        size_t idx = (cb->head + i) % cb->capacity;
        memcpy(&temp[i], &cb->buffer[idx], sizeof(sensor_data_t));
    }
    
    qsort(temp, cb->count, sizeof(sensor_data_t), compare_timestamp);
    
    for (size_t i = 0; i < cb->count; i++) {
        size_t idx = (cb->head + i) % cb->capacity;
        memcpy(&cb->buffer[idx], &temp[i], sizeof(sensor_data_t));
    }
    
    free(temp);
}
