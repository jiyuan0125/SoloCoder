#include "scheduler.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>

static struct timespec get_current_time(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts;
}

static timer_task_t* create_task(task_id_t id,
                                  const char *name,
                                  task_type_t type,
                                  struct timespec expire_time,
                                  long interval_ms,
                                  int max_executions,
                                  task_callback_t callback,
                                  void *arg) {
    timer_task_t *task = (timer_task_t *)malloc(sizeof(timer_task_t));
    if (task == NULL) {
        return NULL;
    }
    
    task->id = id;
    task->type = type;
    task->expire_time = expire_time;
    task->interval_ms = interval_ms;
    task->max_executions = max_executions;
    task->executed_count = 0;
    task->callback = callback;
    task->arg = arg;
    task->cancelled = 0;
    
    if (name != NULL) {
        strncpy(task->name, name, TASK_NAME_LEN - 1);
        task->name[TASK_NAME_LEN - 1] = '\0';
    } else {
        task->name[0] = '\0';
    }
    
    return task;
}

static void* scheduler_thread_func(void *arg) {
    scheduler_t *sched = (scheduler_t *)arg;
    struct timespec current, wait_time;
    long time_to_next;
    
    while (1) {
        pthread_mutex_lock(&sched->heap_mutex);
        
        while (min_heap_is_empty(sched->heap) && sched->running) {
            sched->current_check_interval_ms = MAX_CHECK_INTERVAL_MS;
            wait_time.tv_sec = sched->current_check_interval_ms / 1000;
            wait_time.tv_nsec = (sched->current_check_interval_ms % 1000) * 1000000L;
            pthread_cond_timedwait(&sched->scheduler_cond, &sched->heap_mutex, &wait_time);
        }
        
        if (!sched->running) {
            pthread_mutex_unlock(&sched->heap_mutex);
            break;
        }
        
        current = get_current_time();
        timer_task_t *next_task = min_heap_peek(sched->heap);
        
        if (next_task != NULL) {
            time_to_next = timespec_diff_ms(next_task->expire_time, current);
            
            if (time_to_next <= 0) {
                timer_task_t *expired_task = min_heap_extract_min(sched->heap);
                
                if (expired_task->type == TASK_PERIODIC) {
                    if (expired_task->max_executions == 0 || 
                        expired_task->executed_count < expired_task->max_executions - 1) {
                        
                        timer_task_t *repeated_task = create_task(
                            expired_task->id,
                            expired_task->name,
                            expired_task->type,
                            timespec_add_ms(current, expired_task->interval_ms),
                            expired_task->interval_ms,
                            expired_task->max_executions,
                            expired_task->callback,
                            expired_task->arg
                        );
                        
                        if (repeated_task != NULL) {
                            repeated_task->executed_count = expired_task->executed_count + 1;
                            min_heap_insert(sched->heap, repeated_task);
                        }
                    }
                }
                
                thread_pool_submit(sched->pool, expired_task);
                
                sched->current_check_interval_ms = MIN_CHECK_INTERVAL_MS;
            } else {
                if (time_to_next < HIGH_PRIORITY_THRESHOLD_MS) {
                    sched->current_check_interval_ms = MIN_CHECK_INTERVAL_MS;
                } else if (time_to_next < MAX_CHECK_INTERVAL_MS) {
                    sched->current_check_interval_ms = time_to_next / 2;
                    if (sched->current_check_interval_ms < MIN_CHECK_INTERVAL_MS) {
                        sched->current_check_interval_ms = MIN_CHECK_INTERVAL_MS;
                    }
                } else {
                    sched->current_check_interval_ms = MAX_CHECK_INTERVAL_MS;
                }
                
                wait_time.tv_sec = sched->current_check_interval_ms / 1000;
                wait_time.tv_nsec = (sched->current_check_interval_ms % 1000) * 1000000L;
                pthread_cond_timedwait(&sched->scheduler_cond, &sched->heap_mutex, &wait_time);
            }
        } else {
            sched->current_check_interval_ms = MAX_CHECK_INTERVAL_MS;
            wait_time.tv_sec = sched->current_check_interval_ms / 1000;
            wait_time.tv_nsec = (sched->current_check_interval_ms % 1000) * 1000000L;
            pthread_cond_timedwait(&sched->scheduler_cond, &sched->heap_mutex, &wait_time);
        }
        
        pthread_mutex_unlock(&sched->heap_mutex);
    }
    
    return NULL;
}

scheduler_t* scheduler_create(size_t thread_count, size_t max_queue_size) {
    if (thread_count == 0) {
        thread_count = SCHED_DEFAULT_THREADS;
    }
    if (max_queue_size == 0) {
        max_queue_size = SCHED_DEFAULT_QUEUE_SIZE;
    }
    
    scheduler_t *sched = (scheduler_t *)malloc(sizeof(scheduler_t));
    if (sched == NULL) {
        return NULL;
    }
    
    sched->heap = min_heap_create(64);
    if (sched->heap == NULL) {
        free(sched);
        return NULL;
    }
    
    sched->pool = thread_pool_create(thread_count, max_queue_size);
    if (sched->pool == NULL) {
        min_heap_destroy(sched->heap);
        free(sched);
        return NULL;
    }
    
    if (pthread_mutex_init(&sched->heap_mutex, NULL) != 0) {
        thread_pool_destroy(sched->pool);
        min_heap_destroy(sched->heap);
        free(sched);
        return NULL;
    }
    
    if (pthread_cond_init(&sched->scheduler_cond, NULL) != 0) {
        pthread_mutex_destroy(&sched->heap_mutex);
        thread_pool_destroy(sched->pool);
        min_heap_destroy(sched->heap);
        free(sched);
        return NULL;
    }
    
    sched->next_task_id = 1;
    sched->current_check_interval_ms = MAX_CHECK_INTERVAL_MS;
    sched->running = 1;
    
    if (pthread_create(&sched->scheduler_thread, NULL, scheduler_thread_func, sched) != 0) {
        pthread_cond_destroy(&sched->scheduler_cond);
        pthread_mutex_destroy(&sched->heap_mutex);
        thread_pool_destroy(sched->pool);
        min_heap_destroy(sched->heap);
        free(sched);
        return NULL;
    }
    
    return sched;
}

void scheduler_destroy(scheduler_t *sched) {
    if (sched == NULL) {
        return;
    }
    
    pthread_mutex_lock(&sched->heap_mutex);
    sched->running = 0;
    pthread_cond_signal(&sched->scheduler_cond);
    pthread_mutex_unlock(&sched->heap_mutex);
    
    pthread_join(sched->scheduler_thread, NULL);
    
    min_heap_destroy(sched->heap);
    thread_pool_destroy(sched->pool);
    
    pthread_mutex_destroy(&sched->heap_mutex);
    pthread_cond_destroy(&sched->scheduler_cond);
    
    free(sched);
}

task_id_t scheduler_add_once(scheduler_t *sched, 
                              const char *name,
                              long delay_ms,
                              task_callback_t callback,
                              void *arg) {
    if (sched == NULL || callback == NULL) {
        return 0;
    }
    
    pthread_mutex_lock(&sched->heap_mutex);
    
    task_id_t id = sched->next_task_id++;
    struct timespec current = get_current_time();
    struct timespec expire = timespec_add_ms(current, delay_ms);
    
    timer_task_t *task = create_task(id, name, TASK_ONCE, expire, 0, 1, callback, arg);
    if (task == NULL) {
        pthread_mutex_unlock(&sched->heap_mutex);
        return 0;
    }
    
    if (min_heap_insert(sched->heap, task) != 0) {
        free(task);
        pthread_mutex_unlock(&sched->heap_mutex);
        return 0;
    }
    
    pthread_cond_signal(&sched->scheduler_cond);
    pthread_mutex_unlock(&sched->heap_mutex);
    
    return id;
}

task_id_t scheduler_add_periodic(scheduler_t *sched,
                                  const char *name,
                                  long initial_delay_ms,
                                  long interval_ms,
                                  int max_executions,
                                  task_callback_t callback,
                                  void *arg) {
    if (sched == NULL || callback == NULL || interval_ms <= 0) {
        return 0;
    }
    
    pthread_mutex_lock(&sched->heap_mutex);
    
    task_id_t id = sched->next_task_id++;
    struct timespec current = get_current_time();
    struct timespec expire = timespec_add_ms(current, initial_delay_ms);
    
    timer_task_t *task = create_task(id, name, TASK_PERIODIC, expire, 
                                      interval_ms, max_executions, callback, arg);
    if (task == NULL) {
        pthread_mutex_unlock(&sched->heap_mutex);
        return 0;
    }
    
    if (min_heap_insert(sched->heap, task) != 0) {
        free(task);
        pthread_mutex_unlock(&sched->heap_mutex);
        return 0;
    }
    
    pthread_cond_signal(&sched->scheduler_cond);
    pthread_mutex_unlock(&sched->heap_mutex);
    
    return id;
}

int scheduler_cancel(scheduler_t *sched, task_id_t task_id) {
    if (sched == NULL || task_id == 0) {
        return -1;
    }
    
    pthread_mutex_lock(&sched->heap_mutex);
    
    timer_task_t *task = NULL;
    size_t heap_size = min_heap_size(sched->heap);
    timer_task_t **tasks_copy = NULL;
    
    if (heap_size > 0) {
        tasks_copy = (timer_task_t **)malloc(heap_size * sizeof(timer_task_t *));
        if (tasks_copy == NULL) {
            pthread_mutex_unlock(&sched->heap_mutex);
            return -1;
        }
        
        for (size_t i = 0; i < heap_size; i++) {
            tasks_copy[i] = min_heap_extract_min(sched->heap);
        }
        
        for (size_t i = 0; i < heap_size; i++) {
            if (tasks_copy[i]->id == task_id) {
                tasks_copy[i]->cancelled = 1;
                task = tasks_copy[i];
            }
        }
        
        for (size_t i = 0; i < heap_size; i++) {
            min_heap_insert(sched->heap, tasks_copy[i]);
        }
        
        free(tasks_copy);
    }
    
    pthread_cond_signal(&sched->scheduler_cond);
    pthread_mutex_unlock(&sched->heap_mutex);
    
    return (task != NULL) ? 0 : -1;
}

size_t scheduler_task_count(scheduler_t *sched) {
    if (sched == NULL) {
        return 0;
    }
    pthread_mutex_lock(&sched->heap_mutex);
    size_t count = min_heap_size(sched->heap);
    pthread_mutex_unlock(&sched->heap_mutex);
    return count;
}
