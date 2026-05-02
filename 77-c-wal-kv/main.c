#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "db.h"
#include "kv_store.h"

static void test_basic_crud() {
    printf("=== Test: Basic CRUD ===\n");
    
    DB* db = db_open("./test_db_crud");
    if (!db) {
        printf("FAIL: Cannot open database\n");
        return;
    }
    
    int ret = db_put(db, "key1", 4, "value1", 6);
    printf("Put key1: %s\n", ret == 0 ? "OK" : "FAIL");
    
    ret = db_put(db, "key2", 4, "value2", 6);
    printf("Put key2: %s\n", ret == 0 ? "OK" : "FAIL");
    
    KVEntry* entry = db_get(db, "key1", 4);
    if (entry && strcmp(entry->value, "value1") == 0) {
        printf("Get key1: OK (value=%s)\n", entry->value);
    } else {
        printf("Get key1: FAIL\n");
    }
    kv_entry_free(entry);
    
    entry = db_get(db, "key2", 4);
    if (entry && strcmp(entry->value, "value2") == 0) {
        printf("Get key2: OK (value=%s)\n", entry->value);
    } else {
        printf("Get key2: FAIL\n");
    }
    kv_entry_free(entry);
    
    ret = db_put(db, "key1", 4, "new_value1", 10);
    printf("Update key1: %s\n", ret == 0 ? "OK" : "FAIL");
    
    entry = db_get(db, "key1", 4);
    if (entry && strcmp(entry->value, "new_value1") == 0) {
        printf("Get updated key1: OK (value=%s)\n", entry->value);
    } else {
        printf("Get updated key1: FAIL\n");
    }
    kv_entry_free(entry);
    
    ret = db_delete(db, "key2", 4);
    printf("Delete key2: %s\n", ret == 0 ? "OK" : "FAIL");
    
    entry = db_get(db, "key2", 4);
    if (entry == NULL) {
        printf("Get deleted key2: OK (NULL as expected)\n");
    } else {
        printf("Get deleted key2: FAIL (should be NULL)\n");
        kv_entry_free(entry);
    }
    
    db_close(db);
    printf("=== Basic CRUD Test Done ===\n\n");
}

static void test_scan() {
    printf("=== Test: Scan ===\n");
    
    DB* db = db_open("./test_db_scan");
    if (!db) {
        printf("FAIL: Cannot open database\n");
        return;
    }
    
    db_put(db, "user:1", 6, "Alice", 5);
    db_put(db, "user:2", 6, "Bob", 3);
    db_put(db, "user:10", 7, "Charlie", 7);
    db_put(db, "product:1", 9, "Apple", 5);
    db_put(db, "product:2", 9, "Banana", 6);
    
    size_t count;
    KVEntry** results = db_scan(db, "user:", 5, &count);
    
    printf("Scan 'user:' prefix, found %zu entries:\n", count);
    for (size_t i = 0; i < count; i++) {
        printf("  [%zu] %s -> %s\n", i, results[i]->key, results[i]->value);
    }
    
    if (count == 3) {
        printf("Scan count check: OK\n");
    } else {
        printf("Scan count check: FAIL (expected 3, got %zu)\n", count);
    }
    
    db_scan_results_free(results, count);
    
    db_close(db);
    printf("=== Scan Test Done ===\n\n");
}

static void test_persistence() {
    printf("=== Test: Persistence ===\n");
    
    {
        DB* db = db_open("./test_db_persist");
        if (!db) {
            printf("FAIL: Cannot open database first time\n");
            return;
        }
        
        db_put(db, "persist_key", 11, "persist_value", 13);
        printf("Written: persist_key -> persist_value\n");
        
        db_close(db);
        printf("Database closed\n");
    }
    
    {
        DB* db = db_open("./test_db_persist");
        if (!db) {
            printf("FAIL: Cannot open database second time\n");
            return;
        }
        
        KVEntry* entry = db_get(db, "persist_key", 11);
        if (entry && strcmp(entry->value, "persist_value") == 0) {
            printf("Read after reopen: OK (persist_key -> %s)\n", entry->value);
        } else {
            printf("Read after reopen: FAIL\n");
        }
        kv_entry_free(entry);
        
        db_close(db);
    }
    
    printf("=== Persistence Test Done ===\n\n");
}

static void test_memtable_flush() {
    printf("=== Test: Memtable Flush (approximate) ===\n");
    
    DB* db = db_open("./test_db_flush");
    if (!db) {
        printf("FAIL: Cannot open database\n");
        return;
    }
    
    printf("Writing multiple keys to trigger potential flush...\n");
    char key[64];
    char value[256];
    
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "flush_key_%03d", i);
        snprintf(value, sizeof(value), "This is the value for flush_key_%03d with some padding data to increase size.", i);
        db_put(db, key, strlen(key), value, strlen(value));
    }
    
    printf("Verifying all keys can be read...\n");
    int ok_count = 0;
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "flush_key_%03d", i);
        KVEntry* entry = db_get(db, key, strlen(key));
        if (entry) {
            ok_count++;
            kv_entry_free(entry);
        }
    }
    
    if (ok_count == 100) {
        printf("All 100 keys verified: OK\n");
    } else {
        printf("Verification: FAIL (%d of 100 keys found)\n", ok_count);
    }
    
    db_close(db);
    printf("=== Memtable Flush Test Done ===\n\n");
}

int main() {
    printf("========================================\n");
    printf("  KV Store Engine Tests\n");
    printf("========================================\n\n");
    
    test_basic_crud();
    test_scan();
    test_persistence();
    test_memtable_flush();
    
    printf("========================================\n");
    printf("  All Tests Completed\n");
    printf("========================================\n");
    
    return 0;
}
