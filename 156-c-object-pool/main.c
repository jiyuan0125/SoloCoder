#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include "pool.h"
#include "monster.h"

#define TEST_POOL_SIZE 10
#define TEST_LEAK_TIMEOUT 2
#define NUM_THREADS 5
#define ITERATIONS_PER_THREAD 20

typedef struct ThreadArg {
    MonsterPool *pool;
    int thread_id;
} ThreadArg;

void print_pool_status(MonsterPool *pool, const char *message) {
    unsigned int free_count = pool_get_free_count(pool);
    unsigned int used_count = pool_get_used_count(pool);
    unsigned int capacity = pool_get_capacity(pool);
    unsigned long long total_alloc, total_dealloc, leak_recoveries;
    
    pool_get_statistics(pool, &total_alloc, &total_dealloc, &leak_recoveries);
    
    printf("\n=== %s ===\n", message);
    printf("Pool Capacity: %u\n", capacity);
    printf("Free Objects:  %u\n", free_count);
    printf("Used Objects:  %u\n", used_count);
    printf("Total Allocations:   %llu\n", total_alloc);
    printf("Total Deallocations: %llu\n", total_dealloc);
    printf("Leak Recoveries:     %llu\n", leak_recoveries);
}

void* worker_thread(void *arg) {
    ThreadArg *thread_arg = (ThreadArg*)arg;
    MonsterPool *pool = thread_arg->pool;
    int thread_id = thread_arg->thread_id;
    int i;
    Monster *monsters[5];
    int monster_count = 0;
    
    for (i = 0; i < ITERATIONS_PER_THREAD; i++) {
        Monster *m = pool_acquire(pool);
        
        if (m != NULL) {
            m->monster_id = (thread_id * 1000) + i;
            m->template_id = 100 + (i % 10);
            m->pos_x = (float)(rand() % 1000) / 10.0f;
            m->pos_y = (float)(rand() % 500) / 10.0f;
            m->pos_z = 0.0f;
            m->max_hp = 100 + (rand() % 900);
            m->current_hp = m->max_hp;
            m->level = 1 + (rand() % 10);
            
            if (monster_count < 5) {
                monsters[monster_count++] = m;
            } else {
                usleep(10000 + (rand() % 50000));
                pool_release(pool, m);
            }
        } else {
            usleep(50000);
        }
        
        if (monster_count > 0 && (rand() % 3 == 0)) {
            int idx = rand() % monster_count;
            pool_release(pool, monsters[idx]);
            monsters[idx] = monsters[monster_count - 1];
            monster_count--;
        }
    }
    
    while (monster_count > 0) {
        pool_release(pool, monsters[--monster_count]);
    }
    
    return NULL;
}

void test_basic_operations(void) {
    MonsterPool pool;
    Monster *m1, *m2, *m3;
    int ret;
    
    printf("\n========== Test 1: Basic Operations ==========\n");
    
    ret = pool_create(&pool, 5, 60);
    if (ret != 0) {
        printf("Failed to create pool!\n");
        return;
    }
    
    print_pool_status(&pool, "Initial State");
    
    m1 = pool_acquire(&pool);
    printf("\nAcquired m1: %p\n", (void*)m1);
    if (m1 != NULL) {
        m1->monster_id = 1001;
        m1->template_id = 50;
        m1->pos_x = 100.5f;
        m1->pos_y = 200.3f;
        m1->max_hp = 500;
        m1->current_hp = 500;
        m1->level = 10;
        printf("  Monster ID: %u, Template: %u, HP: %d/%d, Level: %d\n",
               m1->monster_id, m1->template_id, m1->current_hp, m1->max_hp, m1->level);
        printf("  Position: (%.1f, %.1f, %.1f)\n", m1->pos_x, m1->pos_y, m1->pos_z);
    }
    print_pool_status(&pool, "After Acquire m1");
    
    m2 = pool_acquire(&pool);
    m3 = pool_acquire(&pool);
    printf("\nAcquired m2: %p, m3: %p\n", (void*)m2, (void*)m3);
    print_pool_status(&pool, "After Acquire m2, m3");
    
    printf("\nReleasing m1...\n");
    ret = pool_release(&pool, m1);
    printf("Release result: %d (0=success)\n", ret);
    print_pool_status(&pool, "After Release m1");
    
    printf("\nReleasing m2 and m3...\n");
    pool_release(&pool, m2);
    pool_release(&pool, m3);
    print_pool_status(&pool, "After Release All");
    
    printf("\nTest m1 was reset properly:\n");
    m1 = pool_acquire(&pool);
    if (m1 != NULL) {
        printf("  After re-acquire: Monster ID: %u, Template: %u, HP: %d/%d, Level: %d\n",
               m1->monster_id, m1->template_id, m1->current_hp, m1->max_hp, m1->level);
        printf("  Position: (%.1f, %.1f, %.1f)\n", m1->pos_x, m1->pos_y, m1->pos_z);
        pool_release(&pool, m1);
    }
    
    pool_destroy(&pool);
    printf("\nBasic Operations Test Completed.\n");
}

void test_pool_exhaustion(void) {
    MonsterPool pool;
    Monster *monsters[10];
    Monster *extra;
    int i;
    
    printf("\n========== Test 2: Pool Exhaustion ==========\n");
    
    pool_create(&pool, 5, 60);
    print_pool_status(&pool, "Initial");
    
    printf("\nAcquiring all 5 objects...\n");
    for (i = 0; i < 5; i++) {
        monsters[i] = pool_acquire(&pool);
        printf("  Acquired monsters[%d]: %p\n", i, (void*)monsters[i]);
    }
    print_pool_status(&pool, "After Acquiring All 5");
    
    printf("\nTrying to acquire 6th object (should fail)...\n");
    extra = pool_acquire(&pool);
    printf("  Result: %p (expected: NULL)\n", (void*)extra);
    print_pool_status(&pool, "After Failed Acquire");
    
    printf("\nReleasing one object...\n");
    pool_release(&pool, monsters[0]);
    print_pool_status(&pool, "After Release 1");
    
    printf("\nNow acquiring should work...\n");
    extra = pool_acquire(&pool);
    printf("  Result: %p (expected: non-NULL)\n", (void*)extra);
    print_pool_status(&pool, "After Successful Re-Acquire");
    
    for (i = 1; i < 5; i++) {
        pool_release(&pool, monsters[i]);
    }
    pool_release(&pool, extra);
    
    pool_destroy(&pool);
    printf("\nPool Exhaustion Test Completed.\n");
}

void test_leak_detection(void) {
    MonsterPool pool;
    Monster *m1, *m2;
    unsigned int recovered;
    
    printf("\n========== Test 3: Leak Detection ==========\n");
    printf("Leak timeout set to %d seconds\n", TEST_LEAK_TIMEOUT);
    
    pool_create(&pool, 10, TEST_LEAK_TIMEOUT);
    print_pool_status(&pool, "Initial");
    
    printf("\nAcquiring 2 objects but NOT releasing them (simulating leak)...\n");
    m1 = pool_acquire(&pool);
    m2 = pool_acquire(&pool);
    printf("  Acquired m1: %p, m2: %p\n", (void*)m1, (void*)m2);
    print_pool_status(&pool, "After Acquire");
    
    printf("\nWaiting for leak timeout...\n");
    print_pool_status(&pool, "Before Sleep");
    sleep(TEST_LEAK_TIMEOUT + 1);
    
    printf("\nChecking for leaks...\n");
    recovered = pool_recover_leaks(&pool);
    printf("  Recovered %u leaked objects\n", recovered);
    print_pool_status(&pool, "After Leak Recovery");
    
    printf("\nVerifying pool works after recovery...\n");
    m1 = pool_acquire(&pool);
    printf("  Re-acquired: %p\n", (void*)m1);
    if (m1 != NULL) {
        pool_release(&pool, m1);
    }
    print_pool_status(&pool, "Final");
    
    pool_destroy(&pool);
    printf("\nLeak Detection Test Completed.\n");
}

void test_multithreading(void) {
    MonsterPool pool;
    pthread_t threads[NUM_THREADS];
    ThreadArg args[NUM_THREADS];
    int i;
    
    printf("\n========== Test 4: Multi-threading ==========\n");
    printf("Pool size: %d, Threads: %d, Iterations per thread: %d\n",
           TEST_POOL_SIZE, NUM_THREADS, ITERATIONS_PER_THREAD);
    
    pool_create(&pool, TEST_POOL_SIZE, 60);
    print_pool_status(&pool, "Initial");
    
    printf("\nStarting %d worker threads...\n", NUM_THREADS);
    for (i = 0; i < NUM_THREADS; i++) {
        args[i].pool = &pool;
        args[i].thread_id = i;
        pthread_create(&threads[i], NULL, worker_thread, &args[i]);
    }
    
    for (i = 0; i < NUM_THREADS; i++) {
        pthread_join(threads[i], NULL);
    }
    
    printf("\nAll threads completed.\n");
    print_pool_status(&pool, "After All Threads");
    
    pool_destroy(&pool);
    printf("\nMulti-threading Test Completed.\n");
}

int main(void) {
    printf("============================================\n");
    printf("    Monster Object Pool Test Suite\n");
    printf("============================================\n");
    
    test_basic_operations();
    test_pool_exhaustion();
    test_leak_detection();
    test_multithreading();
    
    printf("\n============================================\n");
    printf("    All Tests Completed Successfully!\n");
    printf("============================================\n");
    
    return 0;
}
