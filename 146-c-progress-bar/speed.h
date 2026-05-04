#ifndef SPEED_H
#define SPEED_H

#include <stddef.h>
#include <stdint.h>
#include <time.h>

#define SPEED_WINDOW_SECONDS 5

typedef struct {
    uint64_t bytes;
    time_t timestamp;
} SpeedSample;

typedef struct {
    SpeedSample *samples;
    size_t capacity;
    size_t count;
    size_t head;
} SpeedCalculator;

void speed_calculator_init(SpeedCalculator *calc, size_t max_samples);
void speed_calculator_destroy(SpeedCalculator *calc);
void speed_calculator_add_sample(SpeedCalculator *calc, uint64_t bytes);
double speed_calculator_get_average(const SpeedCalculator *calc);
void speed_format_unit(double bytes_per_sec, char *buf, size_t buf_size);
void time_format_remaining(double seconds, char *buf, size_t buf_size);

#endif
