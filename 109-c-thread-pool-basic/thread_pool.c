#include "thread_pool.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <sys/time.h>

struct tp_worker_info {
    pthread_t thread_id;
    size_t index;
    tp_thread_pool_t *pool;
    sigjmp_buf recovery_env;
    volatile sig_atomic_t running;
    tp_task_t current_task;
    volatile sig_atomic_t has_current_task;
};

static __thread tp_worker_info_t *current_worker = NULL;

static void crash_signal_handler(int sig)
{
    (void)sig;
    if (current_worker) {
        siglongjmp(current_worker->recovery_env, 1);
    }
}

static void setup_crash_handlers(void)
{
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = crash_signal_handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_NODEFER | SA_ONSTACK;

    sigaction(SIGSEGV, &sa, NULL);
    sigaction(SIGBUS, &sa, NULL);
    sigaction(SIGFPE, &sa, NULL);
    sigaction(SIGILL, &sa, NULL);
    sigaction(SIGABRT, &sa, NULL);
}

static tp_status_t completion_queue_init(tp_completion_queue_t *queue, size_t capacity)
{
    if (!queue || capacity == 0) {
        return TP_INVALID_ARG;
    }

    queue->buffer = (tp_completion_t *)calloc(capacity, sizeof(tp_completion_t));
    if (!queue->buffer) {
        return TP_ERROR;
    }

    queue->capacity = capacity;
    queue->head = 0;
    queue->tail = 0;
    queue->count = 0;
    queue->closed = 0;

    if (pthread_mutex_init(&queue->mutex, NULL) != 0) {
        free(queue->buffer);
        return TP_ERROR;
    }

    if (pthread_cond_init(&queue->not_empty, NULL) != 0) {
        pthread_mutex_destroy(&queue->mutex);
        free(queue->buffer);
        return TP_ERROR;
    }

    return TP_OK;
}

static void completion_queue_destroy(tp_completion_queue_t *queue)
{
    if (!queue) return;

    pthread_mutex_lock(&queue->mutex);
    if (queue->buffer) {
        free(queue->buffer);
        queue->buffer = NULL;
    }
    pthread_mutex_unlock(&queue->mutex);

    pthread_cond_destroy(&queue->not_empty);
    pthread_mutex_destroy(&queue->mutex);
}

static void completion_queue_close(tp_completion_queue_t *queue)
{
    if (!queue) return;

    pthread_mutex_lock(&queue->mutex);
    queue->closed = 1;
    pthread_cond_broadcast(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);
}

static tp_status_t completion_queue_push(tp_completion_queue_t *queue,
                                          void *arg, tp_completion_func_t comp, int success)
{
    if (!queue || !comp) {
        return TP_INVALID_ARG;
    }

    pthread_mutex_lock(&queue->mutex);

    if (queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_CLOSED;
    }

    if (queue->count >= queue->capacity) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_FULL;
    }

    queue->buffer[queue->tail].arg = arg;
    queue->buffer[queue->tail].completion = comp;
    queue->buffer[queue->tail].success = success;
    queue->tail = (queue->tail + 1) % queue->capacity;
    queue->count++;

    pthread_cond_signal(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);

    return TP_OK;
}

tp_status_t tp_thread_pool_poll_completion(tp_thread_pool_t *pool,
                                             tp_completion_t *completion, int block)
{
    if (!pool || !completion) {
        return TP_INVALID_ARG;
    }

    tp_completion_queue_t *queue = &pool->completion_queue;

    pthread_mutex_lock(&queue->mutex);

    while (queue->count == 0 && !queue->closed) {
        if (!block) {
            pthread_mutex_unlock(&queue->mutex);
            return TP_FULL;
        }
        pthread_cond_wait(&queue->not_empty, &queue->mutex);
    }

    if (queue->count == 0 && queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_CLOSED;
    }

    memcpy(completion, &queue->buffer[queue->head], sizeof(tp_completion_t));
    queue->head = (queue->head + 1) % queue->capacity;
    queue->count--;

    pthread_mutex_unlock(&queue->mutex);

    return TP_OK;
}

size_t tp_thread_pool_completion_count(tp_thread_pool_t *pool)
{
    if (!pool) return 0;

    tp_completion_queue_t *queue = &pool->completion_queue;

    pthread_mutex_lock(&queue->mutex);
    size_t count = queue->count;
    pthread_mutex_unlock(&queue->mutex);

    return count;
}

static void handle_completion(tp_thread_pool_t *pool, tp_task_t *task, int success)
{
    if (!task->completion) {
        return;
    }

    if (task->callback_mode == TP_CALLBACK_IN_WORKER) {
        task->completion(task->arg, success);
    } else {
        completion_queue_push(&pool->completion_queue, task->arg, task->completion, success);
    }
}

static void* worker_thread(void *arg)
{
    tp_worker_info_t *info = (tp_worker_info_t *)arg;
    tp_thread_pool_t *pool = info->pool;
    tp_task_t task;

    current_worker = info;
    info->has_current_task = 0;

    setup_crash_handlers();

    while (1) {
        pthread_mutex_lock(&pool->pool_mutex);
        if (pool->state == TP_STOPPED) {
            pthread_mutex_unlock(&pool->pool_mutex);
            break;
        }
        pthread_mutex_unlock(&pool->pool_mutex);

        info->has_current_task = 0;

        if (sigsetjmp(info->recovery_env, 1) != 0) {
            tp_stats_inc_failures(&pool->stats);
            tp_stats_dec_active_threads(&pool->stats);

            if (info->has_current_task) {
                handle_completion(pool, &info->current_task, 0);
            }

            info->has_current_task = 0;
            continue;
        }

        tp_status_t status = tp_task_queue_pop(&pool->task_queue, &task);

        if (status == TP_CLOSED) {
            break;
        }

        if (status != TP_OK) {
            continue;
        }

        memcpy(&info->current_task, &task, sizeof(tp_task_t));
        info->has_current_task = 1;

        tp_stats_inc_active_threads(&pool->stats);

        tp_stats_dec_queued(&pool->stats);

        task.func(task.arg);

        tp_stats_inc_completed(&pool->stats);

        tp_stats_dec_active_threads(&pool->stats);

        handle_completion(pool, &task, 1);

        info->has_current_task = 0;
    }

    pthread_mutex_lock(&pool->pool_mutex);
    info->running = 0;

    size_t active = 0;
    for (size_t i = 0; i < pool->num_workers; i++) {
        if (pool->workers[i].running) {
            active++;
        }
    }

    if (active == 0) {
        completion_queue_close(&pool->completion_queue);
        pthread_cond_broadcast(&pool->shutdown_cond);
    }
    pthread_mutex_unlock(&pool->pool_mutex);

    return NULL;
}

tp_status_t tp_thread_pool_init(tp_thread_pool_t *pool, const tp_config_t *config)
{
    if (!pool || !config) {
        return TP_INVALID_ARG;
    }

    if (config->num_threads == 0 || config->queue_capacity == 0) {
        return TP_INVALID_ARG;
    }

    memset(pool, 0, sizeof(tp_thread_pool_t));

    if (tp_stats_init(&pool->stats) != TP_OK) {
        return TP_ERROR;
    }

    if (tp_task_queue_init(&pool->task_queue, config->queue_capacity) != TP_OK) {
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    if (completion_queue_init(&pool->completion_queue, config->queue_capacity * 2) != TP_OK) {
        tp_task_queue_destroy(&pool->task_queue);
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    if (pthread_mutex_init(&pool->pool_mutex, NULL) != 0) {
        completion_queue_destroy(&pool->completion_queue);
        tp_task_queue_destroy(&pool->task_queue);
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    if (pthread_cond_init(&pool->shutdown_cond, NULL) != 0) {
        pthread_mutex_destroy(&pool->pool_mutex);
        completion_queue_destroy(&pool->completion_queue);
        tp_task_queue_destroy(&pool->task_queue);
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    pool->num_workers = config->num_threads;
    pool->queue_capacity = config->queue_capacity;
    pool->reject_policy = config->reject_policy;
    pool->state = TP_RUNNING;

    pool->workers = (tp_worker_info_t *)calloc(config->num_threads, sizeof(tp_worker_info_t));
    if (!pool->workers) {
        pthread_cond_destroy(&pool->shutdown_cond);
        pthread_mutex_destroy(&pool->pool_mutex);
        completion_queue_destroy(&pool->completion_queue);
        tp_task_queue_destroy(&pool->task_queue);
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    for (size_t i = 0; i < config->num_threads; i++) {
        pool->workers[i].index = i;
        pool->workers[i].pool = pool;
        pool->workers[i].running = 1;
        pool->workers[i].has_current_task = 0;

        if (pthread_create(&pool->workers[i].thread_id, NULL, worker_thread, &pool->workers[i]) != 0) {
            for (size_t j = 0; j < i; j++) {
                pool->workers[j].running = 0;
            }
            tp_task_queue_close(&pool->task_queue);
            for (size_t j = 0; j < i; j++) {
                pthread_join(pool->workers[j].thread_id, NULL);
            }
            free(pool->workers);
            pthread_cond_destroy(&pool->shutdown_cond);
            pthread_mutex_destroy(&pool->pool_mutex);
            completion_queue_destroy(&pool->completion_queue);
            tp_task_queue_destroy(&pool->task_queue);
            tp_stats_destroy(&pool->stats);
            return TP_ERROR;
        }
    }

    return TP_OK;
}

void tp_thread_pool_destroy(tp_thread_pool_t *pool)
{
    if (!pool) return;

    if (pool->state != TP_STOPPED) {
        tp_thread_pool_shutdown(pool, TP_FORCE_SHUTDOWN, 0);
    }

    if (pool->workers) {
        for (size_t i = 0; i < pool->num_workers; i++) {
            if (pool->workers[i].thread_id) {
                pthread_join(pool->workers[i].thread_id, NULL);
            }
        }
        free(pool->workers);
        pool->workers = NULL;
    }

    pthread_cond_destroy(&pool->shutdown_cond);
    pthread_mutex_destroy(&pool->pool_mutex);
    completion_queue_destroy(&pool->completion_queue);
    tp_task_queue_destroy(&pool->task_queue);
    tp_stats_destroy(&pool->stats);
}

tp_status_t tp_thread_pool_submit(tp_thread_pool_t *pool, const tp_task_t *task)
{
    if (!pool || !task || !task->func) {
        return TP_INVALID_ARG;
    }

    pthread_mutex_lock(&pool->pool_mutex);
    if (pool->state != TP_RUNNING) {
        pthread_mutex_unlock(&pool->pool_mutex);
        return TP_CLOSED;
    }
    pthread_mutex_unlock(&pool->pool_mutex);

    int block = (pool->reject_policy == TP_REJECT_BLOCK);
    tp_status_t status = tp_task_queue_push(&pool->task_queue, task, block);

    if (status == TP_OK) {
        tp_stats_inc_submitted(&pool->stats);
        tp_stats_inc_queued(&pool->stats);
    } else if (status == TP_FULL) {
        tp_stats_inc_rejected(&pool->stats);
    }

    return status;
}

tp_status_t tp_thread_pool_submit_simple(tp_thread_pool_t *pool, tp_task_func_t func, void *arg)
{
    if (!pool || !func) {
        return TP_INVALID_ARG;
    }

    tp_task_t task = TP_TASK_INIT(func, arg, NULL, TP_CALLBACK_IN_WORKER);
    return tp_thread_pool_submit(pool, &task);
}

tp_status_t tp_thread_pool_shutdown(tp_thread_pool_t *pool, tp_shutdown_mode_t mode, int timeout_ms)
{
    if (!pool) {
        return TP_INVALID_ARG;
    }

    pthread_mutex_lock(&pool->pool_mutex);

    if (pool->state == TP_STOPPED) {
        pthread_mutex_unlock(&pool->pool_mutex);
        return TP_OK;
    }

    pool->state = TP_SHUTDOWN;

    if (mode == TP_FORCE_SHUTDOWN) {
        tp_task_queue_close(&pool->task_queue);
        pool->state = TP_STOPPED;
        pthread_mutex_unlock(&pool->pool_mutex);
        return TP_OK;
    }

    tp_task_queue_close(&pool->task_queue);

    struct timespec ts;
    int use_timeout = (timeout_ms > 0);

    if (use_timeout) {
        clock_gettime(CLOCK_REALTIME, &ts);
        ts.tv_sec += timeout_ms / 1000;
        ts.tv_nsec += (timeout_ms % 1000) * 1000000L;
        if (ts.tv_nsec >= 1000000000L) {
            ts.tv_sec++;
            ts.tv_nsec -= 1000000000L;
        }
    }

    tp_status_t result = TP_OK;

    while (1) {
        size_t active = 0;
        for (size_t i = 0; i < pool->num_workers; i++) {
            if (pool->workers[i].running) {
                active++;
            }
        }

        if (active == 0) {
            pool->state = TP_STOPPED;
            break;
        }

        if (use_timeout) {
            int ret = pthread_cond_timedwait(&pool->shutdown_cond, &pool->pool_mutex, &ts);
            if (ret == ETIMEDOUT) {
                pool->state = TP_STOPPED;
                result = TP_TIMEOUT;
                break;
            }
        } else {
            pthread_cond_wait(&pool->shutdown_cond, &pool->pool_mutex);
        }
    }

    pthread_mutex_unlock(&pool->pool_mutex);
    return result;
}

tp_status_t tp_thread_pool_get_stats(tp_thread_pool_t *pool, tp_stats_snapshot_t *snapshot)
{
    if (!pool || !snapshot) {
        return TP_INVALID_ARG;
    }

    tp_stats_snapshot(&pool->stats, snapshot);
    snapshot->queued_tasks = tp_task_queue_size(&pool->task_queue);
    return TP_OK;
}

tp_state_t tp_thread_pool_get_state(tp_thread_pool_t *pool)
{
    if (!pool) return TP_STOPPED;

    pthread_mutex_lock(&pool->pool_mutex);
    tp_state_t state = pool->state;
    pthread_mutex_unlock(&pool->pool_mutex);

    return state;
}

size_t tp_thread_pool_get_queue_size(tp_thread_pool_t *pool)
{
    if (!pool) return 0;
    return tp_task_queue_size(&pool->task_queue);
}
