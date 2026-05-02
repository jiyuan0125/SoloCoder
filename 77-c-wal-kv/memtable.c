#include "memtable.h"
#include "kv_store.h"
#include <stdlib.h>
#include <string.h>
#include <time.h>

static int random_level(void) {
    int level = 0;
    while (level < SKIPLIST_MAX_LEVEL - 1 && (rand() % 2) == 0) {
        level++;
    }
    return level;
}

static SkipListNode* skiplist_node_create(int level, const char* key, size_t key_len, 
                                            const char* value, size_t value_len, bool deleted) {
    SkipListNode* node = (SkipListNode*)malloc(sizeof(SkipListNode) + level * sizeof(SkipListNode*));
    if (!node) return NULL;
    
    node->key = (char*)malloc(key_len + 1);
    if (!node->key) {
        free(node);
        return NULL;
    }
    memcpy(node->key, key, key_len);
    node->key[key_len] = '\0';
    node->key_len = key_len;
    
    if (value && value_len > 0 && !deleted) {
        node->value = (char*)malloc(value_len + 1);
        if (!node->value) {
            free(node->key);
            free(node);
            return NULL;
        }
        memcpy(node->value, value, value_len);
        node->value[value_len] = '\0';
        node->value_len = value_len;
    } else {
        node->value = NULL;
        node->value_len = 0;
    }
    
    node->deleted = deleted;
    node->size = key_len + node->value_len + sizeof(SkipListNode);
    
    for (int i = 0; i <= level; i++) {
        node->forward[i] = NULL;
    }
    
    return node;
}

static void skiplist_node_destroy(SkipListNode* node) {
    if (!node) return;
    free(node->key);
    free(node->value);
    free(node);
}

SkipList* skiplist_create(void) {
    static int initialized = 0;
    if (!initialized) {
        srand((unsigned int)time(NULL));
        initialized = 1;
    }
    
    SkipList* sl = (SkipList*)malloc(sizeof(SkipList));
    if (!sl) return NULL;
    
    sl->level = 0;
    sl->size = 0;
    sl->count = 0;
    
    sl->header = (SkipListNode*)malloc(sizeof(SkipListNode) + SKIPLIST_MAX_LEVEL * sizeof(SkipListNode*));
    if (!sl->header) {
        free(sl);
        return NULL;
    }
    
    sl->header->key = NULL;
    sl->header->key_len = 0;
    sl->header->value = NULL;
    sl->header->value_len = 0;
    sl->header->deleted = false;
    sl->header->size = 0;
    
    for (int i = 0; i < SKIPLIST_MAX_LEVEL; i++) {
        sl->header->forward[i] = NULL;
    }
    
    return sl;
}

void skiplist_destroy(SkipList* sl) {
    if (!sl) return;
    
    SkipListNode* current = sl->header->forward[0];
    while (current) {
        SkipListNode* next = current->forward[0];
        skiplist_node_destroy(current);
        current = next;
    }
    
    free(sl->header);
    free(sl);
}

static int key_compare(const char* key1, size_t len1, const char* key2, size_t len2) {
    size_t min_len = len1 < len2 ? len1 : len2;
    int cmp = memcmp(key1, key2, min_len);
    if (cmp != 0) return cmp;
    if (len1 < len2) return -1;
    if (len1 > len2) return 1;
    return 0;
}

int skiplist_put(SkipList* sl, const char* key, size_t key_len, const char* value, size_t value_len) {
    if (!sl || !key || key_len == 0) return -1;
    
    SkipListNode* update[SKIPLIST_MAX_LEVEL];
    SkipListNode* current = sl->header;
    
    for (int i = sl->level; i >= 0; i--) {
        while (current->forward[i] != NULL && 
               key_compare(current->forward[i]->key, current->forward[i]->key_len, key, key_len) < 0) {
            current = current->forward[i];
        }
        update[i] = current;
    }
    
    current = current->forward[0];
    
    if (current != NULL && key_compare(current->key, current->key_len, key, key_len) == 0) {
        size_t old_size = current->size;
        
        free(current->value);
        if (value && value_len > 0) {
            current->value = (char*)malloc(value_len + 1);
            if (!current->value) return -1;
            memcpy(current->value, value, value_len);
            current->value[value_len] = '\0';
            current->value_len = value_len;
        } else {
            current->value = NULL;
            current->value_len = 0;
        }
        
        current->deleted = false;
        current->size = key_len + current->value_len + sizeof(SkipListNode);
        
        sl->size = sl->size - old_size + current->size;
        return 0;
    }
    
    int level = random_level();
    
    if (level > sl->level) {
        for (int i = sl->level + 1; i <= level; i++) {
            update[i] = sl->header;
        }
        sl->level = level;
    }
    
    SkipListNode* new_node = skiplist_node_create(level, key, key_len, value, value_len, false);
    if (!new_node) return -1;
    
    for (int i = 0; i <= level; i++) {
        new_node->forward[i] = update[i]->forward[i];
        update[i]->forward[i] = new_node;
    }
    
    sl->size += new_node->size;
    sl->count++;
    return 0;
}

int skiplist_delete(SkipList* sl, const char* key, size_t key_len) {
    if (!sl || !key || key_len == 0) return -1;
    
    SkipListNode* update[SKIPLIST_MAX_LEVEL];
    SkipListNode* current = sl->header;
    
    for (int i = sl->level; i >= 0; i--) {
        while (current->forward[i] != NULL && 
               key_compare(current->forward[i]->key, current->forward[i]->key_len, key, key_len) < 0) {
            current = current->forward[i];
        }
        update[i] = current;
    }
    
    current = current->forward[0];
    
    if (current != NULL && key_compare(current->key, current->key_len, key, key_len) == 0) {
        current->deleted = true;
        free(current->value);
        current->value = NULL;
        current->value_len = 0;
        current->size = key_len + sizeof(SkipListNode);
        return 0;
    }
    
    int level = random_level();
    
    if (level > sl->level) {
        for (int i = sl->level + 1; i <= level; i++) {
            update[i] = sl->header;
        }
        sl->level = level;
    }
    
    SkipListNode* new_node = skiplist_node_create(level, key, key_len, NULL, 0, true);
    if (!new_node) return -1;
    
    for (int i = 0; i <= level; i++) {
        new_node->forward[i] = update[i]->forward[i];
        update[i]->forward[i] = new_node;
    }
    
    sl->size += new_node->size;
    sl->count++;
    return 0;
}

KVEntry* skiplist_get(SkipList* sl, const char* key, size_t key_len) {
    if (!sl || !key || key_len == 0) return NULL;
    
    SkipListNode* current = sl->header;
    
    for (int i = sl->level; i >= 0; i--) {
        while (current->forward[i] != NULL) {
            int cmp = key_compare(current->forward[i]->key, current->forward[i]->key_len, key, key_len);
            if (cmp < 0) {
                current = current->forward[i];
            } else if (cmp == 0) {
                SkipListNode* node = current->forward[i];
                if (node->deleted) return NULL;
                return kv_entry_create(node->key, node->key_len, node->value, node->value_len, node->deleted);
            } else {
                break;
            }
        }
    }
    
    current = current->forward[0];
    if (current != NULL && key_compare(current->key, current->key_len, key, key_len) == 0) {
        if (current->deleted) return NULL;
        return kv_entry_create(current->key, current->key_len, current->value, current->value_len, current->deleted);
    }
    
    return NULL;
}

SkipListNode* skiplist_first(SkipList* sl) {
    if (!sl || !sl->header) return NULL;
    return sl->header->forward[0];
}

SkipListNode* skiplist_next(SkipListNode* node) {
    if (!node) return NULL;
    return node->forward[0];
}

size_t skiplist_size(SkipList* sl) {
    if (!sl) return 0;
    return sl->size;
}

size_t skiplist_count(SkipList* sl) {
    if (!sl) return 0;
    return sl->count;
}
