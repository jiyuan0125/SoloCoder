#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include "thread_pool.h"

void *test_task(void *arg)
{
    int id = *(int *)arg;
    printf("Task %d: running on thread %lu\n", id, (unsigned long)pthread_self());
    usleep(100000);
    int *result = (int *)malloc(sizeof(int));
    *result = id * 2;
    return result;
}

void *long_task(void *arg)
{
    sleep(10);
    return NULL;
}

int main()
{
    thread_pool_t pool;
    int ret;

    printf("=== Test 1: Basic create and submit ===\n");
    ret = thread_pool_create(&pool, 2, 5);
    if (ret != 0)
    {
        fprintf(stderr, "thread_pool_create failed: %d\n", ret);
        return 1;
    }
    printf("Pool created with min=2, max=5\n");

    task_future_t futures[5];
    int args[5];

    for (int i = 0; i < 5; i++)
    {
        args[i] = i + 1;
        future_init(&futures[i]);
        ret = thread_pool_submit(&pool, test_task, &args[i], &futures[i]);
        if (ret != 0)
        {
            fprintf(stderr, "submit failed: %d\n", ret);
        }
    }

    printf("Submitted 5 tasks, waiting for results...\n");

    for (int i = 0; i < 5; i++)
    {
        ret = task_future_get(&futures[i], -1);
        if (ret == 0)
        {
            int *result = (int *)futures[i].result;
            printf("Task %d result: %d\n", i + 1, *result);
            free(result);
        }
        else if (ret == ECANCELED)
        {
            printf("Task %d was canceled\n", i + 1);
        }
        else
        {
            printf("Task %d get failed: %d\n", i + 1, ret);
        }
        future_destroy(&futures[i]);
    }

    printf("\n=== Test 2: Shutdown with no pending tasks ===\n");
    ret = thread_pool_shutdown(&pool, -1);
    printf("shutdown returned: %d (expected 0)\n", ret);
    thread_pool_destroy(&pool);

    printf("\n=== Test 3: Shutdown after submit ESHUTDOWN ===\n");
    thread_pool_t pool2;
    thread_pool_create(&pool2, 1, 2);
    thread_pool_shutdown(&pool2, -1);
    ret = thread_pool_submit(&pool2, test_task, NULL, NULL);
    printf("submit after shutdown: %d (expected ESHUTDOWN=%d)\n", ret, ESHUTDOWN);
    thread_pool_destroy(&pool2);

    printf("\n=== Test 4: Shutdown timeout with unfinished tasks ===\n");
    thread_pool_t pool3;
    thread_pool_create(&pool3, 2, 4);

    task_future_t long_futures[6];
    for (int i = 0; i < 6; i++)
    {
        future_init(&long_futures[i]);
        thread_pool_submit(&pool3, long_task, NULL, &long_futures[i]);
    }
    printf("Submitted 6 long tasks (10s each)\n");

    ret = thread_pool_shutdown(&pool3, 1000);
    printf("shutdown(1000ms) returned: %d unfinished tasks\n", ret);

    for (int i = 0; i < 6; i++)
    {
        ret = task_future_get(&long_futures[i], 0);
        if (ret == ECANCELED)
        {
            printf("Future %d: ECANCELED (as expected for queued tasks)\n", i);
        }
        else if (ret == ETIMEDOUT)
        {
            printf("Future %d: still running (ETIMEDOUT)\n", i);
        }
        future_destroy(&long_futures[i]);
    }
    thread_pool_destroy(&pool3);

    printf("\n=== All tests completed ===\n");
    return 0;
}
