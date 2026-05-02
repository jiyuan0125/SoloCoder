#include "skiplist.h"
#include <stdio.h>
#include <pthread.h>
#include <stdlib.h>

static void print_callback(int key, void *value)
{
    printf("  key=%d, value=%s\n", key, (char *)value);
}

static void *writer_thread(void *arg)
{
    skiplist_t *sl = (skiplist_t *)arg;
    for (int i = 0; i < 100; i++) {
        char *value = (char *)malloc(16);
        snprintf(value, 16, "val-%d", i);
        skiplist_insert(sl, i, value);
    }
    return NULL;
}

static void *reader_thread(void *arg)
{
    skiplist_t *sl = (skiplist_t *)arg;
    for (int i = 0; i < 100; i++) {
        void *val = skiplist_search(sl, i);
        (void)val;
    }
    return NULL;
}

int main(void)
{
    printf("=== Test 1: Basic Insert/Search/Delete ===\n");
    skiplist_t *sl = skiplist_create();
    if (sl == NULL) {
        printf("Failed to create skiplist\n");
        return 1;
    }

    char *v1 = "value1";
    char *v2 = "value2";
    char *v3 = "value3";
    char *v4 = "value4";

    skiplist_insert(sl, 10, v1);
    skiplist_insert(sl, 20, v2);
    skiplist_insert(sl, 30, v3);
    skiplist_insert(sl, 5, v4);

    printf("Count after insert: %lu\n", skiplist_count(sl));

    printf("Search 10: %s\n", (char *)skiplist_search(sl, 10));
    printf("Search 20: %s\n", (char *)skiplist_search(sl, 20));
    printf("Search 99: %p\n", skiplist_search(sl, 99));

    printf("\n=== Test 2: Range Query [10, 30) ===\n");
    skiplist_range_query(sl, 10, 30, print_callback);

    printf("\n=== Test 3: Delete 20 ===\n");
    int ret = skiplist_delete(sl, 20);
    printf("Delete result: %d\n", ret);
    printf("Search 20 after delete: %p\n", skiplist_search(sl, 20));
    printf("Count after delete: %lu\n", skiplist_count(sl));

    printf("\n=== Test 4: Level Statistics ===\n");
    skiplist_level_stats(sl);

    skiplist_destroy(sl);

    printf("\n=== Test 5: Concurrent Test ===\n");
    skiplist_t *sl2 = skiplist_create();
    if (sl2 == NULL) {
        printf("Failed to create skiplist2\n");
        return 1;
    }

    pthread_t writers[5];
    pthread_t readers[10];

    for (int i = 0; i < 5; i++) {
        pthread_create(&writers[i], NULL, writer_thread, sl2);
    }
    for (int i = 0; i < 10; i++) {
        pthread_create(&readers[i], NULL, reader_thread, sl2);
    }

    for (int i = 0; i < 5; i++) {
        pthread_join(writers[i], NULL);
    }
    for (int i = 0; i < 10; i++) {
        pthread_join(readers[i], NULL);
    }

    printf("Final count: %lu\n", skiplist_count(sl2));

    skiplist_node_t *cur = sl2->head->forward[0];
    while (cur != NULL) {
        skiplist_node_t *next = cur->forward[0];
        free(cur->value);
        cur = next;
    }

    skiplist_destroy(sl2);

    printf("\n=== All tests passed ===\n");
    return 0;
}
