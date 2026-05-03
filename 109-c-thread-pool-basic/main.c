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
           "active_threads=%lu, queued=%lu\n",
           prefix,
           (unsigned long)stats->total_submitted,
           (unsigned long)stats->total_completed,
           (unsigned long)stats->total_rejected,
           (unsigned long)stats->total_failures,
           (unsigned long)stats->active_threads,
           stats->queued_tasks);
}

static void task_completion(void *arg, int success)
{
    task_context_t *ctx = (task_context_t *)arg;

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Callback] Task #%d completed %s\n",
           ctx->task_id, success ? "successfully" : "with FAILURE");
    pthread_mutex_unlock(ctx->print_mutex);
}

static void deferred_completion(void *arg, int success)
{
    task_context_t *ctx = (task_context_t *)arg;

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Deferred] Task #%d result: %s (will be polled later)\n",
           ctx->task_id, success ? "success" : "FAILURE");
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

static void long_running_task(void *arg)
{
    task_context_t *ctx = (task_context_t *)arg;
    pthread_t tid = pthread_self();

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Long   ] Task #%d (500ms) starting in thread %lu\n",
           ctx->task_id, (unsigned long)tid);
    pthread_mutex_unlock(ctx->print_mutex);

    usleep(500 * 1000);

    __sync_fetch_and_add(ctx->counter, 1);

    pthread_mutex_lock(ctx->print_mutex);
    printf("[Long   ] Task #%d finished\n", ctx->task_id);
    pthread_mutex_unlock(ctx->print_mutex);
}

int main(void)
{
    printf("========================================\n");
    printf("  Thread Pool Demo - Bug Fixes Verified\n");
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

    printf("=== Test 1: Verify active_threads semantics ===\n");
    printf("Submitting 4 long-running tasks (500ms each)...\n");
    printf("active_threads should go to 4, then back to 0\n\n");

    for (int i = 0; i < 4; i++) {
        contexts[i].task_id = i + 1;
        contexts[i].compute_time_ms = 500;
        contexts[i].counter = (int *)&counter;
        contexts[i].print_mutex = &print_mutex;

        tp_task_t task = {
            .func = long_running_task,
            .arg = &contexts[i],
            .completion = NULL,
            .callback_mode = TP_CALLBACK_IN_WORKER
        };
        tp_thread_pool_submit(&pool, &task);
    }

    usleep(100 * 1000);
    tp_stats_snapshot_t stats;
    tp_thread_pool_get_stats(&pool, &stats);
    print_stats("[During]", &stats);
    printf("Expected: active_threads=4 (all 4 threads busy)\n\n");

    printf("Waiting for tasks to complete...\n");
    sleep(2);
    tp_thread_pool_get_stats(&pool, &stats);
    print_stats("[After ]", &stats);
    printf("Expected: active_threads=0 (all threads idle)\n\n");

    printf("=== Test 2: Verify crash callback is triggered ===\n");
    printf("Submitting 2 crashing tasks with completion callbacks...\n\n");

    for (int i = 4; i < 6; i++) {
        contexts[i].task_id = i + 1;
        contexts[i].compute_time_ms = 100;
        contexts[i].counter = (int *)&counter;
        contexts[i].print_mutex = &print_mutex;

        tp_task_t task = {
            .func = crashing_task,
            .arg = &contexts[i],
            .completion = task_completion,
            .callback_mode = TP_CALLBACK_IN_WORKER
        };
        tp_thread_pool_submit(&pool, &task);
        printf("[Submit ] Task #%d (CRASH TEST) submitted\n", i + 1);
    }

    sleep(1);
    tp_thread_pool_get_stats(&pool, &stats);
    print_stats("[Status]", &stats);
    printf("Expected: failures=2, and you should have seen [Callback] ... with FAILURE above\n\n");

    printf("=== Test 3: Verify TP_CALLBACK_DEFERRED mode ===\n");
    printf("Submitting 4 tasks with DEFERRED callbacks...\n\n");

    for (int i = 6; i < 10; i++) {
        contexts[i].task_id = i + 1;
        contexts[i].compute_time_ms = 100;
        contexts[i].counter = (int *)&counter;
        contexts[i].print_mutex = &print_mutex;

        tp_task_t task = {
            .func = compute_task,
            .arg = &contexts[i],
            .completion = deferred_completion,
            .callback_mode = TP_CALLBACK_DEFERRED
        };
        tp_thread_pool_submit(&pool, &task);
        printf("[Submit ] Task #%d (DEFERRED) submitted\n", i + 1);
    }

    sleep(1);

    size_t comp_count = tp_thread_pool_completion_count(&pool);
    printf("\nCompletion queue has %zu items waiting to be polled\n\n", comp_count);

    printf("Now polling completion queue...\n");
    tp_completion_t comp;
    int polled = 0;
    while (tp_thread_pool_poll_completion(&pool, &comp, 0) == TP_OK) {
        pthread_mutex_lock(&print_mutex);
        printf("[Polled ] Completion: arg=%p, success=%d, calling callback...\n",
               comp.arg, comp.success);
        pthread_mutex_unlock(&print_mutex);
        comp.completion(comp.arg, comp.success);
        polled++;
    }
    printf("\nPolled %d completions from queue\n\n", polled);

    printf("=== Test 4: Submit remaining tasks to verify pool still works ===\n");
    printf("Submitting 10 more normal tasks...\n\n");

    for (int i = 10; i < 20; i++) {
        contexts[i].task_id = i + 1;
        contexts[i].compute_time_ms = 50 + (i % 3) * 25;
        contexts[i].counter = (int *)&counter;
        contexts[i].print_mutex = &print_mutex;

        tp_task_t task = {
            .func = compute_task,
            .arg = &contexts[i],
            .completion = (i % 4 == 0) ? task_completion : NULL,
            .callback_mode = TP_CALLBACK_IN_WORKER
        };
        tp_thread_pool_submit(&pool, &task);
    }

    sleep(2);
    tp_thread_pool_get_stats(&pool, &stats);
    print_stats("[Final ]", &stats);

    printf("\nTotal tasks executed (counter): %d\n", counter);
    printf("Expected: 4(long) + 0(crash) + 4(deferred) + 10(normal) = 18\n\n");

    printf("=== Graceful Shutdown ===\n");
    status = tp_thread_pool_shutdown(&pool, TP_GRACEFUL_SHUTDOWN, 5000);
    if (status == TP_OK) {
        printf("Graceful shutdown completed successfully.\n");
    } else {
        printf("Shutdown status: %d\n", status);
    }

    tp_thread_pool_destroy(&pool);
    pthread_mutex_destroy(&print_mutex);

    printf("\n========================================\n");
    printf("  Demo Complete\n");
    printf("========================================\n");

    return 0;
}
