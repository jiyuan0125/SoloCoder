#include <string.h>
#include <stdlib.h>
#include "storage.h"

void storage_init(storage_ring_buffer_t *storage) {
    memset(storage, 0, sizeof(storage_ring_buffer_t));
}

int storage_add_reading(storage_ring_buffer_t *storage, const sensor_reading_t *reading) {
    if (!storage || !reading) {
        return -1;
    }
    
    uint32_t idx = storage->head;
    memcpy(&storage->buffer[idx], reading, sizeof(sensor_reading_t));
    
    storage->head = (storage->head + 1) % MAX_SENSOR_READINGS;
    if (storage->count < MAX_SENSOR_READINGS) {
        storage->count++;
    }
    storage->total++;
    
    return 0;
}

static uint32_t get_physical_index(storage_ring_buffer_t *storage, uint32_t logical_idx) {
    if (storage->count < MAX_SENSOR_READINGS) {
        return logical_idx % storage->count;
    } else {
        return (storage->head + logical_idx) % MAX_SENSOR_READINGS;
    }
}

sensor_reading_t *storage_get_latest(storage_ring_buffer_t *storage, const char *sensor_id) {
    if (!storage || !sensor_id || storage->count == 0) {
        return NULL;
    }
    
    for (uint32_t i = 0; i < storage->count; i++) {
        uint32_t logical_reverse = storage->count - 1 - i;
        uint32_t idx = get_physical_index(storage, logical_reverse);
        
        if (strcmp(storage->buffer[idx].sensor_id, sensor_id) == 0) {
            return &storage->buffer[idx];
        }
    }
    
    return NULL;
}

int storage_get_history(storage_ring_buffer_t *storage, const char *sensor_id,
                        uint32_t limit, sensor_reading_t *output, uint32_t *out_count) {
    if (!storage || !sensor_id || !output || !out_count) {
        return -1;
    }
    
    *out_count = 0;
    uint32_t found = 0;
    
    for (uint32_t i = 0; i < storage->count && found < limit; i++) {
        uint32_t logical_reverse = storage->count - 1 - i;
        uint32_t idx = get_physical_index(storage, logical_reverse);
        
        if (strcmp(storage->buffer[idx].sensor_id, sensor_id) == 0) {
            memcpy(&output[found], &storage->buffer[idx], sizeof(sensor_reading_t));
            found++;
        }
    }
    
    *out_count = found;
    return 0;
}

static int compare_temperature_desc(const void *a, const void *b) {
    const sensor_reading_t *ra = (const sensor_reading_t *)a;
    const sensor_reading_t *rb = (const sensor_reading_t *)b;
    
    if (ra->value.temperature < rb->value.temperature) return 1;
    if (ra->value.temperature > rb->value.temperature) return -1;
    return 0;
}

int storage_get_top_temperature(storage_ring_buffer_t *storage,
                                uint32_t limit, sensor_reading_t *output, uint32_t *out_count) {
    if (!storage || !output || !out_count) {
        return -1;
    }
    
    uint32_t temp_count = 0;
    sensor_reading_t *temp_readings = NULL;
    
    for (uint32_t i = 0; i < storage->count; i++) {
        uint32_t idx = get_physical_index(storage, i);
        if (storage->buffer[idx].type == SENSOR_TYPE_TEMPERATURE) {
            sensor_reading_t *new_arr = realloc(temp_readings, (temp_count + 1) * sizeof(sensor_reading_t));
            if (!new_arr) {
                free(temp_readings);
                return -1;
            }
            temp_readings = new_arr;
            memcpy(&temp_readings[temp_count], &storage->buffer[idx], sizeof(sensor_reading_t));
            temp_count++;
        }
    }
    
    if (temp_count == 0) {
        free(temp_readings);
        *out_count = 0;
        return 0;
    }
    
    qsort(temp_readings, temp_count, sizeof(sensor_reading_t), compare_temperature_desc);
    
    uint32_t copy_count = (limit < temp_count) ? limit : temp_count;
    memcpy(output, temp_readings, copy_count * sizeof(sensor_reading_t));
    *out_count = copy_count;
    
    free(temp_readings);
    return 0;
}

int storage_count_sensor_readings(storage_ring_buffer_t *storage, const char *sensor_id) {
    if (!storage || !sensor_id) {
        return -1;
    }
    
    int count = 0;
    for (uint32_t i = 0; i < storage->count; i++) {
        uint32_t idx = get_physical_index(storage, i);
        if (strcmp(storage->buffer[idx].sensor_id, sensor_id) == 0) {
            count++;
        }
    }
    return count;
}

int storage_get_all_count(storage_ring_buffer_t *storage) {
    if (!storage) {
        return -1;
    }
    return storage->count;
}

int storage_get_reading_by_index(storage_ring_buffer_t *storage, uint32_t index,
                                  sensor_reading_t *reading) {
    if (!storage || !reading || index >= storage->count) {
        return -1;
    }
    
    uint32_t idx = get_physical_index(storage, index);
    memcpy(reading, &storage->buffer[idx], sizeof(sensor_reading_t));
    return 0;
}
