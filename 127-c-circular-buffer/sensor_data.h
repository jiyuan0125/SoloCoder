#ifndef SENSOR_DATA_H
#define SENSOR_DATA_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#define BUFFER_SIZE 3600
#define ANOMALY_THRESHOLD 10.0
#define MAX_SENSORS 100

typedef uint64_t timestamp_t;

typedef struct {
    timestamp_t timestamp;
    double value;
    bool is_anomaly;
} sensor_data_t;

typedef struct {
    double current;
    double max;
    double min;
    double avg;
    size_t valid_count;
} stats_t;

static inline timestamp_t get_timestamp_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (timestamp_t)ts.tv_sec * 1000 + (timestamp_t)ts.tv_nsec / 1000000;
}

#endif
