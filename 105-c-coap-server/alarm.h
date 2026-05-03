#ifndef ALARM_H
#define ALARM_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>
#include "storage.h"

#define TEMP_ALARM_HIGH_THRESHOLD 40.0f
#define TEMP_ALARM_LOW_THRESHOLD 0.0f
#define ALARM_CONSECUTIVE_COUNT 3
#define MAX_ALARM_SENSORS 64
#define MAX_ALARM_HISTORY 100

typedef enum {
    ALARM_STATE_NORMAL = 0,
    ALARM_STATE_PENDING,
    ALARM_STATE_TRIGGERED
} alarm_state_t;

typedef enum {
    ALARM_TYPE_TEMP_HIGH = 0,
    ALARM_TYPE_TEMP_LOW
} alarm_type_t;

typedef struct {
    char sensor_id[MAX_SENSOR_ID_LEN];
    alarm_state_t state;
    alarm_type_t type;
    uint8_t consecutive_count;
    float last_abnormal_value;
    time_t first_abnormal_time;
    time_t triggered_time;
} alarm_sensor_tracker_t;

typedef struct {
    char sensor_id[MAX_SENSOR_ID_LEN];
    alarm_type_t type;
    float temperature;
    time_t start_time;
    time_t end_time;
    bool is_active;
} alarm_record_t;

typedef struct {
    alarm_sensor_tracker_t sensors[MAX_ALARM_SENSORS];
    uint8_t sensor_count;
    
    alarm_record_t history[MAX_ALARM_HISTORY];
    uint8_t history_count;
    uint8_t history_head;
} alarm_manager_t;

void alarm_manager_init(alarm_manager_t *mgr);
int alarm_check_reading(alarm_manager_t *mgr, const sensor_reading_t *reading);

alarm_record_t *alarm_get_active_alarm(alarm_manager_t *mgr, const char *sensor_id);
int alarm_get_all_active(alarm_manager_t *mgr, alarm_record_t *output, uint8_t *out_count);
int alarm_get_history_count(alarm_manager_t *mgr);
int alarm_get_history_record(alarm_manager_t *mgr, uint8_t index, alarm_record_t *record);

#endif
