#ifndef TASK_QUEUE_H
#define TASK_QUEUE_H

#include <pthread.h>
#include <stdint.h>

typedef enum {
    TASK_PRIORITY_LOW = 0,
    TASK_PRIORITY_MEDIUM = 1,
    TASK_PRIORITY_HIGH = 2
} TaskPriority;

typedef enum {
    TASK_STATUS_PENDING = 0,
    TASK_STATUS_RUNNING = 1,
    TASK_STATUS_COMPLETED = 2,
    TASK_STATUS_FAILED = 3,
    TASK_STATUS_CANCELLED = 4
} TaskStatus;

typedef uint64_t TaskId;

typedef struct TaskResult {
    int success;
    int error_code;
    char error_message[256];
    void *data;
} TaskResult;

typedef struct Task {
    TaskId id;
    TaskPriority priority;
    TaskStatus status;
    char name[64];
    void *(*task_func)(void *arg);
    void *arg;
    TaskResult result;
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    struct Task *next;
} Task;

typedef struct TaskQueue {
    Task *head;
    Task *tail;
    int size;
    int capacity;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    pthread_cond_t not_full;
} TaskQueue;

TaskId task_queue_generate_id(void);

Task* task_create(const char *name, TaskPriority priority, 
                  void *(*task_func)(void *arg), void *arg);

void task_destroy(Task *task);

void task_set_failed(Task *task, int error_code, const char *error_message);

void task_set_completed(Task *task, void *data);

TaskStatus task_get_status(Task *task);

int task_get_result(Task *task, TaskResult *result);

int task_queue_init(TaskQueue *queue, int capacity);

void task_queue_destroy(TaskQueue *queue);

int task_queue_enqueue(TaskQueue *queue, Task *task);

Task* task_queue_dequeue(TaskQueue *queue);

Task* task_queue_try_dequeue(TaskQueue *queue);

int task_queue_size(TaskQueue *queue);

int task_queue_is_empty(TaskQueue *queue);

int task_queue_is_full(TaskQueue *queue);

void task_queue_clear(TaskQueue *queue);

#endif
