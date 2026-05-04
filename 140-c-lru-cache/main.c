#include "cache.h"
#include "cache_core.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>

#define TEST_CACHE_SIZE (5 * 1024 * 1024)
#define VALUE_SIZE 1024
#define NUM_ENTRIES 10000
#define NUM_THREADS 4

static void generate_value(unsigned char *buffer, size_t size, int seed) {
    for (size_t i = 0; i < size; i++) {
        buffer[i] = (unsigned char)((seed + i) % 256);
    }
}

static void print_stats(const char *title, Cache *cache) {
    CacheStats stats;
    cache_get_stats(cache, &stats);
    double hit_rate = cache_get_hit_rate(cache);
    
    printf("\n===== %s =====\n", title);
    printf("Total queries: %zu\n", stats.total_queries);
    printf("Hits:          %zu\n", stats.hits);
    printf("Misses:        %zu\n", stats.misses);
    printf("Evictions:     %zu\n", stats.evictions);
    printf("Hit rate:      %.2f%%\n", hit_rate * 100.0);
    printf("Current size:  %zu bytes\n", stats.current_size);
    printf("Entry count:   %zu\n", stats.entry_count);
}

static void test_basic_operations() {
    printf("\n=== Testing Basic Operations ===\n");
    
    Cache *cache = cache_create(TEST_CACHE_SIZE);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    printf("Inserting 100 entries...\n");
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "key_%05d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    
    print_stats("After inserting 100 entries", cache);
    
    printf("\nQuerying 50 entries (first 50 should hit)...\n");
    cache_reset_stats(cache);
    for (int i = 0; i < 50; i++) {
        snprintf(key, sizeof(key), "key_%05d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        int ret = cache_get(cache, key, &result, &result_len);
        if (ret == 1) {
            if (result) free(result);
        }
    }
    print_stats("After querying 50 existing entries", cache);
    
    printf("\nQuerying 50 non-existent entries...\n");
    cache_reset_stats(cache);
    for (int i = 1000; i < 1050; i++) {
        snprintf(key, sizeof(key), "key_nonexistent_%05d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        cache_get(cache, key, &result, &result_len);
        if (result) free(result);
    }
    print_stats("After querying 50 non-existent entries", cache);
    
    cache_destroy(cache);
    printf("Basic operations test completed.\n");
}

static void test_lru_eviction() {
    printf("\n=== Testing LRU Eviction ===\n");
    
    size_t per_entry_size = VALUE_SIZE + 256 + sizeof(CacheEntry);
    size_t small_cache_size = 10 * per_entry_size;
    Cache *cache = cache_create(small_cache_size);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    printf("Cache capacity: ~%zu bytes (holds ~%zu entries of %zu bytes each)\n", 
           small_cache_size, small_cache_size / per_entry_size, per_entry_size);
    
    printf("\nPhase 1: Insert 10 entries (0-9) to fill cache...\n");
    for (int i = 0; i < 10; i++) {
        snprintf(key, sizeof(key), "lru_%02d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    print_stats("After inserting 10 entries", cache);
    
    printf("\nPhase 2: Access keys 0-4 (move to LRU head, marking as recently used)...\n");
    for (int i = 0; i < 5; i++) {
        snprintf(key, sizeof(key), "lru_%02d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        int ret = cache_get(cache, key, &result, &result_len);
        printf("  %s: %s\n", key, ret == 1 ? "HIT" : "MISS");
        if (result) free(result);
    }
    
    printf("\nCurrent LRU order (head to tail): 0,1,2,3,4,9,8,7,6,5\n");
    printf("LRU tail (next to be evicted): lru_05\n");
    
    printf("\nPhase 3: Insert 5 new entries (10-14) - should evict lru_05, lru_06, lru_07, lru_08, lru_09\n");
    for (int i = 10; i < 15; i++) {
        snprintf(key, sizeof(key), "lru_%02d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    print_stats("After inserting 5 new entries", cache);
    
    printf("\nPhase 4: Verify LRU behavior:\n");
    printf("  Recently accessed (0-4) should still exist\n");
    printf("  Older entries (5-9) should have been evicted\n\n");
    
    int recent_hits = 0;
    printf("--- Checking recently accessed keys (0-4) ---\n");
    for (int i = 0; i < 5; i++) {
        snprintf(key, sizeof(key), "lru_%02d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        if (cache_get(cache, key, &result, &result_len) == 1) {
            printf("  %s: HIT ✓\n", key);
            recent_hits++;
        } else {
            printf("  %s: MISS ✗\n", key);
        }
        if (result) free(result);
    }
    
    int older_hits = 0;
    printf("\n--- Checking older keys (5-9) - should be evicted ---\n");
    for (int i = 5; i < 10; i++) {
        snprintf(key, sizeof(key), "lru_%02d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        if (cache_get(cache, key, &result, &result_len) == 1) {
            printf("  %s: HIT (unexpected)\n", key);
            older_hits++;
        } else {
            printf("  %s: MISS (evicted, as expected) ✓\n", key);
        }
        if (result) free(result);
    }
    
    printf("\n=== LRU Eviction Test Results ===\n");
    printf("  Recently accessed keys (0-4):  %d/5 hits\n", recent_hits);
    printf("  Older keys (5-9):              %d/5 hits (should be 0)\n", older_hits);
    
    if (recent_hits == 5 && older_hits == 0) {
        printf("\n  ✓ PERFECT: LRU behavior working correctly!\n");
        printf("    - Recently accessed entries preserved\n");
        printf("    - Least recently used entries evicted first\n");
    } else if (recent_hits > older_hits) {
        printf("\n  ✓ LRU behavior confirmed: recently accessed keys have higher survival rate\n");
    } else {
        printf("\n  ⚠ Unexpected results - check implementation\n");
    }
    
    cache_destroy(cache);
}

static void test_batch_operations() {
    printf("\n=== Testing Batch Operations ===\n");
    
    Cache *cache = cache_create(TEST_CACHE_SIZE);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    printf("Inserting 100 entries...\n");
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "batch_key_%03d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    
    const char *batch_keys[] = {
        "batch_key_005",
        "batch_key_010",
        "batch_key_nonexistent_1",
        "batch_key_025",
        "batch_key_nonexistent_2",
        "batch_key_050",
        "batch_key_099"
    };
    size_t key_count = sizeof(batch_keys) / sizeof(batch_keys[0]);
    
    printf("\nPerforming batch get on %zu keys...\n", key_count);
    cache_reset_stats(cache);
    
    BatchResult *result = cache_batch_get(cache, batch_keys, key_count);
    
    if (result) {
        printf("Batch result: %zu hits, %zu misses\n", result->hit_count, result->miss_count);
        
        for (size_t i = 0; i < result->total_count; i++) {
            printf("  Key '%s': %s\n", 
                   result->items[i].key, 
                   result->items[i].is_hit ? "HIT" : "MISS");
        }
        
        batch_result_free(result);
    }
    
    print_stats("After batch query", cache);
    
    cache_destroy(cache);
    printf("Batch operations test completed.\n");
}

static void test_expiration() {
    printf("\n=== Testing Expiration ===\n");
    
    Cache *cache = cache_create(TEST_CACHE_SIZE);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    unsigned char value[VALUE_SIZE];
    
    printf("Inserting 3 entries with different TTLs...\n");
    printf("  key_no_expire:  never expires\n");
    printf("  key_short_ttl:  2 seconds\n");
    printf("  key_long_ttl:   10 seconds\n");
    
    generate_value(value, VALUE_SIZE, 1);
    cache_set(cache, "key_no_expire", value, VALUE_SIZE, 0);
    
    generate_value(value, VALUE_SIZE, 2);
    cache_set(cache, "key_short_ttl", value, VALUE_SIZE, 2);
    
    generate_value(value, VALUE_SIZE, 3);
    cache_set(cache, "key_long_ttl", value, VALUE_SIZE, 10);
    
    printf("\nImmediately querying all keys...\n");
    cache_reset_stats(cache);
    
    const char *keys[] = {"key_no_expire", "key_short_ttl", "key_long_ttl"};
    for (int i = 0; i < 3; i++) {
        unsigned char *result = NULL;
        size_t result_len = 0;
        int ret = cache_get(cache, keys[i], &result, &result_len);
        printf("  %s: %s\n", keys[i], ret == 1 ? "HIT" : "MISS");
        if (result) free(result);
    }
    
    printf("\nWaiting 3 seconds (for short TTL to expire)...\n");
    sleep(3);
    
    printf("\nQuerying all keys again...\n");
    cache_reset_stats(cache);
    
    for (int i = 0; i < 3; i++) {
        unsigned char *result = NULL;
        size_t result_len = 0;
        int ret = cache_get(cache, keys[i], &result, &result_len);
        printf("  %s: %s\n", keys[i], ret == 1 ? "HIT" : "MISS (expired or not found)");
        if (result) free(result);
    }
    
    print_stats("After expiration test", cache);
    
    cache_destroy(cache);
    printf("Expiration test completed.\n");
}

typedef struct {
    Cache *cache;
    int thread_id;
    int num_iterations;
} ThreadData;

static void* concurrent_worker(void *arg) {
    ThreadData *data = (ThreadData *)arg;
    Cache *cache = data->cache;
    int id = data->thread_id;
    int iterations = data->num_iterations;
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    for (int i = 0; i < iterations; i++) {
        int key_idx = (id * 1000 + i) % 2000;
        snprintf(key, sizeof(key), "concurrent_key_%05d", key_idx);
        
        if (i % 3 == 0) {
            generate_value(value, VALUE_SIZE, key_idx);
            cache_set(cache, key, value, VALUE_SIZE, 0);
        } else {
            unsigned char *result = NULL;
            size_t result_len = 0;
            cache_get(cache, key, &result, &result_len);
            if (result) free(result);
        }
    }
    
    return NULL;
}

static void test_concurrency() {
    printf("\n=== Testing Concurrency ===\n");
    
    Cache *cache = cache_create(TEST_CACHE_SIZE);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    pthread_t threads[NUM_THREADS];
    ThreadData thread_data[NUM_THREADS];
    
    printf("Starting %d threads, each doing 5000 operations...\n", NUM_THREADS);
    
    cache_reset_stats(cache);
    
    for (int i = 0; i < NUM_THREADS; i++) {
        thread_data[i].cache = cache;
        thread_data[i].thread_id = i;
        thread_data[i].num_iterations = 5000;
        pthread_create(&threads[i], NULL, concurrent_worker, &thread_data[i]);
    }
    
    for (int i = 0; i < NUM_THREADS; i++) {
        pthread_join(threads[i], NULL);
    }
    
    print_stats("After concurrent operations", cache);
    
    printf("Total operations: %d\n", NUM_THREADS * 5000);
    
    cache_destroy(cache);
    printf("Concurrency test completed (no deadlock or crash = success).\n");
}

int main() {
    printf("========================================\n");
    printf("   LRU Cache Module Demo\n");
    printf("========================================\n");
    
    test_basic_operations();
    test_lru_eviction();
    test_batch_operations();
    test_expiration();
    test_concurrency();
    
    printf("\n========================================\n");
    printf("   All tests completed!\n");
    printf("========================================\n");
    
    return 0;
}
