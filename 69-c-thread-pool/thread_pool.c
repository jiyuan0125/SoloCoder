#include "thread_pool.h"
#include <stdlib.h>
#include <errno.h>
#include <sys/time.h>
#include <string.h>

static void *worker_thread(void *arg);
static int create_worker(thread_pool_t *pool, int index);
static void mutex_cleanup_handler(void *arg);
static void task_cleanup_handler(void *arg);

typedef struct {
    thread_pool_t *pool;
    int index;
} worker_arg_t;

typedef struct {
    task_future_t *future;
} task_cleanup_ctx_t;

void task_queue_init_internal(task_queue_t *queue);
int task_queue_push_internal(task_queue_t *queue, task_func_t func, void *arg, task_future_t *future);
int task_queue_pop_internal(task_queue_t *queue, task_func_t *func, void **arg, task_future_t **future);
int task_queue_size_internal(task_queue_t *queue);
void task_queue_clear_internal(task_queue_t *queue);

static void mutex_cleanup_handler(void *arg)
{
    pthread_mutex_t *mutex = (pthread_mutex_t *)arg;
    pthread_mutex_unlock(mutex);
}

static void task_cleanup_handler(void *arg)
{
    task_cleanup_ctx_t *ctx = (task_cleanup_ctx_t *)arg;
    if (ctx->future)
    {
        future_set_canceled(ctx->future);
    }
}

int thread_pool_create(thread_pool_t *pool, int min_threads, int max_threads)
{
    if (!pool || min_threads < 0 || max_threads < min_threads)
        return EINVAL;

    memset(pool, 0, sizeof(thread_pool_t));

    pool->min_threads = min_threads;
    pool->max_threads = max_threads;
    pool->current_threads = 0;
    pool->active_threads = 0;
    pool->idle_threads = 0;
    pool->is_shutting_down = false;
    pool->is_shutdown = false;

    int ret = pthread_mutex_init(&pool->mutex, NULL);
    if (ret != 0)
        return ret;

    ret = pthread_cond_init(&pool->queue_not_empty, NULL);
    if (ret != 0)
    {
        pthread_mutex_destroy(&pool->mutex);
        return ret;
    }

    ret = pthread_cond_init(&pool->all_idle, NULL);
    if (ret != 0)
    {
        pthread_cond_destroy(&pool->queue_not_empty);
        pthread_mutex_destroy(&pool->mutex);
        return ret;
    }

    task_queue_init_internal(&pool->queue);

    pool->threads = (pthread_t *)calloc(max_threads, sizeof(pthread_t));
    if (!pool->threads)
    {
        pthread_cond_destroy(&pool->all_idle);
        pthread_cond_destroy(&pool->queue_not_empty);
        pthread_mutex_destroy(&pool->mutex);
        return ENOMEM;
    }

    for (int i = 0; i < min_threads; i++)
    {
        ret = create_worker(pool, i);
        if (ret != 0)
        {
            pthread_mutex_lock(&pool->mutex);
            pool->is_shutting_down = true;
            pthread_cond_broadcast(&pool->queue_not_empty);
            pthread_mutex_unlock(&pool->mutex);

            for (int j = 0; j < pool->current_threads; j++)
            {
                if (pool->threads[j] != 0)
                {
                    pthread_cancel(pool->threads[j]);
                    pthread_join(pool->threads[j], NULL);
                }
            }

            free(pool->threads);
            pthread_cond_destroy(&pool->all_idle);
            pthread_cond_destroy(&pool->queue_not_empty);
            pthread_mutex_destroy(&pool->mutex);
            return ret;
        }
    }

    return 0;
}

static int create_worker(thread_pool_t *pool, int index)
{
    worker_arg_t *arg = (worker_arg_t *)malloc(sizeof(worker_arg_t));
    if (!arg)
        return ENOMEM;

    arg->pool = pool;
    arg->index = index;

    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    int ret = pthread_create(&pool->threads[index], &attr, worker_thread, arg);
    if (ret != 0)
    {
        free(arg);
        pthread_attr_destroy(&attr);
        return ret;
    }

    pthread_attr_destroy(&attr);
    return 0;
}

static void *worker_thread(void *arg)
{
    worker_arg_t *worker_arg = (worker_arg_t *)arg;
    thread_pool_t *pool = worker_arg->pool;
    int index = worker_arg->index;
    free(arg);

    pthread_setcanceltype(PTHREAD_CANCEL_DEFERRED, NULL);

    pthread_mutex_lock(&pool->mutex);
    pool->current_threads++;
    pool->idle_threads++;
    pthread_mutex_unlock(&pool->mutex);

    while (true)
    {
        task_func_t func = NULL;
        void *task_arg = NULL;
        task_future_t *future = NULL;
        int should_exit = 0;

        pthread_mutex_lock(&pool->mutex);
        pthread_cleanup_push(mutex_cleanup_handler, &pool->mutex);

        while (pool->queue.count == 0 && !pool->is_shutting_down)
        {
            struct timespec ts;
            struct timeval tv;
            gettimeofday(&tv, NULL);
            ts.tv_sec = tv.tv_sec + IDLE_TIMEOUT_SECONDS;
            ts.tv_nsec = tv.tv_usec * 1000;

            int ret = pthread_cond_timedwait(&pool->queue_not_empty, &pool->mutex, &ts);

            if (ret == ETIMEDOUT && !pool->is_shutting_down)
            {
                if (pool->current_threads > pool->min_threads)
                {
                    should_exit = 1;
                    pool->current_threads--;
                    pool->idle_threads--;
                    pool->threads[index] = 0;

                    if (pool->active_threads == 0 && pool->current_threads == 0)
                    {
                        pthread_cond_broadcast(&pool->all_idle);
                    }
                    break;
                }
            }
        }

        if (!should_exit)
        {
            if (pool->is_shutting_down && pool->queue.count == 0)
            {
                should_exit = 1;
                pool->current_threads--;
                pool->idle_threads--;
                pool->threads[index] = 0;

                if (pool->active_threads == 0 && pool->current_threads == 0)
                {
                    pthread_cond_broadcast(&pool->all_idle);
                }
            }
            else if (pool->queue.count > 0)
            {
                task_queue_pop_internal(&pool->queue, &func, &task_arg, &future);
                pool->idle_threads--;
                pool->active_threads++;
            }
        }

        pthread_cleanup_pop(1);

        if (should_exit)
        {
            return NULL;
        }

        void *result = NULL;
        task_cleanup_ctx_t tc;
        tc.future = future;
        pthread_cleanup_push(task_cleanup_handler, &tc);

        if (func)
        {
            result = func(task_arg);
        }

        pthread_cleanup_pop(0);

        if (future)
        {
            future_set_result(future, result);
        }

        pthread_mutex_lock(&pool->mutex);
        pool->active_threads--;
        pool->idle_threads++;

        if (pool->active_threads == 0 && pool->queue.count == 0)
        {
            pthread_cond_broadcast(&pool->all_idle);
        }
        pthread_mutex_unlock(&pool->mutex);
    }

    return NULL;
}

int thread_pool_submit(thread_pool_t *pool, task_func_t func, void *arg, task_future_t *future)
{
    if (!pool || !func)
        return EINVAL;

    pthread_mutex_lock(&pool->mutex);

    if (pool->is_shutting_down || pool->is_shutdown)
    {
        pthread_mutex_unlock(&pool->mutex);
        return ESHUTDOWN;
    }

    int queue_size = pool->queue.count;
    int max_queue = pool->max_threads * MAX_QUEUE_MULTIPLIER;

    if (queue_size >= max_queue && pool->current_threads < pool->max_threads)
    {
        for (int i = 0; i < pool->max_threads; i++)
        {
            if (pool->threads[i] == 0)
            {
                int ret = create_worker(pool, i);
                if (ret != 0)
                {
                    pthread_mutex_unlock(&pool->mutex);
                    return ret;
                }
                break;
            }
        }
    }

    int ret = task_queue_push_internal(&pool->queue, func, arg, future);
    if (ret == 0)
    {
        pthread_cond_signal(&pool->queue_not_empty);
    }

    pthread_mutex_unlock(&pool->mutex);
    return ret;
}

int thread_pool_shutdown(thread_pool_t *pool, int timeout_ms)
{
    if (!pool)
        return EINVAL;

    pthread_mutex_lock(&pool->mutex);

    if (pool->is_shutdown)
    {
        pthread_mutex_unlock(&pool->mutex);
        return 0;
    }

    pool->is_shutting_down = true;
    pthread_cond_broadcast(&pool->queue_not_empty);

    int unfinished = 0;

    if (timeout_ms == -1)
    {
        while (pool->active_threads > 0 || pool->queue.count > 0)
        {
            pthread_cond_wait(&pool->all_idle, &pool->mutex);
        }
        unfinished = pool->queue.count;
    }
    else
    {
        struct timespec ts;
        struct timeval tv;
        gettimeofday(&tv, NULL);
        ts.tv_sec = tv.tv_sec;
        ts.tv_nsec = tv.tv_usec * 1000;

        ts.tv_sec += timeout_ms / 1000;
        ts.tv_nsec += (timeout_ms % 1000) * 1000000;

        if (ts.tv_nsec >= 1000000000)
        {
            ts.tv_sec++;
            ts.tv_nsec -= 1000000000;
        }

        while (pool->active_threads > 0 || pool->queue.count > 0)
        {
            int ret = pthread_cond_timedwait(&pool->all_idle, &pool->mutex, &ts);
            if (ret == ETIMEDOUT)
            {
                break;
            }
        }

        unfinished = pool->queue.count;
        task_queue_clear_internal(&pool->queue);
    }

    pthread_t threads_to_cancel[64];
    int thread_indices[64];
    int cancel_count = 0;
    for (int i = 0; i < pool->max_threads && cancel_count < 64; i++)
    {
        if (pool->threads[i] != 0)
        {
            threads_to_cancel[cancel_count] = pool->threads[i];
            thread_indices[cancel_count] = i;
            cancel_count++;
        }
    }

    pthread_mutex_unlock(&pool->mutex);

    for (int i = 0; i < cancel_count; i++)
    {
        pthread_t tid = threads_to_cancel[i];
        int idx = thread_indices[i];
        pthread_cancel(tid);
        pthread_join(tid, NULL);
        pthread_mutex_lock(&pool->mutex);
        pool->threads[idx] = 0;
        pthread_mutex_unlock(&pool->mutex);
    }

    pthread_mutex_lock(&pool->mutex);
    pool->is_shutdown = true;
    pthread_mutex_unlock(&pool->mutex);

    if (unfinished > 0)
        return unfinished;

    return 0;
}

void thread_pool_destroy(thread_pool_t *pool)
{
    if (!pool)
        return;

    if (!pool->is_shutdown)
    {
        thread_pool_shutdown(pool, -1);
    }

    pthread_mutex_lock(&pool->mutex);
    task_queue_clear_internal(&pool->queue);
    pthread_mutex_unlock(&pool->mutex);

    if (pool->threads)
    {
        free(pool->threads);
        pool->threads = NULL;
    }

    pthread_cond_destroy(&pool->queue_not_empty);
    pthread_cond_destroy(&pool->all_idle);
    pthread_mutex_destroy(&pool->mutex);
}
