#ifndef THREAD_POOL_H
#define THREAD_POOL_H

#include <pthread.h>
#include <stdbool.h>
#include <errno.h>

#ifndef ESHUTDOWN
#define ESHUTDOWN 110
#endif

#ifndef ECANCELED
#define ECANCELED 125
#endif

#ifndef ETIMEDOUT
#define ETIMEDOUT 110
#endif

#define MAX_QUEUE_MULTIPLIER 10
#define IDLE_TIMEOUT_SECONDS 5

typedef void *(*task_func_t)(void *);

typedef struct task_future {
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    void *result;
    int status;
    bool is_ready;
    bool is_canceled;
} task_future_t;

typedef struct task_node {
    task_func_t func;
    void *arg;
    void *result;
    task_future_t *future;
    struct task_node *next;
} task_node_t;

typedef struct {
    task_node_t *front;
    task_node_t *rear;
    int count;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
} task_queue_t;

typedef struct {
    pthread_t *threads;
    int min_threads;
    int max_threads;
    int current_threads;
    int active_threads;
    int idle_threads;
    task_queue_t queue;
    bool is_shutting_down;
    bool is_shutdown;
    pthread_mutex_t mutex;
    pthread_cond_t all_idle;
} thread_pool_t;

int thread_pool_create(thread_pool_t *pool, int min_threads, int max_threads);
int thread_pool_submit(thread_pool_t *pool, task_func_t func, void *arg, task_future_t *future);
int task_future_get(task_future_t *future, int timeout_ms);
int thread_pool_shutdown(thread_pool_t *pool, int timeout_ms);
void thread_pool_destroy(thread_pool_t *pool);

int task_queue_init(task_queue_t *queue);
int task_queue_push(task_queue_t *queue, task_func_t func, void *arg, task_future_t *future);
int task_queue_pop(task_queue_t *queue, task_func_t *func, void **arg, task_future_t **future);
int task_queue_size(task_queue_t *queue);
void task_queue_destroy(task_queue_t *queue);

int future_init(task_future_t *future);
int future_set_result(task_future_t *future, void *result);
int future_set_canceled(task_future_t *future);
void future_destroy(task_future_t *future);

#endif
