#ifndef MIN_HEAP_H
#define MIN_HEAP_H

#include <time.h>
#include <stddef.h>

#define TASK_NAME_LEN 64

typedef unsigned long long task_id_t;
typedef void (*task_callback_t)(void *arg);

typedef enum {
    TASK_ONCE,
    TASK_PERIODIC
} task_type_t;

typedef struct timer_task {
    task_id_t id;
    char name[TASK_NAME_LEN];
    task_type_t type;
    struct timespec expire_time;
    long interval_ms;
    int max_executions;
    int executed_count;
    task_callback_t callback;
    void *arg;
    int cancelled;
} timer_task_t;

typedef struct min_heap {
    timer_task_t **data;
    size_t capacity;
    size_t size;
} min_heap_t;

min_heap_t* min_heap_create(size_t initial_capacity);
void min_heap_destroy(min_heap_t *heap);
int min_heap_insert(min_heap_t *heap, timer_task_t *task);
timer_task_t* min_heap_extract_min(min_heap_t *heap);
timer_task_t* min_heap_peek(min_heap_t *heap);
int min_heap_remove(min_heap_t *heap, task_id_t id);
size_t min_heap_size(min_heap_t *heap);
int min_heap_is_empty(min_heap_t *heap);

int timespec_compare(const struct timespec *a, const struct timespec *b);
struct timespec timespec_add_ms(struct timespec ts, long ms);
long timespec_diff_ms(struct timespec a, struct timespec b);

#endif
