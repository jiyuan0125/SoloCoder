#include "thread_pool.h"
#include <stdlib.h>
#include <errno.h>

int task_queue_init(task_queue_t *queue)
{
    if (!queue)
        return EINVAL;

    queue->front = NULL;
    queue->rear = NULL;
    queue->count = 0;

    int ret = pthread_mutex_init(&queue->mutex, NULL);
    if (ret != 0)
        return ret;

    ret = pthread_cond_init(&queue->not_empty, NULL);
    if (ret != 0)
    {
        pthread_mutex_destroy(&queue->mutex);
        return ret;
    }

    return 0;
}

int task_queue_push(task_queue_t *queue, task_func_t func, void *arg, task_future_t *future)
{
    if (!queue || !func)
        return EINVAL;

    task_node_t *node = (task_node_t *)malloc(sizeof(task_node_t));
    if (!node)
        return ENOMEM;

    node->func = func;
    node->arg = arg;
    node->future = future;
    node->next = NULL;

    pthread_mutex_lock(&queue->mutex);

    if (queue->rear == NULL)
    {
        queue->front = node;
        queue->rear = node;
    }
    else
    {
        queue->rear->next = node;
        queue->rear = node;
    }
    queue->count++;

    pthread_cond_signal(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);

    return 0;
}

int task_queue_pop(task_queue_t *queue, task_func_t *func, void **arg, task_future_t **future)
{
    if (!queue)
        return EINVAL;

    pthread_mutex_lock(&queue->mutex);

    while (queue->count == 0)
    {
        pthread_cond_wait(&queue->not_empty, &queue->mutex);
    }

    task_node_t *node = queue->front;
    queue->front = node->next;

    if (queue->front == NULL)
        queue->rear = NULL;

    queue->count--;

    if (func) *func = node->func;
    if (arg) *arg = node->arg;
    if (future) *future = node->future;

    free(node);
    pthread_mutex_unlock(&queue->mutex);

    return 0;
}

int task_queue_size(task_queue_t *queue)
{
    if (!queue)
        return -1;

    pthread_mutex_lock(&queue->mutex);
    int size = queue->count;
    pthread_mutex_unlock(&queue->mutex);

    return size;
}

void task_queue_destroy(task_queue_t *queue)
{
    if (!queue)
        return;

    pthread_mutex_lock(&queue->mutex);

    while (queue->front != NULL)
    {
        task_node_t *node = queue->front;
        queue->front = node->next;
        if (node->future)
        {
            future_set_canceled(node->future);
        }
        free(node);
    }

    queue->rear = NULL;
    queue->count = 0;

    pthread_mutex_unlock(&queue->mutex);

    pthread_cond_destroy(&queue->not_empty);
    pthread_mutex_destroy(&queue->mutex);
}
