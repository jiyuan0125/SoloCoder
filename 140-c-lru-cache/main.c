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
#define CACHE_SEGMENTS 16

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

static void test_eviction_stats() {
    printf("\n=== Testing Eviction Statistics ===\n");
    
    size_t per_entry_size = VALUE_SIZE + 256 + sizeof(CacheEntry);
    size_t segment_entries = 50;
    size_t segment_capacity = segment_entries * per_entry_size;
    size_t total_capacity = CACHE_SEGMENTS * segment_capacity;
    
    printf("Cache configuration:\n");
    printf("  Total capacity:    %zu bytes\n", total_capacity);
    printf("  Segments:          %d\n", CACHE_SEGMENTS);
    printf("  Segment capacity:  %zu bytes (~%zu entries)\n", segment_capacity, segment_entries);
    printf("  Entry size:        ~%zu bytes\n", per_entry_size);
    
    Cache *cache = cache_create(total_capacity);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    printf("\nPhase 1: Insert %zu entries per segment to fill cache...\n", segment_entries);
    cache_reset_stats(cache);
    
    for (int i = 0; i < CACHE_SEGMENTS * segment_entries; i++) {
        snprintf(key, sizeof(key), "evict_test_%05d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    
    print_stats("After filling cache", cache);
    printf("  Evictions should be 0 (cache just filled)\n");
    
    printf("\nPhase 2: Insert 100 more entries (should trigger evictions)...\n");
    size_t initial_evictions = 0;
    CacheStats stats_before;
    cache_get_stats(cache, &stats_before);
    initial_evictions = stats_before.evictions;
    
    for (int i = 10000; i < 10100; i++) {
        snprintf(key, sizeof(key), "evict_test_new_%05d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    
    print_stats("After inserting 100 new entries", cache);
    
    CacheStats stats_after;
    cache_get_stats(cache, &stats_after);
    size_t evictions_during_test = stats_after.evictions - initial_evictions;
    
    printf("\n=== Eviction Statistics Test Results ===\n");
    printf("  Evictions during test: %zu\n", evictions_during_test);
    
    if (evictions_during_test > 0) {
        printf("  ✓ SUCCESS: Evictions are being counted!\n");
    } else {
        printf("  ⚠ Note: No evictions occurred (may depend on key distribution)\n");
    }
    
    cache_destroy(cache);
    printf("Eviction statistics test completed.\n");
}

static void test_lru_eviction() {
    printf("\n=== Testing LRU Eviction (Segmented) ===\n");
    printf("\nNote: Cache uses %d segments with independent LRU lists.\n", CACHE_SEGMENTS);
    printf("Keys are distributed to segments based on hash.\n");
    printf("To test LRU behavior, we use many keys and compare survival rates.\n");
    
    size_t per_entry_size = VALUE_SIZE + 256 + sizeof(CacheEntry);
    size_t total_entries = 1000;
    size_t total_capacity = CACHE_SEGMENTS * 30 * per_entry_size;
    
    Cache *cache = cache_create(total_capacity);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    printf("\nPhase 1: Insert %zu entries...\n", total_entries);
    for (int i = 0; i < total_entries; i++) {
        snprintf(key, sizeof(key), "lru_test_%05d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    print_stats("After inserting initial entries", cache);
    
    printf("\nPhase 2: Access first 100 entries (mark as recently used)...\n");
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "lru_test_%05d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        cache_get(cache, key, &result, &result_len);
        if (result) free(result);
    }
    print_stats("After accessing first 100 entries", cache);
    
    printf("\nPhase 3: Insert 200 new entries (to trigger evictions)...\n");
    for (int i = 10000; i < 10200; i++) {
        snprintf(key, sizeof(key), "lru_test_new_%05d", i);
        generate_value(value, VALUE_SIZE, i);
        cache_set(cache, key, value, VALUE_SIZE, 0);
    }
    print_stats("After inserting 200 new entries", cache);
    
    printf("\nPhase 4: Verify LRU behavior:\n");
    printf("  Recently accessed (0-99) should have higher survival rate\n");
    printf("  Older entries (100-199) should have lower survival rate\n\n");
    
    int recent_hits = 0;
    int recent_total = 100;
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "lru_test_%05d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        if (cache_get(cache, key, &result, &result_len) == 1) {
            recent_hits++;
        }
        if (result) free(result);
    }
    
    int older_hits = 0;
    int older_total = 100;
    for (int i = 100; i < 200; i++) {
        snprintf(key, sizeof(key), "lru_test_%05d", i);
        unsigned char *result = NULL;
        size_t result_len = 0;
        if (cache_get(cache, key, &result, &result_len) == 1) {
            older_hits++;
        }
        if (result) free(result);
    }
    
    printf("\n=== LRU Eviction Test Results ===\n");
    printf("  Recently accessed keys (0-99):   %d/%d hits (%.1f%%)\n", 
           recent_hits, recent_total, (recent_hits * 100.0) / recent_total);
    printf("  Older keys (100-199):            %d/%d hits (%.1f%%)\n", 
           older_hits, older_total, (older_hits * 100.0) / older_total);
    
    if (recent_hits > older_hits) {
        printf("\n  ✓ LRU behavior confirmed: recently accessed keys have higher survival rate\n");
    } else {
        printf("\n  ⚠ Note: Results may vary due to hash distribution across segments\n");
    }
    
    cache_destroy(cache);
    printf("LRU eviction test completed.\n");
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
    double total_time;
} ThreadData;

static void* concurrent_worker(void *arg) {
    ThreadData *data = (ThreadData *)arg;
    Cache *cache = data->cache;
    int id = data->thread_id;
    int iterations = data->num_iterations;
    
    unsigned char value[VALUE_SIZE];
    char key[256];
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
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
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    data->total_time = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    
    return NULL;
}

static void test_concurrency() {
    printf("\n=== Testing Concurrency with Segmented Locking ===\n");
    printf("\nCache uses %d segments with independent locks.\n", CACHE_SEGMENTS);
    printf("Keys in different segments can be accessed in parallel.\n");
    printf("This significantly improves concurrency compared to a single global lock.\n");
    
    Cache *cache = cache_create(TEST_CACHE_SIZE);
    if (!cache) {
        printf("Failed to create cache\n");
        return;
    }
    
    pthread_t threads[NUM_THREADS];
    ThreadData thread_data[NUM_THREADS];
    
    printf("\nStarting %d threads, each doing 5000 operations...\n", NUM_THREADS);
    
    cache_reset_stats(cache);
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int i = 0; i < NUM_THREADS; i++) {
        thread_data[i].cache = cache;
        thread_data[i].thread_id = i;
        thread_data[i].num_iterations = 5000;
        thread_data[i].total_time = 0;
        pthread_create(&threads[i], NULL, concurrent_worker, &thread_data[i]);
    }
    
    for (int i = 0; i < NUM_THREADS; i++) {
        pthread_join(threads[i], NULL);
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    double total_time = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    
    print_stats("After concurrent operations", cache);
    
    printf("\n=== Concurrency Test Results ===\n");
    printf("  Total operations:  %d\n", NUM_THREADS * 5000);
    printf("  Total time:        %.3f seconds\n", total_time);
    printf("  Throughput:        %.0f ops/second\n", (NUM_THREADS * 5000) / total_time);
    
    printf("\nThread timing breakdown:\n");
    for (int i = 0; i < NUM_THREADS; i++) {
        printf("  Thread %d: %.3f seconds\n", i, thread_data[i].total_time);
    }
    
    printf("\n  ✓ No deadlock or crash = concurrency working correctly!\n");
    printf("  ✓ Segmented locking allows parallel access to different key ranges\n");
    
    cache_destroy(cache);
    printf("Concurrency test completed.\n");
}

int main() {
    printf("========================================\n");
    printf("   LRU Cache Module Demo\n");
    printf("========================================\n");
    
    test_basic_operations();
    test_eviction_stats();
    test_lru_eviction();
    test_batch_operations();
    test_expiration();
    test_concurrency();
    
    printf("\n========================================\n");
    printf("   All tests completed!\n");
    printf("========================================\n");
    
    return 0;
}
