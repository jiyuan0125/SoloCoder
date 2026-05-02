#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <string.h>
#include "hashmap.h"

void test_basic_operations() {
    printf("=== Test 1: Basic Operations ===\n");
    
    hashmap_t* map = hm_create(10);
    assert(map != NULL);
    printf("Created hashmap with capacity: %zu\n", hm_capacity(map));
    
    assert(hm_put(map, "key1", "value1") == 0);
    assert(hm_put(map, "key2", "value2") == 0);
    assert(hm_put(map, "key3", "value3") == 0);
    printf("Put 3 key-value pairs\n");
    printf("Size: %zu, Capacity: %zu\n", hm_size(map), hm_capacity(map));
    
    const char* val;
    assert(hm_get(map, "key1", &val) == 0);
    assert(strcmp(val, "value1") == 0);
    printf("Get key1: %s\n", val);
    
    assert(hm_get(map, "key2", &val) == 0);
    assert(strcmp(val, "value2") == 0);
    printf("Get key2: %s\n", val);
    
    assert(hm_contains(map, "key1") == 1);
    assert(hm_contains(map, "nonexistent") == 0);
    printf("Contains key1: yes, contains nonexistent: no\n");
    
    hm_destroy(map);
    printf("Test 1 PASSED\n\n");
}

void test_update() {
    printf("=== Test 2: Update Existing Key ===\n");
    
    hashmap_t* map = hm_create(10);
    
    hm_put(map, "key1", "old_value");
    
    const char* val;
    hm_get(map, "key1", &val);
    printf("Before update: key1 = %s\n", val);
    
    hm_put(map, "key1", "new_value");
    hm_get(map, "key1", &val);
    printf("After update: key1 = %s\n", val);
    assert(strcmp(val, "new_value") == 0);
    
    assert(hm_size(map) == 1);
    printf("Size is still 1 (correctly updated, not added)\n");
    
    hm_destroy(map);
    printf("Test 2 PASSED\n\n");
}

void test_delete() {
    printf("=== Test 3: Delete with Tombstone ===\n");
    
    hashmap_t* map = hm_create(10);
    
    hm_put(map, "a", "1");
    hm_put(map, "b", "2");
    hm_put(map, "c", "3");
    
    printf("Before delete - Size: %zu\n", hm_size(map));
    assert(hm_size(map) == 3);
    assert(hm_contains(map, "b") == 1);
    
    assert(hm_delete(map, "b") == 0);
    printf("Deleted 'b'\n");
    
    assert(hm_size(map) == 2);
    assert(hm_contains(map, "b") == 0);
    assert(hm_contains(map, "a") == 1);
    assert(hm_contains(map, "c") == 1);
    printf("After delete - Size: %zu, 'a' exists: yes, 'c' exists: yes\n", hm_size(map));
    
    assert(hm_delete(map, "nonexistent") == -1);
    printf("Delete nonexistent returns -1 (correct)\n");
    
    hm_destroy(map);
    printf("Test 3 PASSED\n\n");
}

void test_expansion() {
    printf("=== Test 4: Dynamic Expansion ===\n");
    
    hashmap_t* map = hm_create(7);
    size_t initial_cap = hm_capacity(map);
    printf("Initial capacity: %zu\n", initial_cap);
    
    for (int i = 0; i < 10; i++) {
        char key[32];
        char value[32];
        snprintf(key, sizeof(key), "key_%d", i);
        snprintf(value, sizeof(value), "value_%d", i);
        hm_put(map, key, value);
    }
    
    size_t new_cap = hm_capacity(map);
    printf("After inserting 10 elements - Capacity: %zu, Size: %zu\n", new_cap, hm_size(map));
    assert(new_cap > initial_cap);
    assert(hm_size(map) == 10);
    
    for (int i = 0; i < 10; i++) {
        char key[32];
        char expected[32];
        snprintf(key, sizeof(key), "key_%d", i);
        snprintf(expected, sizeof(expected), "value_%d", i);
        
        const char* val;
        assert(hm_get(map, key, &val) == 0);
        assert(strcmp(val, expected) == 0);
    }
    printf("All 10 keys are still accessible after expansion\n");
    
    hm_destroy(map);
    printf("Test 4 PASSED\n\n");
}

void test_shrinking() {
    printf("=== Test 5: Dynamic Shrinking ===\n");
    
    hashmap_t* map = hm_create(100);
    size_t initial_cap = hm_capacity(map);
    printf("Initial capacity: %zu\n", initial_cap);
    
    for (int i = 0; i < 30; i++) {
        char key[32];
        snprintf(key, sizeof(key), "key_%d", i);
        hm_put(map, key, "value");
    }
    printf("After inserting 30 elements - Size: %zu\n", hm_size(map));
    
    for (int i = 0; i < 25; i++) {
        char key[32];
        snprintf(key, sizeof(key), "key_%d", i);
        hm_delete(map, key);
    }
    
    size_t new_cap = hm_capacity(map);
    printf("After deleting 25 elements - Capacity: %zu, Size: %zu\n", new_cap, hm_size(map));
    
    assert(hm_size(map) == 5);
    
    for (int i = 25; i < 30; i++) {
        char key[32];
        snprintf(key, sizeof(key), "key_%d", i);
        assert(hm_contains(map, key) == 1);
    }
    printf("Remaining 5 keys are still accessible\n");
    
    hm_destroy(map);
    printf("Test 5 PASSED\n\n");
}

void test_robin_hood() {
    printf("=== Test 6: Robin Hood Probing (Rich give to Poor) ===\n");
    
    hashmap_t* map = hm_create(7);
    
    hm_put(map, "apple", "red");
    hm_put(map, "banana", "yellow");
    hm_put(map, "cherry", "red");
    hm_put(map, "date", "brown");
    hm_put(map, "elder", "purple");
    
    printf("Inserted 5 elements - Size: %zu\n", hm_size(map));
    assert(hm_size(map) == 5);
    
    const char* val;
    assert(hm_get(map, "apple", &val) == 0);
    assert(strcmp(val, "red") == 0);
    
    assert(hm_get(map, "banana", &val) == 0);
    assert(strcmp(val, "yellow") == 0);
    
    assert(hm_get(map, "cherry", &val) == 0);
    assert(strcmp(val, "red") == 0);
    
    assert(hm_get(map, "date", &val) == 0);
    assert(strcmp(val, "brown") == 0);
    
    assert(hm_get(map, "elder", &val) == 0);
    assert(strcmp(val, "purple") == 0);
    
    printf("All keys retrieved correctly with Robin Hood probing\n");
    
    hm_destroy(map);
    printf("Test 6 PASSED\n\n");
}

void test_tombstone_probe() {
    printf("=== Test 7: Probe Past Tombstones ===\n");
    
    hashmap_t* map = hm_create(7);
    
    hm_put(map, "a", "1");
    hm_put(map, "b", "2");
    hm_put(map, "c", "3");
    hm_put(map, "d", "4");
    
    hm_delete(map, "b");
    hm_delete(map, "c");
    
    printf("Deleted 'b' and 'c' (created tombstones)\n");
    printf("Size: %zu\n", hm_size(map));
    
    assert(hm_contains(map, "a") == 1);
    assert(hm_contains(map, "d") == 1);
    printf("'a' and 'd' are still accessible (probed past tombstones)\n");
    
    hm_put(map, "new_key", "new_value");
    printf("Inserted 'new_key' (should reuse tombstone slot)\n");
    
    assert(hm_contains(map, "new_key") == 1);
    assert(hm_contains(map, "a") == 1);
    assert(hm_contains(map, "d") == 1);
    printf("All keys accessible after reusing tombstone\n");
    
    hm_destroy(map);
    printf("Test 7 PASSED\n\n");
}

void test_deep_copy() {
    printf("=== Test 8: Deep Copy of Keys and Values ===\n");
    
    hashmap_t* map = hm_create(10);
    
    char* key_buf = malloc(32);
    char* val_buf = malloc(32);
    strcpy(key_buf, "original_key");
    strcpy(val_buf, "original_value");
    
    hm_put(map, key_buf, val_buf);
    
    strcpy(key_buf, "modified_key");
    strcpy(val_buf, "modified_value");
    
    const char* val;
    assert(hm_get(map, "original_key", &val) == 0);
    assert(strcmp(val, "original_value") == 0);
    printf("Original key/value preserved (deep copy works)\n");
    
    free(key_buf);
    free(val_buf);
    hm_destroy(map);
    printf("Test 8 PASSED\n\n");
}

int main() {
    printf("========================================\n");
    printf("Hashmap Test Suite - Robin Hood Hashing\n");
    printf("========================================\n\n");
    
    test_basic_operations();
    test_update();
    test_delete();
    test_expansion();
    test_shrinking();
    test_robin_hood();
    test_tombstone_probe();
    test_deep_copy();
    
    printf("========================================\n");
    printf("All tests PASSED!\n");
    printf("========================================\n");
    
    return 0;
}
