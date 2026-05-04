#ifndef WORKER_H
#define WORKER_H

#include "priority_queue.h"
#include <pthread.h>
#include <stdint.h>

#define MAX_WORKERS 64
#define MAX_TASK_MAP_SIZE 10000

typedef struct WorkerPool WorkerPool;

typedef struct WorkerInfo {
    pthread_t thread_id;
    int worker_index;
    int is_running;
    Task *current_task;
    uint64_t tasks_completed;
    WorkerPool *pool;
} WorkerInfo;

typedef struct TaskMapEntry {
    TaskId id;
    Task *task;
    int in_use;
} TaskMapEntry;

typedef struct WorkerPool {
    PriorityQueue *queue;
    WorkerInfo workers[MAX_WORKERS];
    int num_workers;
    int is_shutting_down;
    
    TaskMapEntry task_map[MAX_TASK_MAP_SIZE];
    pthread_mutex_t task_map_mutex;
    
    uint64_t total_tasks_completed;
    uint64_t total_tasks_failed;
    
    pthread_mutex_t stats_mutex;
    pthread_cond_t all_done;
} WorkerPool;

int worker_pool_init(WorkerPool *pool, PriorityQueue *queue, int num_workers);

void worker_pool_destroy(WorkerPool *pool);

TaskId worker_pool_submit_task(WorkerPool *pool, const char *name, 
                                TaskPriority priority, 
                                void *(*task_func)(void *arg), 
                                void *arg);

TaskStatus worker_pool_get_task_status(WorkerPool *pool, TaskId task_id);

int worker_pool_get_task_result(WorkerPool *pool, TaskId task_id, TaskResult *result);

int worker_pool_get_pending_count(WorkerPool *pool);

int worker_pool_get_pending_count_by_priority(WorkerPool *pool, TaskPriority priority);

int worker_pool_get_running_tasks(WorkerPool *pool, Task **tasks, int max_tasks);

uint64_t worker_pool_get_completed_count(WorkerPool *pool);

uint64_t worker_pool_get_failed_count(WorkerPool *pool);

void worker_pool_shutdown(WorkerPool *pool);

int worker_pool_is_shutdown(WorkerPool *pool);

void worker_pool_wait_for_completion(WorkerPool *pool);

const char* worker_pool_get_worker_info(WorkerPool *pool, int worker_index, char *buffer, int buffer_size);

#endif
