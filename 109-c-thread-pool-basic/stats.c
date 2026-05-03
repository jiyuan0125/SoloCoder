#include "stats.h"
#include <string.h>

tp_status_t tp_stats_init(tp_stats_t *stats)
{
    if (!stats) {
        return TP_INVALID_ARG;
    }

    memset(stats, 0, sizeof(tp_stats_t));

    if (pthread_mutex_init(&stats->mutex, NULL) != 0) {
        return TP_ERROR;
    }

    return TP_OK;
}

void tp_stats_destroy(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_destroy(&stats->mutex);
}

void tp_stats_inc_submitted(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->total_submitted++;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_inc_completed(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->total_completed++;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_inc_rejected(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->total_rejected++;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_inc_failures(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->total_failures++;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_set_active_threads(tp_stats_t *stats, size_t count)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->active_threads = count;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_inc_active_threads(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->active_threads++;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_dec_active_threads(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    if (stats->active_threads > 0) {
        stats->active_threads--;
    }
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_set_queued(tp_stats_t *stats, size_t count)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->queued_tasks = count;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_inc_queued(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->queued_tasks++;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_dec_queued(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    if (stats->queued_tasks > 0) {
        stats->queued_tasks--;
    }
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_snapshot(tp_stats_t *stats, tp_stats_snapshot_t *snapshot)
{
    if (!stats || !snapshot) return;

    pthread_mutex_lock(&stats->mutex);
    snapshot->total_submitted = stats->total_submitted;
    snapshot->total_completed = stats->total_completed;
    snapshot->total_rejected = stats->total_rejected;
    snapshot->total_failures = stats->total_failures;
    snapshot->active_threads = stats->active_threads;
    snapshot->queued_tasks = stats->queued_tasks;
    pthread_mutex_unlock(&stats->mutex);
}

void tp_stats_reset(tp_stats_t *stats)
{
    if (!stats) return;
    pthread_mutex_lock(&stats->mutex);
    stats->total_submitted = 0;
    stats->total_completed = 0;
    stats->total_rejected = 0;
    stats->total_failures = 0;
    pthread_mutex_unlock(&stats->mutex);
}
