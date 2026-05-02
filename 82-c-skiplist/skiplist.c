#include "skiplist.h"
#include <stdlib.h>
#include <string.h>
#include <time.h>

static int random_level(int current_max)
{
    int level = 1;
    while (level < MAX_LEVEL && (rand() & 1) == 0) {
        level++;
    }
    int max_allowed = current_max + 1;
    if (level > max_allowed) {
        level = max_allowed;
    }
    return level;
}

static skiplist_node_t *create_node(int level, int key, void *value)
{
    skiplist_node_t *node = (skiplist_node_t *)malloc(
        sizeof(skiplist_node_t) + level * sizeof(skiplist_node_t *)
    );
    if (node == NULL) {
        return NULL;
    }
    node->key = key;
    node->value = value;
    node->level = level;
    memset(node->forward, 0, level * sizeof(skiplist_node_t *));
    return node;
}

skiplist_t *skiplist_create(void)
{
    srand((unsigned int)time(NULL));

    skiplist_t *sl = (skiplist_t *)malloc(sizeof(skiplist_t));
    if (sl == NULL) {
        return NULL;
    }

    sl->head = create_node(MAX_LEVEL, 0, NULL);
    if (sl->head == NULL) {
        free(sl);
        return NULL;
    }

    pthread_rwlock_init(&sl->rwlock, NULL);
    atomic_init(&sl->count, 0);
    sl->max_level = 1;

    return sl;
}

void skiplist_destroy(skiplist_t *sl)
{
    if (sl == NULL) {
        return;
    }

    skiplist_node_t *current = sl->head->forward[0];
    while (current != NULL) {
        skiplist_node_t *next = current->forward[0];
        free(current);
        current = next;
    }

    free(sl->head);
    pthread_rwlock_destroy(&sl->rwlock);
    free(sl);
}

int skiplist_insert(skiplist_t *sl, int key, void *value)
{
    if (sl == NULL) {
        return -1;
    }

    pthread_rwlock_wrlock(&sl->rwlock);

    skiplist_node_t *update[MAX_LEVEL];
    skiplist_node_t *current = sl->head;

    for (int i = sl->max_level - 1; i >= 0; i--) {
        while (current->forward[i] != NULL && current->forward[i]->key < key) {
            current = current->forward[i];
        }
        update[i] = current;
    }

    current = current->forward[0];
    if (current != NULL && current->key == key) {
        current->value = value;
        pthread_rwlock_unlock(&sl->rwlock);
        return 0;
    }

    int level = random_level(sl->max_level);

    skiplist_node_t *new_node = create_node(level, key, value);
    if (new_node == NULL) {
        pthread_rwlock_unlock(&sl->rwlock);
        return -1;
    }

    for (int i = 0; i < level; i++) {
        if (i < sl->max_level) {
            new_node->forward[i] = update[i]->forward[i];
            update[i]->forward[i] = new_node;
        } else {
            sl->head->forward[i] = new_node;
        }
    }

    if (level > sl->max_level) {
        sl->max_level = level;
    }

    atomic_fetch_add(&sl->count, 1);

    pthread_rwlock_unlock(&sl->rwlock);
    return 0;
}

void *skiplist_search(skiplist_t *sl, int key)
{
    if (sl == NULL) {
        return NULL;
    }

    skiplist_node_t *current = sl->head;

    for (int i = sl->max_level - 1; i >= 0; i--) {
        while (current->forward[i] != NULL && current->forward[i]->key < key) {
            current = current->forward[i];
        }
    }

    current = current->forward[0];
    if (current != NULL && current->key == key) {
        return current->value;
    }

    return NULL;
}

int skiplist_delete(skiplist_t *sl, int key)
{
    if (sl == NULL) {
        return -1;
    }

    pthread_rwlock_wrlock(&sl->rwlock);

    skiplist_node_t *update[MAX_LEVEL];
    skiplist_node_t *current = sl->head;

    for (int i = sl->max_level - 1; i >= 0; i--) {
        while (current->forward[i] != NULL && current->forward[i]->key < key) {
            current = current->forward[i];
        }
        update[i] = current;
    }

    current = current->forward[0];
    if (current == NULL || current->key != key) {
        pthread_rwlock_unlock(&sl->rwlock);
        return -1;
    }

    for (int i = 0; i < current->level; i++) {
        if (update[i] != NULL) {
            update[i]->forward[i] = current->forward[i];
        }
    }

    while (sl->max_level > 1 && sl->head->forward[sl->max_level - 1] == NULL) {
        sl->max_level--;
    }

    free(current);
    atomic_fetch_sub(&sl->count, 1);

    pthread_rwlock_unlock(&sl->rwlock);
    return 0;
}

void skiplist_range_query(skiplist_t *sl, int low, int high, range_callback callback)
{
    if (sl == NULL || callback == NULL || low >= high) {
        return;
    }

    skiplist_node_t *current = sl->head;

    for (int i = sl->max_level - 1; i >= 0; i--) {
        while (current->forward[i] != NULL && current->forward[i]->key < low) {
            current = current->forward[i];
        }
    }

    current = current->forward[0];
    while (current != NULL && current->key < high) {
        callback(current->key, current->value);
        current = current->forward[0];
    }
}

unsigned long skiplist_count(skiplist_t *sl)
{
    if (sl == NULL) {
        return 0;
    }
    return atomic_load(&sl->count);
}

void skiplist_level_stats(skiplist_t *sl)
{
    if (sl == NULL) {
        return;
    }

    int counts[MAX_LEVEL] = {0};
    skiplist_node_t *current = sl->head->forward[0];

    while (current != NULL) {
        for (int i = 0; i < current->level; i++) {
            counts[i]++;
        }
        current = current->forward[0];
    }

    printf("Skip List Level Statistics:\n");
    for (int i = 0; i < sl->max_level; i++) {
        printf("  Level %d: %d nodes\n", i + 1, counts[i]);
    }
}
