#ifndef SCHEDULER_H
#define SCHEDULER_H

#include "min_heap.h"
#include "thread_pool.h"

#define SCHED_DEFAULT_THREADS 4
#define SCHED_DEFAULT_QUEUE_SIZE 128
#define MAX_CHECK_INTERVAL_MS 1000
#define MIN_CHECK_INTERVAL_MS 10
#define HIGH_PRIORITY_THRESHOLD_MS 500

typedef struct scheduler {
    min_heap_t *heap;
    thread_pool_t *pool;
    pthread_mutex_t heap_mutex;
    pthread_cond_t scheduler_cond;
    
    pthread_t scheduler_thread;
    int running;
    
    task_id_t next_task_id;
    long current_check_interval_ms;
} scheduler_t;

scheduler_t* scheduler_create(size_t thread_count, size_t max_queue_size);
void scheduler_destroy(scheduler_t *sched);

task_id_t scheduler_add_once(scheduler_t *sched, 
                              const char *name,
                              long delay_ms,
                              task_callback_t callback,
                              void *arg);

task_id_t scheduler_add_periodic(scheduler_t *sched,
                                  const char *name,
                                  long initial_delay_ms,
                                  long interval_ms,
                                  int max_executions,
                                  task_callback_t callback,
                                  void *arg);

int scheduler_cancel(scheduler_t *sched, task_id_t task_id);

size_t scheduler_task_count(scheduler_t *sched);

#endif
