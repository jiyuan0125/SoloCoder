#ifndef THREAD_POOL_H
#define THREAD_POOL_H

#include "common.h"
#include "task_queue.h"
#include "stats.h"
#include <pthread.h>
#include <signal.h>
#include <setjmp.h>

typedef struct tp_worker_info tp_worker_info_t;

typedef struct {
    tp_completion_t *buffer;
    size_t capacity;
    size_t head;
    size_t tail;
    size_t count;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    int closed;
} tp_completion_queue_t;

typedef struct {
    tp_task_queue_t task_queue;
    tp_completion_queue_t completion_queue;
    tp_stats_t stats;
    tp_worker_info_t *workers;
    size_t num_workers;
    tp_reject_policy_t reject_policy;
    tp_state_t state;
    pthread_mutex_t pool_mutex;
    pthread_cond_t shutdown_cond;
    size_t queue_capacity;
} tp_thread_pool_t;

typedef struct {
    size_t num_threads;
    size_t queue_capacity;
    tp_reject_policy_t reject_policy;
} tp_config_t;

#define TP_CONFIG_INIT() \
    { 4, 16, TP_REJECT_BLOCK }

tp_status_t tp_thread_pool_init(tp_thread_pool_t *pool, const tp_config_t *config);
void tp_thread_pool_destroy(tp_thread_pool_t *pool);
tp_status_t tp_thread_pool_submit(tp_thread_pool_t *pool, const tp_task_t *task);
tp_status_t tp_thread_pool_submit_simple(tp_thread_pool_t *pool, tp_task_func_t func, void *arg);
tp_status_t tp_thread_pool_shutdown(tp_thread_pool_t *pool, tp_shutdown_mode_t mode, int timeout_ms);
tp_status_t tp_thread_pool_get_stats(tp_thread_pool_t *pool, tp_stats_snapshot_t *snapshot);
tp_state_t tp_thread_pool_get_state(tp_thread_pool_t *pool);
size_t tp_thread_pool_get_queue_size(tp_thread_pool_t *pool);

tp_status_t tp_thread_pool_poll_completion(tp_thread_pool_t *pool, tp_completion_t *completion, int block);
size_t tp_thread_pool_completion_count(tp_thread_pool_t *pool);

#endif
