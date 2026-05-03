#include <string.h>
#include <stdio.h>
#include "alarm.h"

void alarm_manager_init(alarm_manager_t *mgr) {
    memset(mgr, 0, sizeof(alarm_manager_t));
}

static bool is_temperature_abnormal(float temp, alarm_type_t *type) {
    if (temp > TEMP_ALARM_HIGH_THRESHOLD) {
        if (type) *type = ALARM_TYPE_TEMP_HIGH;
        return true;
    }
    if (temp < TEMP_ALARM_LOW_THRESHOLD) {
        if (type) *type = ALARM_TYPE_TEMP_LOW;
        return true;
    }
    return false;
}

static alarm_sensor_tracker_t *alarm_find_or_create_sensor(alarm_manager_t *mgr, const char *sensor_id) {
    for (uint8_t i = 0; i < mgr->sensor_count; i++) {
        if (strcmp(mgr->sensors[i].sensor_id, sensor_id) == 0) {
            return &mgr->sensors[i];
        }
    }
    
    if (mgr->sensor_count >= MAX_ALARM_SENSORS) {
        return NULL;
    }
    
    alarm_sensor_tracker_t *tracker = &mgr->sensors[mgr->sensor_count];
    memset(tracker, 0, sizeof(alarm_sensor_tracker_t));
    strncpy(tracker->sensor_id, sensor_id, MAX_SENSOR_ID_LEN - 1);
    tracker->state = ALARM_STATE_NORMAL;
    mgr->sensor_count++;
    
    return tracker;
}

static int alarm_record_create(alarm_manager_t *mgr, const char *sensor_id,
                                alarm_type_t type, float temp, time_t start_time) {
    for (uint8_t i = 0; i < mgr->history_count; i++) {
        uint8_t idx;
        if (mgr->history_count >= MAX_ALARM_HISTORY) {
            idx = (mgr->history_head + i) % MAX_ALARM_HISTORY;
        } else {
            idx = i;
        }
        
        if (mgr->history[idx].is_active && 
            strcmp(mgr->history[idx].sensor_id, sensor_id) == 0) {
            return 0;
        }
    }
    
    uint8_t idx;
    if (mgr->history_count < MAX_ALARM_HISTORY) {
        idx = mgr->history_count;
        mgr->history_count++;
    } else {
        idx = mgr->history_head;
        mgr->history_head = (mgr->history_head + 1) % MAX_ALARM_HISTORY;
    }
    
    alarm_record_t *record = &mgr->history[idx];
    memset(record, 0, sizeof(alarm_record_t));
    strncpy(record->sensor_id, sensor_id, MAX_SENSOR_ID_LEN - 1);
    record->type = type;
    record->temperature = temp;
    record->start_time = start_time;
    record->is_active = true;
    
    return 1;
}

static int alarm_record_resolve(alarm_manager_t *mgr, const char *sensor_id, time_t end_time) {
    int resolved = 0;
    
    for (uint8_t i = 0; i < mgr->history_count; i++) {
        uint8_t idx;
        if (mgr->history_count >= MAX_ALARM_HISTORY) {
            idx = (mgr->history_head + i) % MAX_ALARM_HISTORY;
        } else {
            idx = i;
        }
        
        if (mgr->history[idx].is_active && 
            strcmp(mgr->history[idx].sensor_id, sensor_id) == 0) {
            mgr->history[idx].is_active = false;
            mgr->history[idx].end_time = end_time;
            resolved++;
        }
    }
    
    return resolved;
}

int alarm_check_reading(alarm_manager_t *mgr, const sensor_reading_t *reading) {
    if (!mgr || !reading) {
        return -1;
    }
    
    if (reading->type != SENSOR_TYPE_TEMPERATURE) {
        return 0;
    }
    
    alarm_sensor_tracker_t *tracker = alarm_find_or_create_sensor(mgr, reading->sensor_id);
    if (!tracker) {
        return -1;
    }
    
    alarm_type_t current_type;
    bool is_abnormal = is_temperature_abnormal(reading->value.temperature, &current_type);
    
    int result = 0;
    
    if (is_abnormal) {
        if (tracker->state == ALARM_STATE_NORMAL) {
            tracker->state = ALARM_STATE_PENDING;
            tracker->type = current_type;
            tracker->consecutive_count = 1;
            tracker->last_abnormal_value = reading->value.temperature;
            tracker->first_abnormal_time = reading->timestamp;
        } else if (tracker->state == ALARM_STATE_PENDING) {
            if (tracker->type == current_type) {
                tracker->consecutive_count++;
                tracker->last_abnormal_value = reading->value.temperature;
                
                if (tracker->consecutive_count >= ALARM_CONSECUTIVE_COUNT) {
                    tracker->state = ALARM_STATE_TRIGGERED;
                    tracker->triggered_time = reading->timestamp;
                    
                    int created = alarm_record_create(mgr, reading->sensor_id, current_type,
                                                       reading->value.temperature,
                                                       tracker->first_abnormal_time);
                    result = (created > 0) ? 1 : 0;
                }
            } else {
                tracker->type = current_type;
                tracker->consecutive_count = 1;
                tracker->last_abnormal_value = reading->value.temperature;
                tracker->first_abnormal_time = reading->timestamp;
            }
        } else if (tracker->state == ALARM_STATE_TRIGGERED) {
            tracker->last_abnormal_value = reading->value.temperature;
        }
    } else {
        if (tracker->state == ALARM_STATE_TRIGGERED) {
            alarm_record_resolve(mgr, reading->sensor_id, reading->timestamp);
            result = 2;
        }
        
        tracker->state = ALARM_STATE_NORMAL;
        tracker->consecutive_count = 0;
        tracker->type = 0;
    }
    
    return result;
}

alarm_record_t *alarm_get_active_alarm(alarm_manager_t *mgr, const char *sensor_id) {
    if (!mgr || !sensor_id) {
        return NULL;
    }
    
    for (uint8_t i = 0; i < mgr->history_count; i++) {
        uint8_t idx;
        if (mgr->history_count >= MAX_ALARM_HISTORY) {
            idx = (mgr->history_head + i) % MAX_ALARM_HISTORY;
        } else {
            idx = i;
        }
        
        if (mgr->history[idx].is_active && 
            strcmp(mgr->history[idx].sensor_id, sensor_id) == 0) {
            return &mgr->history[idx];
        }
    }
    
    return NULL;
}

int alarm_get_all_active(alarm_manager_t *mgr, alarm_record_t *output, uint8_t *out_count) {
    if (!mgr || !output || !out_count) {
        return -1;
    }
    
    uint8_t count = 0;
    
    for (uint8_t i = 0; i < mgr->history_count; i++) {
        uint8_t idx;
        if (mgr->history_count >= MAX_ALARM_HISTORY) {
            idx = (mgr->history_head + i) % MAX_ALARM_HISTORY;
        } else {
            idx = i;
        }
        
        if (mgr->history[idx].is_active) {
            memcpy(&output[count], &mgr->history[idx], sizeof(alarm_record_t));
            count++;
        }
    }
    
    *out_count = count;
    return 0;
}

int alarm_get_history_count(alarm_manager_t *mgr) {
    if (!mgr) {
        return -1;
    }
    return mgr->history_count;
}

int alarm_get_history_record(alarm_manager_t *mgr, uint8_t index, alarm_record_t *record) {
    if (!mgr || !record || index >= mgr->history_count) {
        return -1;
    }
    
    uint8_t idx;
    if (mgr->history_count >= MAX_ALARM_HISTORY) {
        idx = (mgr->history_head + index) % MAX_ALARM_HISTORY;
    } else {
        idx = index;
    }
    
    memcpy(record, &mgr->history[idx], sizeof(alarm_record_t));
    return 0;
}
