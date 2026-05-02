#include "thread_pool.h"
#include <stdlib.h>
#include <errno.h>
#include <sys/time.h>
#include <string.h>

static void *worker_thread(void *arg);
static int create_worker(thread_pool_t *pool, int index);

typedef struct {
    thread_pool_t *pool;
    int index;
} worker_arg_t;

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

    ret = pthread_cond_init(&pool->all_idle, NULL);
    if (ret != 0)
    {
        pthread_mutex_destroy(&pool->mutex);
        return ret;
    }

    ret = task_queue_init(&pool->queue);
    if (ret != 0)
    {
        pthread_cond_destroy(&pool->all_idle);
        pthread_mutex_destroy(&pool->mutex);
        return ret;
    }

    pool->threads = (pthread_t *)calloc(max_threads, sizeof(pthread_t));
    if (!pool->threads)
    {
        task_queue_destroy(&pool->queue);
        pthread_cond_destroy(&pool->all_idle);
        pthread_mutex_destroy(&pool->mutex);
        return ENOMEM;
    }

    for (int i = 0; i < min_threads; i++)
    {
        ret = create_worker(pool, i);
        if (ret != 0)
        {
            pool->is_shutting_down = true;
            pthread_cond_broadcast(&pool->queue.not_empty);
            
            for (int j = 0; j < pool->current_threads; j++)
            {
                if (pool->threads[j] != 0)
                {
                    pthread_join(pool->threads[j], NULL);
                }
            }
            
            free(pool->threads);
            task_queue_destroy(&pool->queue);
            pthread_cond_destroy(&pool->all_idle);
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
    pool->current_threads++;
    pool->idle_threads++;

    return 0;
}

static void *worker_thread(void *arg)
{
    worker_arg_t *worker_arg = (worker_arg_t *)arg;
    thread_pool_t *pool = worker_arg->pool;
    int index = worker_arg->index;
    free(arg);

    while (true)
    {
        pthread_mutex_lock(&pool->mutex);

        while (pool->queue.count == 0 && !pool->is_shutting_down)
        {
            struct timespec ts;
            struct timeval tv;
            gettimeofday(&tv, NULL);
            ts.tv_sec = tv.tv_sec + IDLE_TIMEOUT_SECONDS;
            ts.tv_nsec = tv.tv_usec * 1000;

            int ret = pthread_cond_timedwait(&pool->queue.not_empty, &pool->mutex, &ts);

            if (ret == ETIMEDOUT && !pool->is_shutting_down)
            {
                if (pool->current_threads > pool->min_threads)
                {
                    pool->current_threads--;
                    pool->idle_threads--;
                    pool->threads[index] = 0;

                    if (pool->active_threads == 0 && pool->idle_threads == 0)
                    {
                        pthread_cond_broadcast(&pool->all_idle);
                    }

                    pthread_mutex_unlock(&pool->mutex);
                    return NULL;
                }
            }
        }

        if (pool->is_shutting_down && pool->queue.count == 0)
        {
            pool->current_threads--;
            pool->idle_threads--;
            pool->threads[index] = 0;

            if (pool->active_threads == 0 && pool->current_threads == 0)
            {
                pthread_cond_broadcast(&pool->all_idle);
            }

            pthread_mutex_unlock(&pool->mutex);
            return NULL;
        }

        task_func_t func = NULL;
        void *task_arg = NULL;
        task_future_t *future = NULL;

        task_node_t *node = pool->queue.front;
        pool->queue.front = node->next;
        if (pool->queue.front == NULL)
            pool->queue.rear = NULL;
        pool->queue.count--;

        func = node->func;
        task_arg = node->arg;
        future = node->future;
        free(node);

        pool->idle_threads--;
        pool->active_threads++;

        pthread_mutex_unlock(&pool->mutex);

        void *result = NULL;
        if (func)
        {
            result = func(task_arg);
        }

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

    pthread_mutex_unlock(&pool->mutex);

    return task_queue_push(&pool->queue, func, arg, future);
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
    pthread_cond_broadcast(&pool->queue.not_empty);

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

        task_node_t *node = pool->queue.front;
        while (node != NULL)
        {
            if (node->future)
            {
                future_set_canceled(node->future);
            }
            node = node->next;
        }
    }

    pthread_mutex_unlock(&pool->mutex);

    for (int i = 0; i < pool->max_threads; i++)
    {
        if (pool->threads[i] != 0)
        {
            pthread_join(pool->threads[i], NULL);
            pool->threads[i] = 0;
        }
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

    task_queue_destroy(&pool->queue);

    if (pool->threads)
    {
        free(pool->threads);
        pool->threads = NULL;
    }

    pthread_cond_destroy(&pool->all_idle);
    pthread_mutex_destroy(&pool->mutex);
}
