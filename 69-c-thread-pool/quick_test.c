#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/time.h>
#include "thread_pool.h"

void *long_task(void *arg)
{
    sleep(10);
    return NULL;
}

int main()
{
    struct timeval start, end;
    
    printf("Test: shutdown timeout with pthread_cancel\n");
    printf("==========================================\n");
    fflush(stdout);
    
    thread_pool_t pool;
    thread_pool_create(&pool, 2, 4);
    
    task_future_t futures[6];
    for (int i = 0; i < 6; i++)
    {
        future_init(&futures[i]);
        thread_pool_submit(&pool, long_task, NULL, &futures[i]);
    }
    printf("Submitted 6 tasks (each 10s sleep)\n");
    fflush(stdout);
    
    gettimeofday(&start, NULL);
    int ret = thread_pool_shutdown(&pool, 1000);
    gettimeofday(&end, NULL);
    
    double elapsed = (end.tv_sec - start.tv_sec) + (end.tv_usec - start.tv_usec)/1000000.0;
    printf("\nshutdown(1000ms) returned: %d\n", ret);
    printf("Time elapsed: %.3f seconds (should be ~1 second, NOT 10+)\n", elapsed);
    fflush(stdout);
    
    int canceled = 0;
    for (int i = 0; i < 6; i++)
    {
        int s = task_future_get(&futures[i], 0);
        if (s == ECANCELED)
            canceled++;
        future_destroy(&futures[i]);
    }
    printf("Futures marked ECANCELED: %d\n", canceled);
    fflush(stdout);
    
    thread_pool_destroy(&pool);
    
    printf("\n==========================================\n");
    if (elapsed < 2.0 && ret > 0)
    {
        printf("SUCCESS: Shutdown timeout working correctly!\n");
    }
    else
    {
        printf("FAILURE: Something went wrong\n");
    }
    fflush(stdout);
    
    return 0;
}
