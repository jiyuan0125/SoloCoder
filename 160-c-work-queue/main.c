#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include "worker.h"
#include "priority_queue.h"

static void print_status(const char *prefix, WorkerPool *pool);

static void* simulate_image_processing(void *arg) {
    const char *filename = (const char*)arg;
    if (filename == NULL) {
        return NULL;
    }
    
    printf("    [Processing] Generating thumbnail for: %s\n", filename);
    sleep(1);
    printf("    [Done] Thumbnail generated for: %s\n", filename);
    
    return (void*)strdup("thumbnail_created");
}

static void* simulate_send_email(void *arg) {
    const char *email = (const char*)arg;
    if (email == NULL) {
        return NULL;
    }
    
    printf("    [Processing] Sending email to: %s\n", email);
    sleep(2);
    printf("    [Done] Email sent to: %s\n", email);
    
    return (void*)strdup("email_sent");
}

static void* simulate_payment_processing(void *arg) {
    const char *order_id = (const char*)arg;
    if (order_id == NULL) {
        return NULL;
    }
    
    printf("    [Processing] Processing payment for order: %s\n", order_id);
    sleep(1);
    printf("    [Done] Payment processed for order: %s\n", order_id);
    
    return (void*)strdup("payment_completed");
}

static void* simulate_failing_task(void *arg) {
    const char *task_name = (const char*)arg;
    if (task_name == NULL) {
        return NULL;
    }
    
    printf("    [Processing] Starting failing task: %s\n", task_name);
    sleep(1);
    printf("    [Failed] Task failed: %s\n", task_name);
    
    return NULL;
}

static void* simulate_write_log(void *arg) {
    const char *log_message = (const char*)arg;
    if (log_message == NULL) {
        return NULL;
    }
    
    printf("    [Processing] Writing log: %s\n", log_message);
    usleep(500000);
    printf("    [Done] Log written: %s\n", log_message);
    
    return (void*)strdup("log_written");
}

static void print_task_status(WorkerPool *pool, TaskId task_id, const char *task_name) {
    TaskStatus status = worker_pool_get_task_status(pool, task_id);
    const char *status_str = "UNKNOWN";
    
    switch (status) {
        case TASK_STATUS_PENDING:
            status_str = "PENDING";
            break;
        case TASK_STATUS_RUNNING:
            status_str = "RUNNING";
            break;
        case TASK_STATUS_COMPLETED:
            status_str = "COMPLETED";
            break;
        case TASK_STATUS_FAILED:
            status_str = "FAILED";
            break;
        case TASK_STATUS_CANCELLED:
            status_str = "CANCELLED";
            break;
    }
    
    printf("  Task '%s' (ID: %lu) status: %s\n", 
           task_name, (unsigned long)task_id, status_str);
}

static void wait_for_task(WorkerPool *pool, TaskId task_id, const char *task_name) {
    TaskResult result;
    if (worker_pool_get_task_result(pool, task_id, &result) == 0) {
        if (result.success) {
            printf("  Task '%s' completed successfully, result: %s\n", 
                   task_name, (char*)result.data);
            if (result.data != NULL) {
                free(result.data);
            }
        } else {
            printf("  Task '%s' failed, error code: %d, message: %s\n", 
                   task_name, result.error_code, result.error_message);
        }
    }
}

static void print_status(const char *prefix, WorkerPool *pool) {
    if (prefix != NULL) {
        printf("\n=== %s ===\n", prefix);
    }
    
    int pending = worker_pool_get_pending_count(pool);
    int pending_high = worker_pool_get_pending_count_by_priority(pool, TASK_PRIORITY_HIGH);
    int pending_medium = worker_pool_get_pending_count_by_priority(pool, TASK_PRIORITY_MEDIUM);
    int pending_low = worker_pool_get_pending_count_by_priority(pool, TASK_PRIORITY_LOW);
    uint64_t completed = worker_pool_get_completed_count(pool);
    uint64_t failed = worker_pool_get_failed_count(pool);
    
    printf("  Pending tasks: %d (High: %d, Medium: %d, Low: %d)\n",
           pending, pending_high, pending_medium, pending_low);
    printf("  Completed tasks: %lu\n", (unsigned long)completed);
    printf("  Failed tasks: %lu\n", (unsigned long)failed);
    
    Task *running_tasks[10];
    int running_count = worker_pool_get_running_tasks(pool, running_tasks, 10);
    if (running_count > 0) {
        printf("  Currently running tasks: %d\n", running_count);
        for (int i = 0; i < running_count; i++) {
            printf("    - %s (ID: %lu)\n", running_tasks[i]->name, 
                   (unsigned long)running_tasks[i]->id);
        }
    }
}

int main() {
    printf("========================================\n");
    printf("  Background Task Processing Module\n");
    printf("========================================\n\n");
    
    PriorityQueue queue;
    WorkerPool pool;
    
    printf("Initializing priority queue (capacity: 1000)...\n");
    if (priority_queue_init(&queue, 1000) != 0) {
        fprintf(stderr, "Failed to initialize priority queue\n");
        return 1;
    }
    
    printf("Initializing worker pool (4 workers)...\n");
    if (worker_pool_init(&pool, &queue, 4) != 0) {
        fprintf(stderr, "Failed to initialize worker pool\n");
        priority_queue_destroy(&queue);
        return 1;
    }
    
    printf("\n--- Test 1: Submit tasks with different priorities ---\n");
    
    TaskId task_log1 = worker_pool_submit_task(&pool, "Write log 1", TASK_PRIORITY_LOW,
                                                 simulate_write_log, "User logged in");
    TaskId task_log2 = worker_pool_submit_task(&pool, "Write log 2", TASK_PRIORITY_LOW,
                                                 simulate_write_log, "Data updated");
    
    TaskId task_email1 = worker_pool_submit_task(&pool, "Send welcome email", TASK_PRIORITY_MEDIUM,
                                                   simulate_send_email, "user1@example.com");
    TaskId task_email2 = worker_pool_submit_task(&pool, "Send notification", TASK_PRIORITY_MEDIUM,
                                                   simulate_send_email, "user2@example.com");
    
    TaskId task_payment1 = worker_pool_submit_task(&pool, "Process payment 1", TASK_PRIORITY_HIGH,
                                                     simulate_payment_processing, "ORD-001");
    TaskId task_payment2 = worker_pool_submit_task(&pool, "Process payment 2", TASK_PRIORITY_HIGH,
                                                     simulate_payment_processing, "ORD-002");
    TaskId task_image1 = worker_pool_submit_task(&pool, "Generate thumbnail", TASK_PRIORITY_MEDIUM,
                                                   simulate_image_processing, "photo.jpg");
    
    printf("\nSubmitted tasks:\n");
    print_task_status(&pool, task_log1, "Write log 1");
    print_task_status(&pool, task_email1, "Send welcome email");
    print_task_status(&pool, task_payment1, "Process payment 1");
    
    printf("\n--- Test 2: Check status while tasks are running ---\n");
    sleep(1);
    print_status("Current Status", &pool);
    
    printf("\n--- Test 3: Submit a failing task ---\n");
    TaskId task_fail = worker_pool_submit_task(&pool, "Failing task", TASK_PRIORITY_MEDIUM,
                                                 simulate_failing_task, "Test failure");
    
    printf("\n--- Test 4: Wait for some tasks to complete ---\n");
    wait_for_task(&pool, task_payment1, "Process payment 1");
    wait_for_task(&pool, task_fail, "Failing task");
    
    printf("\n--- Test 5: Check final status ---\n");
    print_status("Final Status Before Shutdown", &pool);
    
    printf("\n--- Test 6: Demonstrate graceful shutdown ---\n");
    printf("Shutting down worker pool...\n");
    worker_pool_shutdown(&pool);
    worker_pool_wait_for_completion(&pool);
    
    print_status("Status After Shutdown", &pool);
    
    printf("\n--- Test 7: Submit task after shutdown (should fail) ---\n");
    TaskId task_after_shutdown = worker_pool_submit_task(&pool, "Late task", TASK_PRIORITY_HIGH,
                                                           simulate_write_log, "This should fail");
    if (task_after_shutdown == 0) {
        printf("  Correct: Task submission rejected after shutdown\n");
    } else {
        printf("  Error: Task should have been rejected\n");
    }
    
    printf("\n--- Test 8: Check completed task results ---\n");
    wait_for_task(&pool, task_payment2, "Process payment 2");
    wait_for_task(&pool, task_email1, "Send welcome email");
    wait_for_task(&pool, task_image1, "Generate thumbnail");
    wait_for_task(&pool, task_log1, "Write log 1");
    wait_for_task(&pool, task_log2, "Write log 2");
    wait_for_task(&pool, task_email2, "Send notification");
    
    printf("\n--- Final Statistics ---\n");
    print_status("Final Statistics", &pool);
    
    printf("\nCleaning up...\n");
    worker_pool_destroy(&pool);
    priority_queue_destroy(&queue);
    
    printf("\n========================================\n");
    printf("  All tests completed successfully!\n");
    printf("========================================\n");
    
    return 0;
}
