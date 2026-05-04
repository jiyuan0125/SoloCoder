#ifndef PROGRESS_H
#define PROGRESS_H

#include <stdint.h>
#include <stddef.h>
#include "speed.h"

#define DEFAULT_BAR_WIDTH 40
#define MIN_TERMINAL_WIDTH 50

typedef struct {
    uint64_t total_bytes;
    uint64_t current_bytes;
    int bar_width;
    int is_complete;
    SpeedCalculator speed_calc;
    double start_time;
} ProgressState;

typedef struct {
    ProgressState *states;
    size_t count;
    size_t capacity;
    int terminal_supported;
    int terminal_width;
} ProgressManager;

void progress_state_init(ProgressState *state, uint64_t total_bytes, int bar_width);
void progress_state_destroy(ProgressState *state);
void progress_state_update(ProgressState *state, uint64_t bytes_added);
double progress_state_get_percent(const ProgressState *state);
double progress_state_get_remaining_seconds(const ProgressState *state);

void progress_manager_init(ProgressManager *manager, size_t initial_capacity);
void progress_manager_destroy(ProgressManager *manager);
size_t progress_manager_add(ProgressManager *manager, uint64_t total_bytes, int bar_width);
void progress_manager_update(ProgressManager *manager, size_t index, uint64_t bytes_added);
int progress_manager_all_complete(const ProgressManager *manager);
void progress_manager_detect_terminal(ProgressManager *manager);

#endif
