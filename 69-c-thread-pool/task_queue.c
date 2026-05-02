#include "thread_pool.h"
#include <stdlib.h>
#include <errno.h>

void task_queue_init_internal(task_queue_t *queue)
{
    queue->front = NULL;
    queue->rear = NULL;
    queue->count = 0;
}

int task_queue_push_internal(task_queue_t *queue, task_func_t func, void *arg, task_future_t *future)
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

    return 0;
}

int task_queue_pop_internal(task_queue_t *queue, task_func_t *func, void **arg, task_future_t **future)
{
    if (!queue || queue->count == 0)
        return EINVAL;

    task_node_t *node = queue->front;
    queue->front = node->next;

    if (queue->front == NULL)
        queue->rear = NULL;

    queue->count--;

    if (func) *func = node->func;
    if (arg) *arg = node->arg;
    if (future) *future = node->future;

    free(node);
    return 0;
}

int task_queue_size_internal(task_queue_t *queue)
{
    return queue->count;
}

void task_queue_clear_internal(task_queue_t *queue)
{
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
}
