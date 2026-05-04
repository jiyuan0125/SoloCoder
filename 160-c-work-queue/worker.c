#include "worker.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <stdarg.h>

static void log_message(const char *level, const char *format, ...) {
    va_list args;
    time_t now = time(NULL);
    struct tm *t = localtime(&now);
    
    fprintf(stderr, "[%04d-%02d-%02d %02d:%02d:%02d] [%s] ",
            t->tm_year + 1900, t->tm_mon + 1, t->tm_mday,
            t->tm_hour, t->tm_min, t->tm_sec, level);
    
    va_start(args, format);
    vfprintf(stderr, format, args);
    va_end(args);
    
    fprintf(stderr, "\n");
}

static void* worker_thread_func(void *arg);

static Task* task_map_find(WorkerPool *pool, TaskId id) {
    pthread_mutex_lock(&pool->task_map_mutex);
    for (int i = 0; i < MAX_TASK_MAP_SIZE; i++) {
        if (pool->task_map[i].in_use && pool->task_map[i].id == id) {
            Task *task = pool->task_map[i].task;
            pthread_mutex_unlock(&pool->task_map_mutex);
            return task;
        }
    }
    pthread_mutex_unlock(&pool->task_map_mutex);
    return NULL;
}

static int task_map_add(WorkerPool *pool, Task *task) {
    pthread_mutex_lock(&pool->task_map_mutex);
    for (int i = 0; i < MAX_TASK_MAP_SIZE; i++) {
        if (!pool->task_map[i].in_use) {
            pool->task_map[i].id = task->id;
            pool->task_map[i].task = task;
            pool->task_map[i].in_use = 1;
            pthread_mutex_unlock(&pool->task_map_mutex);
            return 0;
        }
    }
    pthread_mutex_unlock(&pool->task_map_mutex);
    return -1;
}

static void task_map_remove(WorkerPool *pool, TaskId id) {
    pthread_mutex_lock(&pool->task_map_mutex);
    for (int i = 0; i < MAX_TASK_MAP_SIZE; i++) {
        if (pool->task_map[i].in_use && pool->task_map[i].id == id) {
            pool->task_map[i].in_use = 0;
            pool->task_map[i].task = NULL;
            break;
        }
    }
    pthread_mutex_unlock(&pool->task_map_mutex);
}

int worker_pool_init(WorkerPool *pool, PriorityQueue *queue, int num_workers) {
    if (pool == NULL || queue == NULL || num_workers <= 0 || num_workers > MAX_WORKERS) {
        return -1;
    }
    
    memset(pool, 0, sizeof(WorkerPool));
    pool->queue = queue;
    pool->num_workers = num_workers;
    pool->is_shutting_down = 0;
    
    if (pthread_mutex_init(&pool->task_map_mutex, NULL) != 0) {
        return -1;
    }
    
    if (pthread_mutex_init(&pool->stats_mutex, NULL) != 0) {
        pthread_mutex_destroy(&pool->task_map_mutex);
        return -1;
    }
    
    if (pthread_cond_init(&pool->all_done, NULL) != 0) {
        pthread_mutex_destroy(&pool->stats_mutex);
        pthread_mutex_destroy(&pool->task_map_mutex);
        return -1;
    }
    
    memset(pool->task_map, 0, sizeof(pool->task_map));
    
    for (int i = 0; i < num_workers; i++) {
        pool->workers[i].worker_index = i;
        pool->workers[i].is_running = 1;
        pool->workers[i].current_task = NULL;
        pool->workers[i].tasks_completed = 0;
        pool->workers[i].pool = pool;
        
        if (pthread_create(&pool->workers[i].thread_id, NULL, 
                           worker_thread_func, &pool->workers[i]) != 0) {
            for (int j = 0; j < i; j++) {
                pool->workers[j].is_running = 0;
            }
            priority_queue_shutdown(queue);
            
            for (int j = 0; j < i; j++) {
                pthread_join(pool->workers[j].thread_id, NULL);
            }
            
            pthread_cond_destroy(&pool->all_done);
            pthread_mutex_destroy(&pool->stats_mutex);
            pthread_mutex_destroy(&pool->task_map_mutex);
            return -1;
        }
    }
    
    log_message("INFO", "Worker pool initialized with %d workers", num_workers);
    return 0;
}

static void* worker_thread_func(void *arg) {
    WorkerInfo *info = (WorkerInfo*)arg;
    WorkerPool *pool = info->pool;
    
    if (pool == NULL) {
        return NULL;
    }
    
    log_message("INFO", "Worker %d started", info->worker_index);
    
    while (info->is_running && !priority_queue_is_shutdown(pool->queue)) {
        Task *task = priority_queue_dequeue(pool->queue);
        
        if (task == NULL) {
            continue;
        }
        
        pthread_mutex_lock(&pool->stats_mutex);
        info->current_task = task;
        pthread_mutex_unlock(&pool->stats_mutex);
        
        pthread_mutex_lock(&task->mutex);
        task->status = TASK_STATUS_RUNNING;
        pthread_mutex_unlock(&task->mutex);
        
        log_message("INFO", "Worker %d executing task: %s (ID: %lu)", 
                   info->worker_index, task->name, (unsigned long)task->id);
        
        void *result_data = NULL;
        int failed = 0;
        int error_code = 0;
        char error_msg[256] = "";
        
        if (task->task_func != NULL) {
            result_data = task->task_func(task->arg);
            if (result_data == NULL && task->arg != NULL) {
                failed = 1;
                error_code = -1;
                snprintf(error_msg, sizeof(error_msg), "Task function returned NULL");
            }
        } else {
            failed = 1;
            error_code = -2;
            snprintf(error_msg, sizeof(error_msg), "Task function is NULL");
        }
        
        if (failed) {
            log_message("ERROR", "Task %s (ID: %lu) failed: %s (code: %d)",
                       task->name, (unsigned long)task->id, error_msg, error_code);
            task_set_failed(task, error_code, error_msg);
            
            pthread_mutex_lock(&pool->stats_mutex);
            pool->total_tasks_failed++;
            pthread_mutex_unlock(&pool->stats_mutex);
        } else {
            log_message("INFO", "Task %s (ID: %lu) completed successfully",
                       task->name, (unsigned long)task->id);
            task_set_completed(task, result_data);
            
            pthread_mutex_lock(&pool->stats_mutex);
            pool->total_tasks_completed++;
            info->tasks_completed++;
            pthread_mutex_unlock(&pool->stats_mutex);
        }
        
        pthread_mutex_lock(&pool->stats_mutex);
        info->current_task = NULL;
        pthread_mutex_unlock(&pool->stats_mutex);
        
        task_map_remove(pool, task->id);
        task_destroy(task);
    }
    
    log_message("INFO", "Worker %d exiting", info->worker_index);
    return NULL;
}

void worker_pool_destroy(WorkerPool *pool) {
    if (pool == NULL) {
        return;
    }
    
    worker_pool_shutdown(pool);
    worker_pool_wait_for_completion(pool);
    
    for (int i = 0; i < MAX_TASK_MAP_SIZE; i++) {
        if (pool->task_map[i].in_use && pool->task_map[i].task != NULL) {
            Task *task = pool->task_map[i].task;
            TaskStatus status = task_get_status(task);
            if (status == TASK_STATUS_PENDING || status == TASK_STATUS_CANCELLED) {
                task_destroy(task);
            }
            pool->task_map[i].in_use = 0;
            pool->task_map[i].task = NULL;
        }
    }
    
    pthread_cond_destroy(&pool->all_done);
    pthread_mutex_destroy(&pool->stats_mutex);
    pthread_mutex_destroy(&pool->task_map_mutex);
    
    log_message("INFO", "Worker pool destroyed");
}

TaskId worker_pool_submit_task(WorkerPool *pool, const char *name, 
                                TaskPriority priority, 
                                void *(*task_func)(void *arg), 
                                void *arg) {
    if (pool == NULL || task_func == NULL) {
        return 0;
    }
    
    if (pool->is_shutting_down) {
        log_message("WARN", "Cannot submit task: pool is shutting down");
        return 0;
    }
    
    if (priority_queue_is_full(pool->queue)) {
        log_message("WARN", "Task queue is full, dropping task: %s", 
                   name ? name : "unnamed");
        return 0;
    }
    
    Task *task = task_create(name, priority, task_func, arg);
    if (task == NULL) {
        log_message("ERROR", "Failed to create task");
        return 0;
    }
    
    if (task_map_add(pool, task) != 0) {
        log_message("WARN", "Task map is full, dropping task: %s", task->name);
        task_destroy(task);
        return 0;
    }
    
    int result = priority_queue_enqueue(pool->queue, task);
    if (result != 0) {
        log_message("WARN", "Failed to enqueue task: %s (code: %d)", 
                   task->name, result);
        task_map_remove(pool, task->id);
        task_destroy(task);
        return 0;
    }
    
    log_message("INFO", "Task submitted: %s (ID: %lu, priority: %d)", 
               task->name, (unsigned long)task->id, priority);
    
    return task->id;
}

TaskStatus worker_pool_get_task_status(WorkerPool *pool, TaskId task_id) {
    if (pool == NULL || task_id == 0) {
        return TASK_STATUS_FAILED;
    }
    
    Task *task = task_map_find(pool, task_id);
    if (task == NULL) {
        return TASK_STATUS_FAILED;
    }
    
    return task_get_status(task);
}

int worker_pool_get_task_result(WorkerPool *pool, TaskId task_id, TaskResult *result) {
    if (pool == NULL || task_id == 0 || result == NULL) {
        return -1;
    }
    
    Task *task = task_map_find(pool, task_id);
    if (task == NULL) {
        return -1;
    }
    
    return task_get_result(task, result);
}

int worker_pool_get_pending_count(WorkerPool *pool) {
    if (pool == NULL) {
        return -1;
    }
    
    return priority_queue_total_size(pool->queue);
}

int worker_pool_get_pending_count_by_priority(WorkerPool *pool, TaskPriority priority) {
    if (pool == NULL) {
        return -1;
    }
    
    return priority_queue_size_by_priority(pool->queue, priority);
}

int worker_pool_get_running_tasks(WorkerPool *pool, Task **tasks, int max_tasks) {
    if (pool == NULL || tasks == NULL || max_tasks <= 0) {
        return 0;
    }
    
    int count = 0;
    
    pthread_mutex_lock(&pool->stats_mutex);
    for (int i = 0; i < pool->num_workers && count < max_tasks; i++) {
        if (pool->workers[i].current_task != NULL) {
            tasks[count++] = pool->workers[i].current_task;
        }
    }
    pthread_mutex_unlock(&pool->stats_mutex);
    
    return count;
}

uint64_t worker_pool_get_completed_count(WorkerPool *pool) {
    if (pool == NULL) {
        return 0;
    }
    
    uint64_t count;
    pthread_mutex_lock(&pool->stats_mutex);
    count = pool->total_tasks_completed;
    pthread_mutex_unlock(&pool->stats_mutex);
    
    return count;
}

uint64_t worker_pool_get_failed_count(WorkerPool *pool) {
    if (pool == NULL) {
        return 0;
    }
    
    uint64_t count;
    pthread_mutex_lock(&pool->stats_mutex);
    count = pool->total_tasks_failed;
    pthread_mutex_unlock(&pool->stats_mutex);
    
    return count;
}

void worker_pool_shutdown(WorkerPool *pool) {
    if (pool == NULL) {
        return;
    }
    
    if (pool->is_shutting_down) {
        return;
    }
    
    log_message("INFO", "Worker pool shutting down...");
    pool->is_shutting_down = 1;
    
    priority_queue_shutdown(pool->queue);
    priority_queue_clear(pool->queue);
}

int worker_pool_is_shutdown(WorkerPool *pool) {
    if (pool == NULL) {
        return 1;
    }
    
    return pool->is_shutting_down;
}

void worker_pool_wait_for_completion(WorkerPool *pool) {
    if (pool == NULL) {
        return;
    }
    
    for (int i = 0; i < pool->num_workers; i++) {
        pool->workers[i].is_running = 0;
    }
    
    priority_queue_shutdown(pool->queue);
    
    for (int i = 0; i < pool->num_workers; i++) {
        if (pool->workers[i].thread_id != 0) {
            pthread_join(pool->workers[i].thread_id, NULL);
            pool->workers[i].thread_id = 0;
        }
    }
    
    log_message("INFO", "All workers have completed");
}

const char* worker_pool_get_worker_info(WorkerPool *pool, int worker_index, 
                                         char *buffer, int buffer_size) {
    if (pool == NULL || buffer == NULL || buffer_size <= 0 || 
        worker_index < 0 || worker_index >= pool->num_workers) {
        if (buffer != NULL && buffer_size > 0) {
            buffer[0] = '\0';
        }
        return NULL;
    }
    
    pthread_mutex_lock(&pool->stats_mutex);
    WorkerInfo *info = &pool->workers[worker_index];
    
    if (info->current_task != NULL) {
        snprintf(buffer, buffer_size, 
                "Worker %d: running task \"%s\" (ID: %lu), completed: %lu",
                worker_index, info->current_task->name, 
                (unsigned long)info->current_task->id,
                (unsigned long)info->tasks_completed);
    } else {
        snprintf(buffer, buffer_size, 
                "Worker %d: idle, completed: %lu",
                worker_index, (unsigned long)info->tasks_completed);
    }
    pthread_mutex_unlock(&pool->stats_mutex);
    
    return buffer;
}
