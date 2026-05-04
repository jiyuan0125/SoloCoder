#include "task_queue.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static pthread_mutex_t g_id_mutex = PTHREAD_MUTEX_INITIALIZER;
static TaskId g_next_task_id = 1;

TaskId task_queue_generate_id(void) {
    TaskId id;
    pthread_mutex_lock(&g_id_mutex);
    id = g_next_task_id++;
    pthread_mutex_unlock(&g_id_mutex);
    return id;
}

Task* task_create(const char *name, TaskPriority priority, 
                  void *(*task_func)(void *arg), void *arg) {
    Task *task = (Task*)malloc(sizeof(Task));
    if (task == NULL) {
        return NULL;
    }
    
    memset(task, 0, sizeof(Task));
    task->id = task_queue_generate_id();
    task->priority = priority;
    task->status = TASK_STATUS_PENDING;
    task->task_func = task_func;
    task->arg = arg;
    task->next = NULL;
    
    if (name != NULL) {
        strncpy(task->name, name, sizeof(task->name) - 1);
        task->name[sizeof(task->name) - 1] = '\0';
    }
    
    pthread_mutex_init(&task->mutex, NULL);
    pthread_cond_init(&task->cond, NULL);
    
    return task;
}

void task_destroy(Task *task) {
    if (task == NULL) {
        return;
    }
    
    pthread_mutex_destroy(&task->mutex);
    pthread_cond_destroy(&task->cond);
    free(task);
}

void task_set_failed(Task *task, int error_code, const char *error_message) {
    if (task == NULL) {
        return;
    }
    
    pthread_mutex_lock(&task->mutex);
    task->status = TASK_STATUS_FAILED;
    task->result.success = 0;
    task->result.error_code = error_code;
    if (error_message != NULL) {
        strncpy(task->result.error_message, error_message, 
                sizeof(task->result.error_message) - 1);
        task->result.error_message[sizeof(task->result.error_message) - 1] = '\0';
    }
    pthread_cond_broadcast(&task->cond);
    pthread_mutex_unlock(&task->mutex);
}

void task_set_completed(Task *task, void *data) {
    if (task == NULL) {
        return;
    }
    
    pthread_mutex_lock(&task->mutex);
    task->status = TASK_STATUS_COMPLETED;
    task->result.success = 1;
    task->result.error_code = 0;
    task->result.error_message[0] = '\0';
    task->result.data = data;
    pthread_cond_broadcast(&task->cond);
    pthread_mutex_unlock(&task->mutex);
}

TaskStatus task_get_status(Task *task) {
    if (task == NULL) {
        return TASK_STATUS_FAILED;
    }
    
    TaskStatus status;
    pthread_mutex_lock(&task->mutex);
    status = task->status;
    pthread_mutex_unlock(&task->mutex);
    return status;
}

int task_get_result(Task *task, TaskResult *result) {
    if (task == NULL || result == NULL) {
        return -1;
    }
    
    pthread_mutex_lock(&task->mutex);
    while (task->status == TASK_STATUS_PENDING || task->status == TASK_STATUS_RUNNING) {
        pthread_cond_wait(&task->cond, &task->mutex);
    }
    
    *result = task->result;
    pthread_mutex_unlock(&task->mutex);
    return 0;
}

int task_queue_init(TaskQueue *queue, int capacity) {
    if (queue == NULL || capacity <= 0) {
        return -1;
    }
    
    queue->head = NULL;
    queue->tail = NULL;
    queue->size = 0;
    queue->capacity = capacity;
    
    if (pthread_mutex_init(&queue->mutex, NULL) != 0) {
        return -1;
    }
    
    if (pthread_cond_init(&queue->not_empty, NULL) != 0) {
        pthread_mutex_destroy(&queue->mutex);
        return -1;
    }
    
    if (pthread_cond_init(&queue->not_full, NULL) != 0) {
        pthread_cond_destroy(&queue->not_empty);
        pthread_mutex_destroy(&queue->mutex);
        return -1;
    }
    
    return 0;
}

void task_queue_destroy(TaskQueue *queue) {
    if (queue == NULL) {
        return;
    }
    
    task_queue_clear(queue);
    
    pthread_mutex_destroy(&queue->mutex);
    pthread_cond_destroy(&queue->not_empty);
    pthread_cond_destroy(&queue->not_full);
}

int task_queue_enqueue(TaskQueue *queue, Task *task) {
    if (queue == NULL || task == NULL) {
        return -1;
    }
    
    pthread_mutex_lock(&queue->mutex);
    
    while (queue->size >= queue->capacity) {
        pthread_cond_wait(&queue->not_full, &queue->mutex);
    }
    
    if (queue->tail == NULL) {
        queue->head = task;
        queue->tail = task;
    } else {
        queue->tail->next = task;
        queue->tail = task;
    }
    
    queue->size++;
    task->next = NULL;
    
    pthread_cond_signal(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);
    
    return 0;
}

Task* task_queue_dequeue(TaskQueue *queue) {
    if (queue == NULL) {
        return NULL;
    }
    
    pthread_mutex_lock(&queue->mutex);
    
    while (queue->head == NULL) {
        pthread_cond_wait(&queue->not_empty, &queue->mutex);
    }
    
    Task *task = queue->head;
    queue->head = task->next;
    queue->size--;
    
    if (queue->head == NULL) {
        queue->tail = NULL;
    }
    
    task->next = NULL;
    pthread_cond_signal(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
    
    return task;
}

Task* task_queue_try_dequeue(TaskQueue *queue) {
    if (queue == NULL) {
        return NULL;
    }
    
    pthread_mutex_lock(&queue->mutex);
    
    if (queue->head == NULL) {
        pthread_mutex_unlock(&queue->mutex);
        return NULL;
    }
    
    Task *task = queue->head;
    queue->head = task->next;
    queue->size--;
    
    if (queue->head == NULL) {
        queue->tail = NULL;
    }
    
    task->next = NULL;
    pthread_cond_signal(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
    
    return task;
}

int task_queue_size(TaskQueue *queue) {
    if (queue == NULL) {
        return -1;
    }
    
    int size;
    pthread_mutex_lock(&queue->mutex);
    size = queue->size;
    pthread_mutex_unlock(&queue->mutex);
    return size;
}

int task_queue_is_empty(TaskQueue *queue) {
    if (queue == NULL) {
        return 1;
    }
    
    int empty;
    pthread_mutex_lock(&queue->mutex);
    empty = (queue->size == 0);
    pthread_mutex_unlock(&queue->mutex);
    return empty;
}

int task_queue_is_full(TaskQueue *queue) {
    if (queue == NULL) {
        return 1;
    }
    
    int full;
    pthread_mutex_lock(&queue->mutex);
    full = (queue->size >= queue->capacity);
    pthread_mutex_unlock(&queue->mutex);
    return full;
}

void task_queue_clear(TaskQueue *queue) {
    if (queue == NULL) {
        return;
    }
    
    pthread_mutex_lock(&queue->mutex);
    
    while (queue->head != NULL) {
        Task *task = queue->head;
        queue->head = task->next;
        
        pthread_mutex_lock(&task->mutex);
        task->status = TASK_STATUS_CANCELLED;
        pthread_cond_broadcast(&task->cond);
        pthread_mutex_unlock(&task->mutex);
        
        task_destroy(task);
    }
    
    queue->tail = NULL;
    queue->size = 0;
    
    pthread_cond_broadcast(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
}
