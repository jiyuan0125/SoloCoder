#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include "thread_pool.h"
#include "stats.h"

typedef struct {
    int task_id;
    int compute_time_ms;
    volatile int *counter;
    pthread_mutex_t *print_mutex;
} task_context_t;

static void print_stats(const char *prefix, tp_stats_snapshot_t *stats)
{
    printf("%s Stats: submitted=%lu, completed=%lu, rejected=%lu, failures=%lu, "
           "active_threads=%zu, queued=%zu\n",
           prefix,
           (unsigned long)stats->total_submitted,
           (unsigned long)stats->total_completed,
           (unsigned long)stats->total_rejected,
           (unsigned long)stats->total_failures,
           stats->active_threads,
           stats->queued_tasks);
}

static void task_completion(void *arg, int success)
{
    task_context_t *ctx = (task_context_t *)arg;

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Callback] Task #%d completed %s\n",
           ctx->task_id, success ? "successfully" : "with failure");
    pthread_mutex_unlock(ctx->print_mutex);
}

static void compute_task(void *arg)
{
    task_context_t *ctx = (task_context_t *)arg;
    pthread_t tid = pthread_self();

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Start  ] Task #%d running in thread %lu\n",
           ctx->task_id, (unsigned long)tid);
    pthread_mutex_unlock(ctx->print_mutex);

    if (ctx->compute_time_ms > 0) {
        usleep(ctx->compute_time_ms * 1000);
    }

    __sync_fetch_and_add(ctx->counter, 1);

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Finish ] Task #%d finished in thread %lu\n",
           ctx->task_id, (unsigned long)tid);
    pthread_mutex_unlock(ctx->print_mutex);
}

static void crashing_task(void *arg)
{
    task_context_t *ctx = (task_context_t *)arg;
    pthread_t tid = pthread_self();

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Crash  ] Task #%d about to crash in thread %lu\n",
           ctx->task_id, (unsigned long)tid);
    pthread_mutex_unlock(ctx->print_mutex);

    int *invalid_ptr = NULL;
    *invalid_ptr = 42;

    (void)invalid_ptr;
}

int main(void)
{
    printf("========================================\n");
    printf("  Thread Pool Demo - 4 threads, 20 tasks\n");
    printf("========================================\n\n");

    tp_config_t config = {
        .num_threads = 4,
        .queue_capacity = 32,
        .reject_policy = TP_REJECT_BLOCK
    };

    tp_thread_pool_t pool;
    tp_status_t status = tp_thread_pool_init(&pool, &config);

    if (status != TP_OK) {
        fprintf(stderr, "Failed to initialize thread pool: %d\n", status);
        return 1;
    }

    printf("Thread pool created with %zu threads, queue capacity %zu\n\n",
           config.num_threads, config.queue_capacity);

    pthread_mutex_t print_mutex;
    pthread_mutex_init(&print_mutex, NULL);

    volatile int counter = 0;
    const int num_tasks = 20;
    task_context_t contexts[num_tasks];

    printf("Submitting %d tasks...\n\n", num_tasks);

    for (int i = 0; i < num_tasks; i++) {
        contexts[i].task_id = i + 1;
        contexts[i].compute_time_ms = 100 + (i % 3) * 50;
        contexts[i].counter = (int *)&counter;
        contexts[i].print_mutex = &print_mutex;

        if (i == 8 || i == 15) {
            tp_task_t task = {
                .func = crashing_task,
                .arg = &contexts[i],
                .completion = task_completion,
                .callback_mode = TP_CALLBACK_IN_WORKER
            };
            status = tp_thread_pool_submit(&pool, &task);
            pthread_mutex_lock(&print_mutex);
            printf("[Submit ] Task #%d (CRASH TEST) submitted, status: %d\n", i + 1, status);
            pthread_mutex_unlock(&print_mutex);
        } else {
            tp_task_t task = {
                .func = compute_task,
                .arg = &contexts[i],
                .completion = (i % 4 == 0) ? task_completion : NULL,
                .callback_mode = TP_CALLBACK_IN_WORKER
            };
            status = tp_thread_pool_submit(&pool, &task);
            pthread_mutex_lock(&print_mutex);
            printf("[Submit ] Task #%d submitted, status: %d\n", i + 1, status);
            pthread_mutex_unlock(&print_mutex);
        }
    }

    printf("\nAll tasks submitted. Tasks are executing concurrently...\n\n");

    tp_stats_snapshot_t stats;
    for (int i = 0; i < 5; i++) {
        sleep(1);
        tp_thread_pool_get_stats(&pool, &stats);
        print_stats("[Status]", &stats);
    }

    printf("\nInitiating graceful shutdown (waiting for all tasks to complete)...\n");

    status = tp_thread_pool_shutdown(&pool, TP_GRACEFUL_SHUTDOWN, 5000);

    if (status == TP_TIMEOUT) {
        printf("Shutdown timed out! Forcing shutdown...\n");
    } else if (status == TP_OK) {
        printf("Graceful shutdown completed successfully.\n");
    }

    tp_thread_pool_get_stats(&pool, &stats);
    print_stats("\n[Final ]", &stats);

    printf("\nTotal tasks executed (counter): %d\n", counter);
    printf("Expected: %d (minus 2 crashing tasks)\n", num_tasks - 2);

    tp_thread_pool_destroy(&pool);
    pthread_mutex_destroy(&print_mutex);

    printf("\n========================================\n");
    printf("  Demo Complete\n");
    printf("========================================\n");

    return 0;
}
