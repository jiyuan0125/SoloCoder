#include "progress.h"
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <unistd.h>
#include <stdio.h>
#include <time.h>
#include <sys/time.h>

static double get_current_time(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (double)tv.tv_sec + (double)tv.tv_usec / 1000000.0;
}

void progress_state_init(ProgressState *state, uint64_t total_bytes, int bar_width) {
    state->total_bytes = total_bytes;
    state->current_bytes = 0;
    state->bar_width = (bar_width > 0) ? bar_width : DEFAULT_BAR_WIDTH;
    state->is_complete = 0;
    state->start_time = get_current_time();
    speed_calculator_init(&state->speed_calc, 100);
}

void progress_state_destroy(ProgressState *state) {
    speed_calculator_destroy(&state->speed_calc);
}

void progress_state_update(ProgressState *state, uint64_t bytes_added) {
    if (state->is_complete) return;
    
    if (bytes_added > 0) {
        speed_calculator_add_sample(&state->speed_calc, bytes_added);
    }
    
    state->current_bytes += bytes_added;
    
    if (state->current_bytes >= state->total_bytes) {
        state->current_bytes = state->total_bytes;
        state->is_complete = 1;
    }
}

double progress_state_get_percent(const ProgressState *state) {
    if (state->total_bytes == 0) {
        return 0.0;
    }
    return (double)state->current_bytes / (double)state->total_bytes * 100.0;
}

double progress_state_get_remaining_seconds(const ProgressState *state) {
    if (state->is_complete) {
        return 0.0;
    }
    
    double speed = speed_calculator_get_average(&state->speed_calc);
    if (speed <= 0.0) {
        return -1.0;
    }
    
    uint64_t remaining_bytes = state->total_bytes - state->current_bytes;
    return (double)remaining_bytes / speed;
}

void progress_manager_init(ProgressManager *manager, size_t initial_capacity) {
    if (initial_capacity < 1) {
        initial_capacity = 4;
    }
    manager->states = (ProgressState *)malloc(initial_capacity * sizeof(ProgressState));
    manager->count = 0;
    manager->capacity = initial_capacity;
    manager->terminal_supported = 1;
    manager->terminal_width = 80;
    progress_manager_detect_terminal(manager);
}

void progress_manager_destroy(ProgressManager *manager) {
    for (size_t i = 0; i < manager->count; i++) {
        progress_state_destroy(&manager->states[i]);
    }
    free(manager->states);
    manager->states = NULL;
    manager->count = 0;
    manager->capacity = 0;
}

size_t progress_manager_add(ProgressManager *manager, uint64_t total_bytes, int bar_width) {
    if (manager->count >= manager->capacity) {
        size_t new_capacity = manager->capacity * 2;
        ProgressState *new_states = (ProgressState *)realloc(manager->states, 
                                                               new_capacity * sizeof(ProgressState));
        if (new_states == NULL) {
            return (size_t)-1;
        }
        manager->states = new_states;
        manager->capacity = new_capacity;
    }
    
    size_t index = manager->count;
    progress_state_init(&manager->states[index], total_bytes, bar_width);
    manager->count++;
    return index;
}

void progress_manager_update(ProgressManager *manager, size_t index, uint64_t bytes_added) {
    if (index >= manager->count) {
        return;
    }
    progress_state_update(&manager->states[index], bytes_added);
}

int progress_manager_all_complete(const ProgressManager *manager) {
    for (size_t i = 0; i < manager->count; i++) {
        if (!manager->states[i].is_complete) {
            return 0;
        }
    }
    return 1;
}

void progress_manager_detect_terminal(ProgressManager *manager) {
    manager->terminal_supported = isatty(STDOUT_FILENO);
    
    if (manager->terminal_supported) {
        struct winsize ws;
        if (ioctl(STDOUT_FILENO, TIOCGWINSZ, &ws) == 0) {
            manager->terminal_width = ws.ws_col;
        } else {
            manager->terminal_width = 80;
        }
    } else {
        manager->terminal_width = 80;
    }
}
