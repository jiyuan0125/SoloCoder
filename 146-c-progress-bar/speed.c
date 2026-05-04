#include "speed.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>

void speed_calculator_init(SpeedCalculator *calc, size_t max_samples) {
    calc->samples = (SpeedSample *)malloc(max_samples * sizeof(SpeedSample));
    calc->capacity = max_samples;
    calc->count = 0;
    calc->head = 0;
}

void speed_calculator_destroy(SpeedCalculator *calc) {
    if (calc->samples) {
        free(calc->samples);
        calc->samples = NULL;
    }
    calc->capacity = 0;
    calc->count = 0;
    calc->head = 0;
}

void speed_calculator_add_sample(SpeedCalculator *calc, uint64_t bytes) {
    if (calc->capacity == 0) return;
    
    time_t now = time(NULL);
    size_t index = calc->head;
    
    calc->samples[index].bytes = bytes;
    calc->samples[index].timestamp = now;
    
    calc->head = (calc->head + 1) % calc->capacity;
    if (calc->count < calc->capacity) {
        calc->count++;
    }
}

double speed_calculator_get_average(const SpeedCalculator *calc) {
    if (calc->count < 2) {
        return 0.0;
    }
    
    time_t now = time(NULL);
    time_t window_start = now - SPEED_WINDOW_SECONDS;
    
    uint64_t total_bytes = 0;
    time_t earliest_time = now;
    time_t latest_time = 0;
    int valid_samples = 0;
    
    for (size_t i = 0; i < calc->count; i++) {
        size_t idx;
        if (calc->count == calc->capacity) {
            idx = (calc->head + i) % calc->capacity;
        } else {
            idx = i;
        }
        
        if (calc->samples[idx].timestamp >= window_start) {
            total_bytes += calc->samples[idx].bytes;
            if (calc->samples[idx].timestamp < earliest_time) {
                earliest_time = calc->samples[idx].timestamp;
            }
            if (calc->samples[idx].timestamp > latest_time) {
                latest_time = calc->samples[idx].timestamp;
            }
            valid_samples++;
        }
    }
    
    if (valid_samples < 2 || latest_time <= earliest_time) {
        return 0.0;
    }
    
    double duration = difftime(latest_time, earliest_time);
    if (duration <= 0.0) {
        return 0.0;
    }
    
    return (double)total_bytes / duration;
}

void speed_format_unit(double bytes_per_sec, char *buf, size_t buf_size) {
    const char *units[] = {"B/s", "KB/s", "MB/s", "GB/s"};
    int unit_index = 0;
    double speed = bytes_per_sec;
    
    while (speed >= 1024.0 && unit_index < 3) {
        speed /= 1024.0;
        unit_index++;
    }
    
    if (speed < 10.0) {
        snprintf(buf, buf_size, "%.2f %s", speed, units[unit_index]);
    } else if (speed < 100.0) {
        snprintf(buf, buf_size, "%.1f %s", speed, units[unit_index]);
    } else {
        snprintf(buf, buf_size, "%.0f %s", speed, units[unit_index]);
    }
}

void time_format_remaining(double seconds, char *buf, size_t buf_size) {
    if (seconds < 0.0 || isnan(seconds) || isinf(seconds)) {
        snprintf(buf, buf_size, "--:--");
        return;
    }
    
    int total_seconds = (int)round(seconds);
    
    if (total_seconds < 0) {
        snprintf(buf, buf_size, "--:--");
        return;
    }
    
    int hours = total_seconds / 3600;
    int minutes = (total_seconds % 3600) / 60;
    int secs = total_seconds % 60;
    
    if (hours > 0) {
        snprintf(buf, buf_size, "%02d:%02d:%02d", hours, minutes, secs);
    } else {
        snprintf(buf, buf_size, "%02d:%02d", minutes, secs);
    }
}
