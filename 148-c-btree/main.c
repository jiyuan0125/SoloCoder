#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include "btree.h"

static void print_key(const bt_key_t *key) {
    int i;
    printf("[");
    for (i = 0; i < key->key_len; i++) {
        printf("%c", key->key_data[i]);
    }
    printf("]");
}

static void print_value(const bt_value_t *value) {
    int i;
    printf("[len=%u, data='", value->value_len);
    for (i = 0; i < (int)value->value_len && i < 50; i++) {
        printf("%c", value->value_data[i]);
    }
    if (value->value_len > 50) {
        printf("...");
    }
    printf("']");
}

static bt_key_t make_key(const char *str) {
    bt_key_t key;
    size_t len = strlen(str);
    if (len > MAX_KEY_SIZE) {
        len = MAX_KEY_SIZE;
    }
    key.key_len = (uint16_t)len;
    memcpy(key.key_data, str, len);
    return key;
}

static bt_value_t make_value(const char *str) {
    bt_value_t value;
    size_t len = strlen(str);
    if (len > MAX_VALUE_SIZE) {
        len = MAX_VALUE_SIZE;
    }
    value.value_len = (uint32_t)len;
    memcpy(value.value_data, str, len);
    return value;
}

int main(int argc, char *argv[]) {
    btree_t btree;
    bt_key_t key;
    bt_value_t value;
    bt_value_t read_value;
    int ret;
    int i;
    const char *data_path = "btree.db";
    const char *wal_path = "btree.wal";

    printf("========================================\n");
    printf("B+ Tree KV Storage Engine Demo\n");
    printf("Page Size: %d bytes\n", PAGE_SIZE);
    printf("Max Key Size: %d bytes\n", MAX_KEY_SIZE);
    printf("Max Value Size: %d bytes\n", MAX_VALUE_SIZE);
    printf("Cache Capacity: %d pages\n", CACHE_CAPACITY);
    printf("========================================\n\n");

    printf("[1] Creating new database...\n");
    ret = btree_create(&btree, data_path, wal_path);
    if (ret < 0) {
        printf("  Failed to create database!\n");
        return 1;
    }
    printf("  Database created successfully.\n\n");

    printf("[2] Inserting key-value pairs...\n");

    key = make_key("key1");
    value = make_value("This is the value for key1. Hello World!");
    printf("  Inserting: ");
    print_key(&key);
    printf(" -> ");
    print_value(&value);
    printf("\n");
    ret = btree_put(&btree, &key, &value);
    assert(ret == 0);

    key = make_key("key2");
    value = make_value("Value for key2 with some longer text content to test.");
    printf("  Inserting: ");
    print_key(&key);
    printf(" -> ");
    print_value(&value);
    printf("\n");
    ret = btree_put(&btree, &key, &value);
    assert(ret == 0);

    key = make_key("key3");
    value = make_value("The third key's value.");
    printf("  Inserting: ");
    print_key(&key);
    printf(" -> ");
    print_value(&value);
    printf("\n");
    ret = btree_put(&btree, &key, &value);
    assert(ret == 0);

    key = make_key("key0");
    value = make_value("This is key0 (should come first).");
    printf("  Inserting: ");
    print_key(&key);
    printf(" -> ");
    print_value(&value);
    printf("\n");
    ret = btree_put(&btree, &key, &value);
    assert(ret == 0);

    printf("  All insertions completed.\n\n");

    printf("[3] Querying keys...\n");

    key = make_key("key1");
    printf("  Querying: ");
    print_key(&key);
    printf("\n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Found: ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    NOT FOUND!\n");
    }

    key = make_key("key0");
    printf("  Querying: ");
    print_key(&key);
    printf("\n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Found: ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    NOT FOUND!\n");
    }

    key = make_key("nonexistent");
    printf("  Querying (non-existent): ");
    print_key(&key);
    printf("\n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Found (unexpected!): ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    Correctly NOT FOUND.\n");
    }

    printf("  Queries completed.\n\n");

    printf("[4] Updating existing key (key2)...\n");
    key = make_key("key2");
    value = make_value("UPDATED: This is a new value for key2!");
    printf("  Before update: querying key2\n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Old value: ");
        print_value(&read_value);
        printf("\n");
    }

    printf("  Inserting new value...\n");
    ret = btree_put(&btree, &key, &value);
    assert(ret == 0);

    printf("  After update: querying key2\n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    New value: ");
        print_value(&read_value);
        printf("\n");
    }
    printf("  Update completed.\n\n");

    printf("[5] Partial update on key1...\n");
    key = make_key("key1");
    printf("  Before partial update: \n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Value: ");
        print_value(&read_value);
        printf("\n");
    }

    printf("  Updating bytes 8-12 with 'XXXX'\n");
    ret = btree_update_partial(&btree, &key, 8, (const uint8_t*)"XXXX", 4);
    assert(ret == 0);

    printf("  After partial update: \n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Value: ");
        print_value(&read_value);
        printf("\n");
    }
    printf("  Partial update completed.\n\n");

    printf("[6] Deleting key 'key3'...\n");
    key = make_key("key3");
    printf("  Before deletion: \n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Found, value: ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    Not found.\n");
    }

    printf("  Deleting...\n");
    ret = btree_delete(&btree, &key);
    printf("    Delete returned: %d\n", ret);

    printf("  After deletion: \n");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Still FOUND (error!): ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    Correctly NOT FOUND.\n");
    }
    printf("  Deletion completed.\n\n");

    printf("[7] Bulk insert test (to trigger page splits)...\n");
    char key_buf[64];
    char val_buf[128];
    int insert_count = 0;

    for (i = 0; i < 100; i++) {
        snprintf(key_buf, sizeof(key_buf), "bulk_key_%04d", i);
        snprintf(val_buf, sizeof(val_buf),
                 "This is the value for bulk_key_%04d. "
                 "Adding more content to ensure page fills up. "
                 "The quick brown fox jumps over the lazy dog.",
                 i);

        key = make_key(key_buf);
        value = make_value(val_buf);

        ret = btree_put(&btree, &key, &value);
        if (ret == 0) {
            insert_count++;
        }
    }
    printf("  Bulk inserted %d keys.\n", insert_count);

    printf("  Verifying bulk inserts...\n");
    int verify_count = 0;
    for (i = 0; i < 100; i++) {
        snprintf(key_buf, sizeof(key_buf), "bulk_key_%04d", i);
        key = make_key(key_buf);
        ret = btree_get(&btree, &key, &read_value);
        if (ret == 0) {
            verify_count++;
        }
    }
    printf("  Verified %d keys.\n", verify_count);
    printf("  Bulk insert test completed.\n\n");

    printf("[8] Flushing and syncing to disk...\n");
    ret = btree_sync(&btree);
    if (ret == 0) {
        printf("  Sync successful.\n");
    } else {
        printf("  Sync failed!\n");
    }
    printf("\n");

    printf("[9] Closing database...\n");
    ret = btree_close(&btree);
    if (ret == 0) {
        printf("  Database closed successfully.\n");
    } else {
        printf("  Close failed!\n");
    }
    printf("\n");

    printf("[10] Re-opening database to verify persistence...\n");
    ret = btree_open(&btree, data_path, wal_path);
    if (ret < 0) {
        printf("  Failed to re-open database!\n");
        return 1;
    }
    printf("  Database re-opened.\n");

    printf("  Verifying key1 exists after reopen:\n");
    key = make_key("key1");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Found: ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    NOT FOUND (data lost?)\n");
    }

    printf("  Verifying bulk key_0050 exists:\n");
    key = make_key("bulk_key_0050");
    ret = btree_get(&btree, &key, &read_value);
    if (ret == 0) {
        printf("    Found: ");
        print_value(&read_value);
        printf("\n");
    } else {
        printf("    NOT FOUND (data lost?)\n");
    }

    printf("  Closing again...\n");
    btree_close(&btree);
    printf("\n");

    printf("========================================\n");
    printf("Demo completed successfully!\n");
    printf("Created files:\n");
    printf("  - %s (data file)\n", data_path);
    printf("  - %s (WAL log file)\n", wal_path);
    printf("========================================\n");

    return 0;
}
