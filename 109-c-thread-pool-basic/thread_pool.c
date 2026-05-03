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

static void* worker_thread(void *arg)
{
    tp_worker_info_t *info = (tp_worker_info_t *)arg;
    tp_thread_pool_t *pool = info->pool;
    tp_task_t task;
    int task_success;

    current_worker = info;

    setup_crash_handlers();

    pthread_mutex_lock(&pool->pool_mutex);
    tp_stats_inc_active_threads(&pool->stats);
    pthread_mutex_unlock(&pool->pool_mutex);

    while (1) {
        pthread_mutex_lock(&pool->pool_mutex);
        if (pool->state == TP_STOPPED) {
            pthread_mutex_unlock(&pool->pool_mutex);
            break;
        }
        pthread_mutex_unlock(&pool->pool_mutex);

        task_success = 1;
        if (sigsetjmp(info->recovery_env, 1) != 0) {
            task_success = 0;
            tp_stats_inc_failures(&pool->stats);
            continue;
        }

        tp_status_t status = tp_task_queue_pop(&pool->task_queue, &task);

        if (status == TP_CLOSED) {
            break;
        }

        if (status != TP_OK) {
            continue;
        }

        tp_stats_dec_queued(&pool->stats);

        task.func(task.arg);

        tp_stats_inc_completed(&pool->stats);

        if (task.completion && task.callback_mode == TP_CALLBACK_IN_WORKER) {
            task.completion(task.arg, task_success);
        }
    }

    pthread_mutex_lock(&pool->pool_mutex);
    tp_stats_dec_active_threads(&pool->stats);
    pool->workers[info->index].running = 0;

    size_t active = 0;
    for (size_t i = 0; i < pool->num_workers; i++) {
        if (pool->workers[i].running) {
            active++;
        }
    }

    if (active == 0) {
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

    if (pthread_mutex_init(&pool->pool_mutex, NULL) != 0) {
        tp_task_queue_destroy(&pool->task_queue);
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    if (pthread_cond_init(&pool->shutdown_cond, NULL) != 0) {
        pthread_mutex_destroy(&pool->pool_mutex);
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
        tp_task_queue_destroy(&pool->task_queue);
        tp_stats_destroy(&pool->stats);
        return TP_ERROR;
    }

    for (size_t i = 0; i < config->num_threads; i++) {
        pool->workers[i].index = i;
        pool->workers[i].pool = pool;
        pool->workers[i].running = 1;

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
