#ifndef STORAGE_H
#define STORAGE_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#define MAX_SENSOR_READINGS 1000
#define MAX_SENSOR_ID_LEN 32
#define MAX_SENSOR_TYPE_LEN 16

typedef enum {
    SENSOR_TYPE_TEMPERATURE = 0,
    SENSOR_TYPE_HUMIDITY,
    SENSOR_TYPE_DOOR_WINDOW
} sensor_type_t;

typedef struct {
    char sensor_id[MAX_SENSOR_ID_LEN];
    sensor_type_t type;
    union {
        float temperature;
        float humidity;
        bool is_open;
    } value;
    uint8_t battery_percent;
    time_t timestamp;
} sensor_reading_t;

typedef struct {
    sensor_reading_t buffer[MAX_SENSOR_READINGS];
    uint32_t head;
    uint32_t count;
    uint32_t total;
} storage_ring_buffer_t;

void storage_init(storage_ring_buffer_t *storage);
int storage_add_reading(storage_ring_buffer_t *storage, const sensor_reading_t *reading);

sensor_reading_t *storage_get_latest(storage_ring_buffer_t *storage, const char *sensor_id);
int storage_get_history(storage_ring_buffer_t *storage, const char *sensor_id,
                        uint32_t limit, sensor_reading_t *output, uint32_t *out_count);
int storage_get_top_temperature(storage_ring_buffer_t *storage,
                                uint32_t limit, sensor_reading_t *output, uint32_t *out_count);

int storage_count_sensor_readings(storage_ring_buffer_t *storage, const char *sensor_id);
int storage_get_all_count(storage_ring_buffer_t *storage);
int storage_get_reading_by_index(storage_ring_buffer_t *storage, uint32_t index,
                                  sensor_reading_t *reading);

#endif
