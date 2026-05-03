#ifndef STATS_H
#define STATS_H

#include "common.h"
#include <pthread.h>
#include <stdint.h>

typedef struct {
    uint64_t total_submitted;
    uint64_t total_completed;
    uint64_t total_rejected;
    uint64_t total_failures;
    size_t active_threads;
    size_t queued_tasks;
    pthread_mutex_t mutex;
} tp_stats_t;

typedef struct {
    uint64_t total_submitted;
    uint64_t total_completed;
    uint64_t total_rejected;
    uint64_t total_failures;
    size_t active_threads;
    size_t queued_tasks;
} tp_stats_snapshot_t;

tp_status_t tp_stats_init(tp_stats_t *stats);
void tp_stats_destroy(tp_stats_t *stats);
void tp_stats_inc_submitted(tp_stats_t *stats);
void tp_stats_inc_completed(tp_stats_t *stats);
void tp_stats_inc_rejected(tp_stats_t *stats);
void tp_stats_inc_failures(tp_stats_t *stats);
void tp_stats_set_active_threads(tp_stats_t *stats, size_t count);
void tp_stats_inc_active_threads(tp_stats_t *stats);
void tp_stats_dec_active_threads(tp_stats_t *stats);
void tp_stats_set_queued(tp_stats_t *stats, size_t count);
void tp_stats_inc_queued(tp_stats_t *stats);
void tp_stats_dec_queued(tp_stats_t *stats);
void tp_stats_snapshot(tp_stats_t *stats, tp_stats_snapshot_t *snapshot);
void tp_stats_reset(tp_stats_t *stats);

#endif
