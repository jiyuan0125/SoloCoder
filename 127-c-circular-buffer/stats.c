#include "stats.h"
#include <stdlib.h>
#include <string.h>
#include <math.h>

struct sensor_buffer_internal {
    circular_buffer_t* buffer;
    pthread_rwlock_t buffer_lock;
    
    double prev_valid_value;
    bool has_prev_valid;
    
    stats_t cached_stats;
    bool stats_dirty;
    double sum_valid;
    size_t count_valid;
    
    pthread_rwlock_t stats_lock;
};

sensor_buffer_t* sensor_buffer_create(size_t capacity) {
    struct sensor_buffer_internal* sb = 
        (struct sensor_buffer_internal*)malloc(sizeof(struct sensor_buffer_internal));
    if (!sb) return NULL;
    
    sb->buffer = circular_buffer_create(capacity);
    if (!sb->buffer) {
        free(sb);
        return NULL;
    }
    
    if (pthread_rwlock_init(&sb->buffer_lock, NULL) != 0) {
        circular_buffer_destroy(sb->buffer);
        free(sb);
        return NULL;
    }
    
    if (pthread_rwlock_init(&sb->stats_lock, NULL) != 0) {
        pthread_rwlock_destroy(&sb->buffer_lock);
        circular_buffer_destroy(sb->buffer);
        free(sb);
        return NULL;
    }
    
    sb->prev_valid_value = 0.0;
    sb->has_prev_valid = false;
    sb->stats_dirty = true;
    sb->sum_valid = 0.0;
    sb->count_valid = 0;
    
    memset(&sb->cached_stats, 0, sizeof(stats_t));
    
    return (sensor_buffer_t*)sb;
}

void sensor_buffer_destroy(sensor_buffer_t* sb) {
    if (!sb) return;
    struct sensor_buffer_internal* internal = (struct sensor_buffer_internal*)sb;
    
    pthread_rwlock_destroy(&internal->stats_lock);
    pthread_rwlock_destroy(&internal->buffer_lock);
    circular_buffer_destroy(internal->buffer);
    free(internal);
}

static bool is_anomaly(double value, double prev_value, bool has_prev) {
    if (!has_prev) return false;
    return fabs(value - prev_value) > ANOMALY_THRESHOLD;
}

int sensor_buffer_write(sensor_buffer_t* sb, timestamp_t timestamp, double value) {
    if (!sb) return -1;
    struct sensor_buffer_internal* internal = (struct sensor_buffer_internal*)sb;
    
    if (pthread_rwlock_wrlock(&internal->buffer_lock) != 0) {
        return -1;
    }
    
    sensor_data_t data;
    data.timestamp = timestamp;
    data.value = value;
    data.is_anomaly = is_anomaly(value, internal->prev_valid_value, internal->has_prev_valid);
    
    if (circular_buffer_is_full(internal->buffer)) {
        sensor_data_t oldest;
        if (circular_buffer_read_recent(internal->buffer, &oldest, 1) == 1) {
            sensor_data_t all_data[BUFFER_SIZE];
            size_t count = circular_buffer_read_recent(internal->buffer, all_data, BUFFER_SIZE);
            if (count > 0 && !all_data[0].is_anomaly) {
                pthread_rwlock_wrlock(&internal->stats_lock);
                internal->sum_valid -= all_data[0].value;
                internal->count_valid--;
                internal->stats_dirty = true;
                pthread_rwlock_unlock(&internal->stats_lock);
            }
        }
    }
    
    if (!data.is_anomaly) {
        pthread_rwlock_wrlock(&internal->stats_lock);
        internal->sum_valid += value;
        internal->count_valid++;
        internal->stats_dirty = true;
        pthread_rwlock_unlock(&internal->stats_lock);
        
        internal->prev_valid_value = value;
        internal->has_prev_valid = true;
    }
    
    int result = circular_buffer_write(internal->buffer, &data);
    
    pthread_rwlock_unlock(&internal->buffer_lock);
    return result;
}

int sensor_buffer_read_latest(const sensor_buffer_t* sb, sensor_data_t* out) {
    if (!sb || !out) return -1;
    const struct sensor_buffer_internal* internal = (const struct sensor_buffer_internal*)sb;
    
    if (pthread_rwlock_rdlock(&((struct sensor_buffer_internal*)internal)->buffer_lock) != 0) {
        return -1;
    }
    
    int result = circular_buffer_read_latest(internal->buffer, out);
    
    pthread_rwlock_unlock(&((struct sensor_buffer_internal*)internal)->buffer_lock);
    return result;
}

size_t sensor_buffer_read_recent(const sensor_buffer_t* sb, sensor_data_t* out, size_t n) {
    if (!sb || !out) return 0;
    const struct sensor_buffer_internal* internal = (const struct sensor_buffer_internal*)sb;
    
    if (pthread_rwlock_rdlock(&((struct sensor_buffer_internal*)internal)->buffer_lock) != 0) {
        return 0;
    }
    
    size_t result = circular_buffer_read_recent(internal->buffer, out, n);
    
    pthread_rwlock_unlock(&((struct sensor_buffer_internal*)internal)->buffer_lock);
    return result;
}

size_t sensor_buffer_read_range(const sensor_buffer_t* sb, sensor_data_t* out, 
                                  timestamp_t start, timestamp_t end, size_t max_count) {
    if (!sb || !out) return 0;
    const struct sensor_buffer_internal* internal = (const struct sensor_buffer_internal*)sb;
    
    if (pthread_rwlock_rdlock(&((struct sensor_buffer_internal*)internal)->buffer_lock) != 0) {
        return 0;
    }
    
    size_t result = circular_buffer_read_range(internal->buffer, out, start, end, max_count);
    
    pthread_rwlock_unlock(&((struct sensor_buffer_internal*)internal)->buffer_lock);
    return result;
}

static void recalculate_stats(struct sensor_buffer_internal* internal) {
    if (!internal->stats_dirty) return;
    
    if (pthread_rwlock_rdlock(&internal->buffer_lock) != 0) {
        return;
    }
    
    size_t count = circular_buffer_size(internal->buffer);
    if (count == 0) {
        memset(&internal->cached_stats, 0, sizeof(stats_t));
        internal->stats_dirty = false;
        pthread_rwlock_unlock(&internal->buffer_lock);
        return;
    }
    
    sensor_data_t* all_data = (sensor_data_t*)malloc(count * sizeof(sensor_data_t));
    if (!all_data) {
        pthread_rwlock_unlock(&internal->buffer_lock);
        return;
    }
    
    circular_buffer_read_recent(internal->buffer, all_data, count);
    
    double current = all_data[count - 1].value;
    double max_val = -INFINITY;
    double min_val = INFINITY;
    
    for (size_t i = 0; i < count; i++) {
        if (!all_data[i].is_anomaly) {
            if (all_data[i].value > max_val) max_val = all_data[i].value;
            if (all_data[i].value < min_val) min_val = all_data[i].value;
        }
    }
    
    internal->cached_stats.current = current;
    internal->cached_stats.max = (max_val == -INFINITY) ? 0.0 : max_val;
    internal->cached_stats.min = (min_val == INFINITY) ? 0.0 : min_val;
    internal->cached_stats.avg = (internal->count_valid > 0) ? 
        (internal->sum_valid / internal->count_valid) : 0.0;
    internal->cached_stats.valid_count = internal->count_valid;
    
    internal->stats_dirty = false;
    
    free(all_data);
    pthread_rwlock_unlock(&internal->buffer_lock);
}

int sensor_buffer_get_stats(const sensor_buffer_t* sb, stats_t* out) {
    if (!sb || !out) return -1;
    struct sensor_buffer_internal* internal = (struct sensor_buffer_internal*)sb;
    
    if (pthread_rwlock_wrlock(&internal->stats_lock) != 0) {
        return -1;
    }
    
    if (internal->stats_dirty) {
        recalculate_stats(internal);
    }
    
    memcpy(out, &internal->cached_stats, sizeof(stats_t));
    
    pthread_rwlock_unlock(&internal->stats_lock);
    return 0;
}

size_t sensor_buffer_size(const sensor_buffer_t* sb) {
    if (!sb) return 0;
    const struct sensor_buffer_internal* internal = (const struct sensor_buffer_internal*)sb;
    
    if (pthread_rwlock_rdlock(&((struct sensor_buffer_internal*)internal)->buffer_lock) != 0) {
        return 0;
    }
    
    size_t result = circular_buffer_size(internal->buffer);
    
    pthread_rwlock_unlock(&((struct sensor_buffer_internal*)internal)->buffer_lock);
    return result;
}

size_t sensor_buffer_capacity(const sensor_buffer_t* sb) {
    if (!sb) return 0;
    const struct sensor_buffer_internal* internal = (const struct sensor_buffer_internal*)sb;
    return circular_buffer_capacity(internal->buffer);
}
