#include "thread_pool.h"
#include <errno.h>
#include <sys/time.h>

int future_init(task_future_t *future)
{
    if (!future)
        return EINVAL;

    future->result = NULL;
    future->status = 0;
    future->is_ready = false;
    future->is_canceled = false;

    int ret = pthread_mutex_init(&future->mutex, NULL);
    if (ret != 0)
        return ret;

    ret = pthread_cond_init(&future->cond, NULL);
    if (ret != 0)
    {
        pthread_mutex_destroy(&future->mutex);
        return ret;
    }

    return 0;
}

int future_set_result(task_future_t *future, void *result)
{
    if (!future)
        return EINVAL;

    pthread_mutex_lock(&future->mutex);

    if (!future->is_ready && !future->is_canceled)
    {
        future->result = result;
        future->is_ready = true;
        future->status = 0;
        pthread_cond_broadcast(&future->cond);
    }

    pthread_mutex_unlock(&future->mutex);
    return 0;
}

int future_set_canceled(task_future_t *future)
{
    if (!future)
        return EINVAL;

    pthread_mutex_lock(&future->mutex);

    if (!future->is_ready)
    {
        future->is_canceled = true;
        future->status = ECANCELED;
        pthread_cond_broadcast(&future->cond);
    }

    pthread_mutex_unlock(&future->mutex);
    return 0;
}

int task_future_get(task_future_t *future, int timeout_ms)
{
    if (!future)
        return EINVAL;

    pthread_mutex_lock(&future->mutex);

    int ret = 0;
    if (!future->is_ready && !future->is_canceled)
    {
        if (timeout_ms == -1)
        {
            while (!future->is_ready && !future->is_canceled)
            {
                pthread_cond_wait(&future->cond, &future->mutex);
            }
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

            while (!future->is_ready && !future->is_canceled)
            {
                ret = pthread_cond_timedwait(&future->cond, &future->mutex, &ts);
                if (ret == ETIMEDOUT)
                {
                    pthread_mutex_unlock(&future->mutex);
                    return ETIMEDOUT;
                }
            }
        }
    }

    if (future->is_canceled)
    {
        ret = ECANCELED;
    }
    else
    {
        ret = future->status;
    }

    pthread_mutex_unlock(&future->mutex);
    return ret;
}

void future_destroy(task_future_t *future)
{
    if (!future)
        return;

    pthread_cond_destroy(&future->cond);
    pthread_mutex_destroy(&future->mutex);
}
